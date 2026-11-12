package auth

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	// PasswordScheme names the only hash format this package writes.
	PasswordScheme = "pbkdf2-sha256"

	// PasswordIterations follows the OWASP guidance for PBKDF2-SHA256. Raising
	// it invalidates every stored hash, so it is a deliberate change.
	PasswordIterations = 210_000

	saltLength = 16
	keyLength  = 32

	// maxIterations bounds what VerifyPassword will honour. A hash pasted with a
	// silly count would otherwise block the login request for minutes.
	maxIterations = 1_000_000
)

var (
	// ErrEmptyPassword rejects a blank password at hashing time: an empty secret
	// is a misconfiguration, not a choice.
	ErrEmptyPassword = errors.New("auth: password must not be empty")

	// ErrInvalidHash reports a stored hash that is not in the expected format.
	ErrInvalidHash = errors.New("auth: password hash is not a valid pbkdf2-sha256 string")
)

// HashPassword derives a fresh hash with a random salt. The result is
// "pbkdf2-sha256:<iterations>:<salt>:<key>", both blobs in unpadded base64.
//
// The separator is ":" and not "$" on purpose: the value is stored in .env,
// where godotenv expands anything that looks like $NAME, which would eat the
// iteration count and the two blobs.
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}

	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("auth: reading salt: %w", err)
	}

	key, err := pbkdf2.Key(sha256.New, password, salt, PasswordIterations, keyLength)
	if err != nil {
		return "", fmt.Errorf("auth: deriving key: %w", err)
	}

	return fmt.Sprintf("%s:%d:%s:%s",
		PasswordScheme,
		PasswordIterations,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

// VerifyPassword reports whether password matches encoded. A malformed hash
// fails the check instead of panicking: it is configuration, not user input.
func VerifyPassword(encoded, password string) bool {
	if password == "" {
		return false
	}

	iterations, salt, wantKey, err := parsePasswordHash(encoded)
	if err != nil {
		return false
	}

	gotKey, err := pbkdf2.Key(sha256.New, password, salt, iterations, len(wantKey))
	if err != nil {
		return false
	}

	return subtle.ConstantTimeCompare(gotKey, wantKey) == 1
}

// ValidatePasswordHash lets the server refuse to start on a hash it could never
// verify, instead of failing every login afterwards.
func ValidatePasswordHash(encoded string) error {
	_, _, _, err := parsePasswordHash(encoded)
	return err
}

func parsePasswordHash(encoded string) (int, []byte, []byte, error) {
	parts := strings.Split(encoded, ":")
	if len(parts) != 4 || parts[0] != PasswordScheme {
		return 0, nil, nil, ErrInvalidHash
	}

	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations <= 0 || iterations > maxIterations {
		return 0, nil, nil, ErrInvalidHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil || len(salt) == 0 {
		return 0, nil, nil, ErrInvalidHash
	}

	key, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil || len(key) == 0 {
		return 0, nil, nil, ErrInvalidHash
	}

	return iterations, salt, key, nil
}
