package server

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"testing"
	"time"
)

func TestEnrollmentCodeIsHighEntropyAndHashed(t *testing.T) {
	first, firstHash, err := newEnrollmentCode()
	if err != nil {
		t.Fatal(err)
	}
	second, secondHash, err := newEnrollmentCode()
	if err != nil {
		t.Fatal(err)
	}
	if first == second || firstHash == secondHash {
		t.Fatal("enrollment codes must be unique")
	}
	if len(firstHash) != 64 || firstHash != hashEnrollmentCode(first) {
		t.Fatal("enrollment code hash is invalid")
	}
	if first == firstHash {
		t.Fatal("raw enrollment code must not equal stored hash")
	}
}

func TestSecureEnrollmentInputValidation(t *testing.T) {
	publicKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if !validEd25519PublicKey(base64.RawStdEncoding.EncodeToString(publicKey)) {
		t.Fatal("valid Ed25519 public key rejected")
	}
	if validEd25519PublicKey("not-a-key") {
		t.Fatal("invalid Ed25519 public key accepted")
	}
	if !validUUIDv4("123e4567-e89b-42d3-a456-426614174000") {
		t.Fatal("valid UUIDv4 rejected")
	}
	if validUUIDv4("123e4567-e89b-12d3-a456-426614174000") {
		t.Fatal("non-v4 UUID accepted")
	}
}

func TestEnrollmentRateLimiter(t *testing.T) {
	secureEnrollmentRate.Lock()
	secureEnrollmentRate.window = time.Time{}
	secureEnrollmentRate.attempts = 0
	secureEnrollmentRate.Unlock()
	now := time.Now().UTC()
	for i := 0; i < secureEnrollmentAttemptsPerMinute; i++ {
		if !allowSecureEnrollment(now) {
			t.Fatalf("request %d unexpectedly limited", i+1)
		}
	}
	if allowSecureEnrollment(now) {
		t.Fatal("rate limiter did not enforce the per-minute cap")
	}
	if !allowSecureEnrollment(now.Add(time.Minute)) {
		t.Fatal("rate limiter did not reset after one minute")
	}
}
