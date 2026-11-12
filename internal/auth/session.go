package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// CookieName is the single cookie carrying the session. One name, one place.
const CookieName = "rita_session"

// MinSecretLength is the shortest key worth signing a cookie with: a shorter
// secret can be brute forced offline from a single captured cookie.
const MinSecretLength = 32

// ErrInvalidSession is the only failure Parse reports. Telling "expired" apart
// from "tampered" would only tell an attacker which half to work on.
var ErrInvalidSession = errors.New("auth: invalid session")

// Claims is what the cookie carries. Operator stays empty between login and the
// moment the user types a name.
type Claims struct {
	Operator  string `json:"op"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

// Manager signs and verifies session tokens. It keeps no state: the whole
// session lives in the cookie, so a deploy does not log anyone out.
type Manager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

// NewManager refuses a secret too short to defend a cookie.
func NewManager(secret []byte, ttl time.Duration) (*Manager, error) {
	if len(secret) < MinSecretLength {
		return nil, fmt.Errorf("auth: session secret must be at least %d bytes, got %d", MinSecretLength, len(secret))
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("auth: session ttl must be positive, got %s", ttl)
	}
	return &Manager{secret: secret, ttl: ttl, now: time.Now}, nil
}

// TTL is how long a freshly issued token lives.
func (m *Manager) TTL() time.Duration {
	return m.ttl
}

// Issue signs a token for the operator. The claims hold one string and two
// integers, so encoding them cannot fail.
func (m *Manager) Issue(operator string) string {
	now := m.now()
	claims := Claims{
		Operator:  operator,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(m.ttl).Unix(),
	}

	payload, _ := json.Marshal(claims)

	return base64.RawURLEncoding.EncodeToString(payload) + "." +
		base64.RawURLEncoding.EncodeToString(m.sign(payload))
}

// Parse checks shape, signature and expiry, in that order. Every failure is
// reported as ErrInvalidSession.
func (m *Manager) Parse(token string) (Claims, error) {
	var claims Claims

	payloadPart, signaturePart, ok := strings.Cut(token, ".")
	if !ok || payloadPart == "" || signaturePart == "" || strings.Contains(signaturePart, ".") {
		return claims, ErrInvalidSession
	}

	payload, err := base64.RawURLEncoding.DecodeString(payloadPart)
	if err != nil {
		return claims, ErrInvalidSession
	}

	signature, err := base64.RawURLEncoding.DecodeString(signaturePart)
	if err != nil {
		return claims, ErrInvalidSession
	}

	// hmac.Equal compares in constant time; a plain bytes.Equal would leak how
	// much of a forged signature was already right.
	if !hmac.Equal(signature, m.sign(payload)) {
		return claims, ErrInvalidSession
	}

	if err := json.Unmarshal(payload, &claims); err != nil {
		return claims, ErrInvalidSession
	}

	if claims.ExpiresAt <= m.now().Unix() {
		return claims, ErrInvalidSession
	}

	return claims, nil
}

func (m *Manager) sign(payload []byte) []byte {
	mac := hmac.New(sha256.New, m.secret)
	mac.Write(payload)
	return mac.Sum(nil)
}
