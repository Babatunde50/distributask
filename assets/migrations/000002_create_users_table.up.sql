CREATE EXTENSION IF NOT EXISTS citext;
CREATE TABLE users (
    id SERIAL NOT NULL PRIMARY KEY,
    email citext NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    version integer NOT NULL DEFAULT 1
);
CREATE OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS $$ BEGIN NEW.updated_at = NOW();
RETURN NEW;
END;

$$ LANGUAGE plpgsql;

CREATE TRIGGER update_updated_at BEFORE
UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE OR REPLACE FUNCTION update_version_column() RETURNS TRIGGER AS $$ BEGIN NEW.version := OLD.version + 1;
RETURN NEW;
END;

$$ LANGUAGE plpgsql;
CREATE TRIGGER update_version_column BEFORE
UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_version_column();