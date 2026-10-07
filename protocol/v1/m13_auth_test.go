package v1

import (
 "crypto/ed25519"
 "crypto/rand"
 "strings"
 "testing"
)

func TestM13AuthChallengeRoundTrip(t *testing.T) {
 pub, priv, err := ed25519.GenerateKey(rand.Reader)
 if err != nil { t.Fatal(err) }
 challenge, err := NewM13Challenge()
 if err != nil { t.Fatal(err) }
 signature, err := SignM13Challenge(priv, "node-001", challenge)
 if err != nil { t.Fatal(err) }
 if err := VerifyM13Challenge(pub, "node-001", challenge, signature); err != nil { t.Fatal(err) }
 if err := VerifyM13Challenge(pub, "node-002", challenge, signature); err == nil { t.Fatal("signature must be bound to node ID") }
 other, err := NewM13Challenge()
 if err != nil { t.Fatal(err) }
 if challenge == other { t.Fatal("two random challenges collided") }
 if err := VerifyM13Challenge(pub, "node-001", other, signature); err == nil { t.Fatal("signature must be bound to challenge") }
}

func TestM13AuthRejectsMalformedInputs(t *testing.T) {
 pub, priv, err := ed25519.GenerateKey(rand.Reader)
 if err != nil { t.Fatal(err) }
 challenge, err := NewM13Challenge()
 if err != nil { t.Fatal(err) }
 signature, err := SignM13Challenge(priv, "node-001", challenge)
 if err != nil { t.Fatal(err) }
 for _, bad := range []string{"", "not-base64", "AA", challenge+"="} {
  if _, err := SignM13Challenge(priv, "node-001", bad); err == nil { t.Fatalf("accepted bad challenge %q", bad) }
 }
 if _, err := SignM13Challenge(priv, "", challenge); err == nil { t.Fatal("accepted empty node ID") }
 if _, err := SignM13Challenge(priv, strings.Repeat("x", 129), challenge); err == nil { t.Fatal("accepted oversized node ID") }
 if _, err := SignM13Challenge(nil, "node-001", challenge); err == nil { t.Fatal("accepted missing private key") }
 if err := VerifyM13Challenge(nil, "node-001", challenge, signature); err == nil { t.Fatal("accepted missing public key") }
 if err := VerifyM13Challenge(pub, "node-001", challenge, "bad"); err == nil { t.Fatal("accepted bad signature") }
 if err := VerifyM13Challenge(pub, "node-001", challenge, signature+"="); err == nil { t.Fatal("accepted noncanonical signature") }
}
