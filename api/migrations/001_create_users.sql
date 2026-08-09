-- Write your migrate up statements here
CREATE TABLE
  users (
    id UUID PRIMARY KEY,
    first_name text,
    last_name text,
    email text NOT NULL
  );

---- create above / drop below ----
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
DROP TABLE IF EXISTS users;