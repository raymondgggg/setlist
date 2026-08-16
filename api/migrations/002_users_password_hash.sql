-- Write your migrate up statements here
ALTER TABLE users
ADD COLUMN password_hash text;

---- create above / drop below ----
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
ALTER TABLE users
DROP COLUMN password_hash;