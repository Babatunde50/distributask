ALTER TABLE users 
    ADD COLUMN username; 

DROP TRIGGER update_version_column ON users;
