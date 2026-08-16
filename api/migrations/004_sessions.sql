-- Write your migrate up statements here
CREATE TABLE
  sessions (
    id UUID NOT NULL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now (),
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    UNIQUE (token_hash)
  );

CREATE INDEX sessions_user_id_idx ON sessions (user_id);

CREATE INDEX sessions_token_hash_idx ON sessions (token_hash);

---- create above / drop below ----
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
DROP TABLE IF EXISTS sessions;