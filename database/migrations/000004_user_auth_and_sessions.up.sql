BEGIN;
ALTER TABLE users ADD COLUMN password_hash text;
CREATE UNIQUE INDEX users_email_lower_unique ON users ((lower(email)));
CREATE TABLE user_sessions(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash text UNIQUE NOT NULL CHECK (token_hash ~ '^[0-9a-f]{64}$'),
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX user_sessions_user_id_idx ON user_sessions(user_id);
CREATE INDEX user_sessions_expires_at_idx ON user_sessions(expires_at);
REVOKE ALL ON user_sessions FROM PUBLIC;
COMMIT;
