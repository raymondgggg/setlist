-- Write your migrate up statements here
CREATE TABLE
  email_verification_tokens (
    id UUID NOT NULL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now (),
    expires_at timestamptz NOT NULL,
    used_at timestamptz,
    UNIQUE (token_hash)
  );

CREATE INDEX email_verification_tokens_user_id_idx ON email_verification_tokens (user_id);

---- create above / drop below ----
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
DROP TABLE IF EXISTS email_verification_tokens;
