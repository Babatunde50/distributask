ALTER TABLE users 
    DROP COLUMN username; 

CREATE OR REPLACE FUNCTION update_version_column() 
RETURNS TRIGGER AS $$ 
BEGIN NEW.version := OLD.version + 1;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_version_column
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_version_column();