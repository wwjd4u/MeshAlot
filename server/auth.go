package server

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

const sessionCookieName = "meshalot_session"
const sessionLifetime = 12 * time.Hour

type userContextKey struct{}

var ErrInvalidCredentials = errors.New("invalid credentials")

func (p *PostgresStore) Authenticate(ctx context.Context, email, password string) (protocol.User, error) {
	var user protocol.User
	err := p.db.QueryRowContext(ctx, `SELECT id::text,email FROM users
        WHERE lower(email)=lower($1) AND password_hash IS NOT NULL
          AND password_hash=crypt($2,password_hash)`, email, password).Scan(&user.ID, &user.Email)
	if errors.Is(err, sql.ErrNoRows) {
		return protocol.User{}, ErrInvalidCredentials
	}
	return user, err
}

func (p *PostgresStore) CreateSession(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	_, err := p.db.ExecContext(ctx, `INSERT INTO user_sessions(user_id,token_hash,expires_at)
        VALUES ($1::uuid,$2,$3)`, userID, tokenHash, expiresAt)
	return err
}

func (p *PostgresStore) SessionUser(ctx context.Context, tokenHash string) (protocol.User, error) {
	var user protocol.User
	err := p.db.QueryRowContext(ctx, `SELECT u.id::text,u.email FROM user_sessions s
        JOIN users u ON u.id=s.user_id
        WHERE s.token_hash=$1 AND s.expires_at>now()`, tokenHash).Scan(&user.ID, &user.Email)
	return user, err
}

func (p *PostgresStore) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := p.db.ExecContext(ctx, "DELETE FROM user_sessions WHERE token_hash=$1", tokenHash)
	return err
}

func newSessionToken() (string, string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw[:])
	sum := sha256.Sum256([]byte(token))
	return token, hex.EncodeToString(sum[:]), nil
}

func hashSessionToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *Service) login(w http.ResponseWriter, r *http.Request) {
	if s.postgres == nil {
		writeError(w, 503, "login unavailable")
		return
	}
	var req protocol.LoginRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
	if decoder.Decode(&req) != nil {
		writeError(w, 400, "invalid login request")
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || len(req.Email) > 320 || req.Password == "" || len(req.Password) > 1024 {
		writeError(w, 400, "email and password are required")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	user, err := s.postgres.Authenticate(ctx, req.Email, req.Password)
	if errors.Is(err, ErrInvalidCredentials) {
		writeError(w, 401, "invalid email or password")
		return
	}
	if err != nil {
		s.databaseError(w, err)
		return
	}
	token, tokenHash, err := newSessionToken()
	if err != nil {
		s.logger.Error("session token generation failed")
		writeError(w, 500, "login failed")
		return
	}
	expires := time.Now().UTC().Add(sessionLifetime)
	if err = s.postgres.CreateSession(ctx, user.ID, tokenHash, expires); err != nil {
		s.databaseError(w, err)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Value: token, Path: "/", Expires: expires, MaxAge: int(sessionLifetime.Seconds()), HttpOnly: true, Secure: isHTTPS(r), SameSite: http.SameSiteLaxMode})
	writeJSON(w, 200, protocol.SessionResponse{User: user})
}

func (s *Service) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil && s.postgres != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if err = s.postgres.DeleteSession(ctx, hashSessionToken(cookie.Value)); err != nil {
			s.databaseError(w, err)
			return
		}
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: isHTTPS(r), SameSite: http.SameSiteLaxMode})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Service) requireSession(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.postgres == nil {
			writeError(w, 503, "login unavailable")
			return
		}
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || cookie.Value == "" {
			writeError(w, 401, "sign in required")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		user, err := s.postgres.SessionUser(ctx, hashSessionToken(cookie.Value))
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, 401, "session expired; sign in again")
			return
		}
		if err != nil {
			s.databaseError(w, err)
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), userContextKey{}, user)))
	}
}

func currentUser(r *http.Request) protocol.User {
	user, _ := r.Context().Value(userContextKey{}).(protocol.User)
	return user
}

func (s *Service) me(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, protocol.SessionResponse{User: currentUser(r)})
}

func isHTTPS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-Proto"), ",")[0]), "https")
}
