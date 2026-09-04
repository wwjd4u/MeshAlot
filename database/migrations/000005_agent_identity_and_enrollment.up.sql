BEGIN;
ALTER TABLE nodes ADD COLUMN identity_public_key text;
CREATE UNIQUE INDEX nodes_identity_public_key_unique
    ON nodes(identity_public_key)
    WHERE identity_public_key IS NOT NULL;
ALTER TABLE enrollment_tokens ADD COLUMN created_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE enrollment_tokens ADD COLUMN consumed_node_id uuid REFERENCES nodes(id);
CREATE INDEX enrollment_tokens_user_active_idx
    ON enrollment_tokens(user_id,expires_at)
    WHERE consumed_at IS NULL;
COMMIT;
