-- Write your migrate up statements here
CREATE FUNCTION populate_updated_at()
   RETURNS TRIGGER
   LANGUAGE PLPGSQL
AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$;

ALTER TABLE users
ADD COLUMN created_at timestamptz NOT NULL DEFAULT now(),
ADD COLUMN updated_at timestamptz NOT NULL DEFAULT now();

CREATE TRIGGER users_updated_at_tgr BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION populate_updated_at();
CREATE TRIGGER connected_accounts_updated_at_tgr BEFORE UPDATE ON connected_accounts FOR EACH ROW EXECUTE FUNCTION populate_updated_at();

---- create above / drop below ----
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
DROP TRIGGER users_updated_at_tgr;
DROP TRIGGER connected_accounts_updated_at_tgr;

ALTER TABLE users
DROP COLUMN updated_at;

DROP FUNCTION populate_updated_at;
