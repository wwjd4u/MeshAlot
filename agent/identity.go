package agent

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const identityFileVersion = 1

type Identity struct {
	Version    int    `json:"version"`
	NodeID     string `json:"node_id"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

func DefaultIdentityPath() (string, error) {
	root, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config directory: %w", err)
	}
	return filepath.Join(root, "meshalot", "agent-identity.json"), nil
}

func LoadOrCreateIdentity(path string) (Identity, bool, error) {
	if strings.TrimSpace(path) == "" {
		return Identity{}, false, errors.New("identity path is required")
	}
	identity, err := LoadIdentity(path)
	if err == nil {
		return identity, false, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return Identity{}, false, err
	}
	identity, err = generateIdentity()
	if err != nil {
		return Identity{}, false, err
	}
	if err = writeIdentity(path, identity); err != nil {
		if errors.Is(err, os.ErrExist) {
			identity, loadErr := LoadIdentity(path)
			return identity, false, loadErr
		}
		return Identity{}, false, err
	}
	return identity, true, nil
}

func LoadIdentity(path string) (Identity, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return Identity{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return Identity{}, errors.New("identity file must be a regular file, not a symlink")
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		return Identity{}, fmt.Errorf("identity file permissions are too broad: %04o (require 0600 or stricter)", info.Mode().Perm())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Identity{}, fmt.Errorf("read identity: %w", err)
	}
	var identity Identity
	if err = json.Unmarshal(data, &identity); err != nil {
		return Identity{}, errors.New("identity file is invalid")
	}
	if err = validateIdentity(identity); err != nil {
		return Identity{}, err
	}
	return identity, nil
}

func generateIdentity() (Identity, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return Identity{}, fmt.Errorf("generate Ed25519 identity: %w", err)
	}
	nodeID, err := newUUIDv4()
	if err != nil {
		return Identity{}, err
	}
	return Identity{
		Version:    identityFileVersion,
		NodeID:     nodeID,
		PublicKey:  base64.RawStdEncoding.EncodeToString(publicKey),
		PrivateKey: base64.RawStdEncoding.EncodeToString(privateKey),
	}, nil
}

func validateIdentity(identity Identity) error {
	if identity.Version != identityFileVersion {
		return fmt.Errorf("unsupported identity file version %d", identity.Version)
	}
	if !validUUIDv4(identity.NodeID) {
		return errors.New("identity node ID is not a UUIDv4")
	}
	publicKey, err := base64.RawStdEncoding.DecodeString(identity.PublicKey)
	if err != nil || len(publicKey) != ed25519.PublicKeySize {
		return errors.New("identity public key is invalid")
	}
	privateKey, err := base64.RawStdEncoding.DecodeString(identity.PrivateKey)
	if err != nil || len(privateKey) != ed25519.PrivateKeySize {
		return errors.New("identity private key is invalid")
	}
	derived := ed25519.PrivateKey(privateKey).Public().(ed25519.PublicKey)
	if !ed25519.PublicKey(publicKey).Equal(derived) {
		return errors.New("identity public and private keys do not match")
	}
	return nil
}

func writeIdentity(path string, identity Identity) error {
	dir := filepath.Dir(path)
	if err := ensurePrivateDirectory(dir); err != nil {
		return err
	}
	data, err := json.MarshalIndent(identity, "", "  ")
	if err != nil {
		return fmt.Errorf("encode identity: %w", err)
	}
	data = append(data, '\n')
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	cleanup := true
	defer func() {
		_ = file.Close()
		if cleanup {
			_ = os.Remove(path)
		}
	}()
	if _, err = file.Write(data); err != nil {
		return fmt.Errorf("write identity: %w", err)
	}
	if err = file.Sync(); err != nil {
		return fmt.Errorf("sync identity: %w", err)
	}
	if err = file.Close(); err != nil {
		return fmt.Errorf("close identity: %w", err)
	}
	cleanup = false
	return nil
}

func ensurePrivateDirectory(dir string) error {
	info, err := os.Stat(dir)
	if errors.Is(err, os.ErrNotExist) {
		if err = os.MkdirAll(dir, 0o700); err != nil {
			return fmt.Errorf("create identity directory: %w", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect identity directory: %w", err)
	}
	if !info.IsDir() {
		return errors.New("identity parent path is not a directory")
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("identity directory permissions are too broad: %04o (require 0700 or stricter)", info.Mode().Perm())
	}
	return nil
}

func PublicKeyFingerprint(identity Identity) (string, error) {
	publicKey, err := base64.RawStdEncoding.DecodeString(identity.PublicKey)
	if err != nil || len(publicKey) != ed25519.PublicKeySize {
		return "", errors.New("identity public key is invalid")
	}
	sum := sha256.Sum256(publicKey)
	return hex.EncodeToString(sum[:8]), nil
}

func NodeIDFingerprint(nodeID string) string {
	sum := sha256.Sum256([]byte(nodeID))
	return hex.EncodeToString(sum[:6])
}

func newUUIDv4() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate node ID: %w", err)
	}
	raw[6] = (raw[6] & 0x0f) | 0x40
	raw[8] = (raw[8] & 0x3f) | 0x80
	hexValue := hex.EncodeToString(raw[:])
	return hexValue[0:8] + "-" + hexValue[8:12] + "-" + hexValue[12:16] + "-" + hexValue[16:20] + "-" + hexValue[20:32], nil
}

func validUUIDv4(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		return false
	}
	raw, err := hex.DecodeString(strings.ReplaceAll(value, "-", ""))
	if err != nil || len(raw) != 16 {
		return false
	}
	return raw[6]>>4 == 4 && raw[8]&0xc0 == 0x80
}
