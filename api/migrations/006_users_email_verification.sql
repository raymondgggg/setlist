-- Write your migrate up statements here
UPDATE users
SET
  email = lower(email);

ALTER TABLE users
ADD COLUMN email_verified_at timestamptz,
ADD CONSTRAINT users_email_key UNIQUE (email),
ADD CONSTRAINT users_email_lowercase CHECK (email = lower(email));

---- create above / drop below ----
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
ALTER TABLE users
DROP CONSTRAINT users_email_lowercase,
DROP CONSTRAINT users_email_key,
DROP COLUMN email_verified_at;