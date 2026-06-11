-- Usernames no longer need to be unique (the numeric id identifies users).
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_key;
