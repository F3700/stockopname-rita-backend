package auth

import (
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"testing"
)

func TestPasswordHashRoundTrip(t *testing.T) {
	hash, err := HashPassword("rahasia-gudang")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !VerifyPassword(hash, "rahasia-gudang") {
		t.Error("expected the correct password to verify")
	}
	if VerifyPassword(hash, "rahasia-gudang ") {
		t.Error("expected a wrong password to be rejected")
	}
}

func TestPasswordHashFormat(t *testing.T) {
	hash, err := HashPassword("rahasia-gudang")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parts := strings.Split(hash, ":")
	if len(parts) != 4 {
		t.Fatalf("expected four :-separated parts, got %d in %q", len(parts), hash)
	}
	if parts[0] != PasswordScheme {
		t.Errorf("expected scheme %q, got %q", PasswordScheme, parts[0])
	}
	if iterations, _ := strconv.Atoi(parts[1]); iterations != PasswordIterations {
		t.Errorf("expected %d iterations, got %q", PasswordIterations, parts[1])
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil || len(salt) != saltLength {
		t.Errorf("expected a %d byte salt, got %d (%v)", saltLength, len(salt), err)
	}

	key, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil || len(key) != keyLength {
		t.Errorf("expected a %d byte key, got %d (%v)", keyLength, len(key), err)
	}
}

func TestPasswordHashUsesFreshSalt(t *testing.T) {
	first, err := HashPassword("rahasia-gudang")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second, err := HashPassword("rahasia-gudang")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if first == second {
		t.Error("two hashes of the same password must differ")
	}
}

func TestHashPasswordRejectsEmpty(t *testing.T) {
	if _, err := HashPassword(""); !errors.Is(err, ErrEmptyPassword) {
		t.Errorf("expected ErrEmptyPassword, got %v", err)
	}
}

func TestVerifyPasswordRejectsMalformedHashes(t *testing.T) {
	tests := map[string]string{
		"empty":            "",
		"plain text":       "rahasia-gudang",
		"wrong scheme":     "md5:210000:c2FsdA:c2FsdA",
		"missing parts":    "pbkdf2-sha256:210000:c2FsdA",
		"bad iterations":   "pbkdf2-sha256:abc:c2FsdA:c2FsdA",
		"zero iterations":  "pbkdf2-sha256:0:c2FsdA:c2FsdA",
		"silly iterations": "pbkdf2-sha256:999999999:c2FsdA:c2FsdA",
		"bad salt":         "pbkdf2-sha256:210000:!!!:c2FsdA",
		"bad key":          "pbkdf2-sha256:210000:c2FsdA:!!!",
	}

	for name, encoded := range tests {
		t.Run(name, func(t *testing.T) {
			if VerifyPassword(encoded, "rahasia-gudang") {
				t.Error("a malformed hash must never verify")
			}
		})
	}
}

func TestVerifyPasswordRejectsEmptyPassword(t *testing.T) {
	hash, err := HashPassword("rahasia-gudang")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if VerifyPassword(hash, "") {
		t.Error("an empty password must never verify")
	}
}

func TestValidatePasswordHash(t *testing.T) {
	hash, err := HashPassword("rahasia-gudang")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := ValidatePasswordHash(hash); err != nil {
		t.Errorf("expected a generated hash to be valid, got %v", err)
	}
	if err := ValidatePasswordHash("rahasia-gudang"); !errors.Is(err, ErrInvalidHash) {
		t.Errorf("expected ErrInvalidHash, got %v", err)
	}
}
