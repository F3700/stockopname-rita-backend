package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

const testSecret = "0123456789abcdef0123456789abcdef"

func newTestManager(t *testing.T) *Manager {
	t.Helper()

	manager, err := NewManager([]byte(testSecret), time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return manager
}

func TestSessionRoundTrip(t *testing.T) {
	manager := newTestManager(t)

	claims, err := manager.Parse(manager.Issue("Niko"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.Operator != "Niko" {
		t.Errorf("expected operator %q, got %q", "Niko", claims.Operator)
	}
	if got := claims.ExpiresAt - claims.IssuedAt; got != int64(time.Hour.Seconds()) {
		t.Errorf("expected a one hour session, got %d seconds", got)
	}
}

// An empty operator is a real state: the user logged in but has not typed a
// name yet, and the cookie must carry that faithfully.
func TestSessionRoundTripWithoutOperator(t *testing.T) {
	manager := newTestManager(t)

	claims, err := manager.Parse(manager.Issue(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.Operator != "" {
		t.Errorf("expected an empty operator, got %q", claims.Operator)
	}
}

func TestSessionRejectsTamperedPayload(t *testing.T) {
	manager := newTestManager(t)

	payload, signature, _ := strings.Cut(manager.Issue("Niko"), ".")

	decoded, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		t.Fatalf("decoding the payload: %v", err)
	}

	var claims Claims
	if err := json.Unmarshal(decoded, &claims); err != nil {
		t.Fatalf("decoding the claims: %v", err)
	}
	claims.Operator = "Budi"
	forged, _ := json.Marshal(claims)

	token := base64.RawURLEncoding.EncodeToString(forged) + "." + signature
	if _, err := manager.Parse(token); !errors.Is(err, ErrInvalidSession) {
		t.Errorf("expected ErrInvalidSession for a rewritten payload, got %v", err)
	}
}

func TestSessionRejectsOtherSecret(t *testing.T) {
	issuer := newTestManager(t)
	other, err := NewManager([]byte("fedcba9876543210fedcba9876543210"), time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := other.Parse(issuer.Issue("Niko")); !errors.Is(err, ErrInvalidSession) {
		t.Errorf("expected ErrInvalidSession, got %v", err)
	}
}

func TestSessionExpires(t *testing.T) {
	manager := newTestManager(t)
	issued := time.Date(2026, 11, 12, 8, 0, 0, 0, time.UTC)
	manager.now = func() time.Time { return issued }

	token := manager.Issue("Niko")

	manager.now = func() time.Time { return issued.Add(59 * time.Minute) }
	if _, err := manager.Parse(token); err != nil {
		t.Fatalf("a fresh session must be valid, got %v", err)
	}

	manager.now = func() time.Time { return issued.Add(2 * time.Hour) }
	if _, err := manager.Parse(token); !errors.Is(err, ErrInvalidSession) {
		t.Errorf("expected ErrInvalidSession after the ttl, got %v", err)
	}
}

func TestSessionRejectsMalformedTokens(t *testing.T) {
	manager := newTestManager(t)

	tests := map[string]string{
		"empty":            "",
		"no separator":     "abc",
		"empty payload":    ".abc",
		"empty signature":  "abc.",
		"three parts":      "abc.def.ghi",
		"not base64":       "!!!.???",
		"valid base64 txt": base64.RawURLEncoding.EncodeToString([]byte("hello")) + "." + base64.RawURLEncoding.EncodeToString([]byte("world")),
	}

	for name, token := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := manager.Parse(token); !errors.Is(err, ErrInvalidSession) {
				t.Errorf("expected ErrInvalidSession, got %v", err)
			}
		})
	}
}

func TestNewManagerRejectsWeakSecret(t *testing.T) {
	if _, err := NewManager([]byte("short"), time.Hour); err == nil {
		t.Error("expected a short secret to be rejected")
	}

	if _, err := NewManager([]byte(testSecret), 0); err == nil {
		t.Error("expected a non-positive ttl to be rejected")
	}
}
