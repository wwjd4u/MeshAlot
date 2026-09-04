package server

import (
	"encoding/hex"
	"testing"
)

func TestSessionTokenHash(t *testing.T) {
	tokenA, hashA, err := newSessionToken()
	if err != nil {
		t.Fatal(err)
	}
	tokenB, hashB, err := newSessionToken()
	if err != nil {
		t.Fatal(err)
	}
	if tokenA == tokenB || hashA == hashB {
		t.Fatal("session token generator repeated output")
	}
	if len(hashA) != 64 || hashSessionToken(tokenA) != hashA {
		t.Fatal("session token hash mismatch")
	}
	if _, err = hex.DecodeString(hashA); err != nil {
		t.Fatal("session hash is not hexadecimal")
	}
}
