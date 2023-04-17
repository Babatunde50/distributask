CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
    id SERIAL NOT NULL PRIMARY KEY,
    email citext NOT NULL UNIQUE,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    role TEXT NOT NULL DEFAULT 'user',
    version integer NOT NULL DEFAULT 1,
    activated bool NOT NULL DEFAULT false
);

CREATE OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS $$ BEGIN NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();