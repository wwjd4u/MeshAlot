package v1

import (
 "crypto/ed25519"
 "crypto/rand"
 "encoding/base64"
 "errors"
)

// M13AuthDomain binds authentication signatures to the MeshAlot WebSocket
// challenge protocol, rather than permitting reuse of signed inventory data.
const M13AuthDomain = "meshalot/m13/websocket-auth/v1\n"
const M13NonceBytes = 32

// NewM13Challenge creates a cryptographically random, single-use challenge.
// The server must associate it with one connection, enforce a short expiry,
// and atomically consume it on the first verification attempt.
func NewM13Challenge() (string, error) {
 nonce := make([]byte, M13NonceBytes)
 if _, err := rand.Read(nonce); err != nil { return "", err }
 return base64.RawURLEncoding.EncodeToString(nonce), nil
}

// M13AuthMessage is the exact domain-separated byte string to sign.
// The challenge is issued by the server and bound to one connection there.
func M13AuthMessage(nodeID, challenge string) ([]byte, error) {
 if nodeID == "" || len(nodeID) > 128 { return nil, errors.New("invalid node id") }
 nonce, err := base64.RawURLEncoding.DecodeString(challenge)
 if err != nil || len(nonce) != M13NonceBytes || base64.RawURLEncoding.EncodeToString(nonce) != challenge {
  return nil, errors.New("invalid challenge")
 }
 return []byte(M13AuthDomain + nodeID + "\n" + challenge), nil
}

// SignM13Challenge proves possession of the enrolled node's private key.
func SignM13Challenge(privateKey ed25519.PrivateKey, nodeID, challenge string) (string, error) {
 if len(privateKey) != ed25519.PrivateKeySize { return "", errors.New("invalid private key") }
 message, err := M13AuthMessage(nodeID, challenge)
 if err != nil { return "", err }
 return base64.RawURLEncoding.EncodeToString(ed25519.Sign(privateKey, message)), nil
}

// VerifyM13Challenge validates a signature with the enrolled public key.
// Replay protection must be enforced by the server's challenge lifecycle.
func VerifyM13Challenge(publicKey ed25519.PublicKey, nodeID, challenge, signature string) error {
 if len(publicKey) != ed25519.PublicKeySize { return errors.New("invalid public key") }
 message, err := M13AuthMessage(nodeID, challenge)
 if err != nil { return err }
 raw, err := base64.RawURLEncoding.DecodeString(signature)
 if err != nil || len(raw) != ed25519.SignatureSize || base64.RawURLEncoding.EncodeToString(raw) != signature {
  return errors.New("invalid signature encoding")
 }
 if !ed25519.Verify(publicKey, message, raw) { return errors.New("signature verification failed") }
 return nil
}
