package server

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/lib/pq"
	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

const enrollmentCodeLifetime = 10 * time.Minute
const maxActiveEnrollmentCodes = 5
const secureEnrollmentAttemptsPerMinute = 120

var ErrInvalidEnrollmentCode = errors.New("invalid enrollment code")
var ErrTooManyActiveEnrollmentCodes = errors.New("too many active enrollment codes")
var ErrNodeIdentityConflict = errors.New("node identity conflict")

var secureEnrollmentRate = struct {
	sync.Mutex
	window   time.Time
	attempts int
}{}

func (s *Service) issueEnrollmentCode(w http.ResponseWriter, r *http.Request) {
	if s.postgres == nil {
		writeError(w, 503, "enrollment unavailable")
		return
	}
	code, tokenHash, err := newEnrollmentCode()
	if err != nil {
		s.logger.Error("enrollment code generation failed")
		writeError(w, 500, "unable to create enrollment code")
		return
	}
	expiresAt := time.Now().UTC().Add(enrollmentCodeLifetime)
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	if err = s.postgres.CreateEnrollmentToken(ctx, currentUser(r).ID, tokenHash, expiresAt); err != nil {
		s.databaseError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusCreated, protocol.EnrollmentCodeResponse{EnrollmentCode: code, ExpiresAt: expiresAt})
}

func (s *Service) secureEnroll(w http.ResponseWriter, r *http.Request) {
	if s.postgres == nil {
		writeError(w, 503, "enrollment unavailable")
		return
	}
	if !allowSecureEnrollment(time.Now().UTC()) {
		writeError(w, http.StatusTooManyRequests, "enrollment temporarily rate limited")
		return
	}
	var req protocol.SecureEnrollRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
	if decoder.Decode(&req) != nil {
		writeError(w, 400, "invalid enrollment request")
		return
	}
	req.NodeID = strings.TrimSpace(req.NodeID)
	req.EnrollmentCode = strings.TrimSpace(req.EnrollmentCode)
	req.AgentVersion = strings.TrimSpace(req.AgentVersion)
	req.PublicKey = strings.TrimSpace(req.PublicKey)
	if !validUUIDv4(req.NodeID) || req.EnrollmentCode == "" || len(req.EnrollmentCode) > 128 || req.AgentVersion == "" || len(req.AgentVersion) > 128 || !validEd25519PublicKey(req.PublicKey) {
		writeError(w, 400, "invalid enrollment request")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	err := s.postgres.RedeemEnrollmentToken(ctx, hashEnrollmentCode(req.EnrollmentCode), req.NodeID, req.PublicKey, req.AgentVersion)
	if errors.Is(err, ErrInvalidEnrollmentCode) {
		writeError(w, 401, "invalid or expired enrollment code")
		return
	}
	if err != nil {
		s.databaseError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	s.logger.Info("secure node enrollment accepted", "node_fingerprint", shortFingerprint(req.NodeID))
	writeJSON(w, http.StatusCreated, protocol.EnrollResponse{NodeID: req.NodeID, Accepted: true})
}

func (p *PostgresStore) CreateEnrollmentToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	var active int
	if err := p.db.QueryRowContext(ctx, `SELECT count(*) FROM enrollment_tokens
        WHERE user_id=$1::uuid AND consumed_at IS NULL AND expires_at>now()`, userID).Scan(&active); err != nil {
		return err
	}
	if active >= maxActiveEnrollmentCodes {
		return ErrTooManyActiveEnrollmentCodes
	}
	_, err := p.db.ExecContext(ctx, `INSERT INTO enrollment_tokens(user_id,token_hash,expires_at)
        VALUES ($1::uuid,$2,$3)`, userID, tokenHash, expiresAt)
	return err
}

func (p *PostgresStore) RedeemEnrollmentToken(ctx context.Context, tokenHash, nodeID, publicKey, agentVersion string) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var tokenID, userID string
	err = tx.QueryRowContext(ctx, `SELECT id::text,user_id::text FROM enrollment_tokens
        WHERE token_hash=$1 AND consumed_at IS NULL AND expires_at>now() FOR UPDATE`, tokenHash).Scan(&tokenID, &userID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalidEnrollmentCode
	}
	if err != nil {
		return err
	}

	var nodeDatabaseID, existingUser string
	var existingPublicKey sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT id::text,user_id::text,identity_public_key
        FROM nodes WHERE node_key=$1 FOR UPDATE`, nodeID).Scan(&nodeDatabaseID, &existingUser, &existingPublicKey)
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRowContext(ctx, `INSERT INTO nodes(user_id,node_key,agent_version,identity_public_key)
            VALUES ($1::uuid,$2,$3,$4) RETURNING id::text`, userID, nodeID, agentVersion, publicKey).Scan(&nodeDatabaseID)
		if err != nil {
			if isUniqueViolation(err) {
				return ErrNodeIdentityConflict
			}
			return err
		}
	} else if err != nil {
		return err
	} else {
		if existingUser != userID {
			return ErrNodeOwnership
		}
		if existingPublicKey.Valid && existingPublicKey.String != "" && existingPublicKey.String != publicKey {
			return ErrNodeIdentityConflict
		}
		_, err = tx.ExecContext(ctx, `UPDATE nodes SET agent_version=$2,
            identity_public_key=COALESCE(identity_public_key,$3) WHERE id=$1::uuid`, nodeDatabaseID, agentVersion, publicKey)
		if err != nil {
			if isUniqueViolation(err) {
				return ErrNodeIdentityConflict
			}
			return err
		}
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO node_status(node_id,status,observed_at)
        VALUES ($1::uuid,'enrolled',now()) ON CONFLICT(node_id) DO NOTHING`, nodeDatabaseID)
	if err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `UPDATE enrollment_tokens SET consumed_at=now(),consumed_node_id=$2::uuid
        WHERE id=$1::uuid AND consumed_at IS NULL`, tokenID, nodeDatabaseID)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return ErrInvalidEnrollmentCode
	}
	return tx.Commit()
}

func newEnrollmentCode() (string, string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", "", err
	}
	code := "mesh_" + base64.RawURLEncoding.EncodeToString(raw[:])
	return code, hashEnrollmentCode(code), nil
}

func hashEnrollmentCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

func validEd25519PublicKey(value string) bool {
	decoded, err := base64.RawStdEncoding.DecodeString(value)
	return err == nil && len(decoded) == ed25519.PublicKeySize
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

func shortFingerprint(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:6])
}

func isUniqueViolation(err error) bool {
	var pqError *pq.Error
	return errors.As(err, &pqError) && pqError.Code == "23505"
}

func allowSecureEnrollment(now time.Time) bool {
	secureEnrollmentRate.Lock()
	defer secureEnrollmentRate.Unlock()
	if secureEnrollmentRate.window.IsZero() || now.Sub(secureEnrollmentRate.window) >= time.Minute || now.Before(secureEnrollmentRate.window) {
		secureEnrollmentRate.window = now
		secureEnrollmentRate.attempts = 0
	}
	if secureEnrollmentRate.attempts >= secureEnrollmentAttemptsPerMinute {
		return false
	}
	secureEnrollmentRate.attempts++
	return true
}
