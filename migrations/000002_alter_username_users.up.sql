ALTER TABLE users
ADD COLUMN telegram_user_id BIGINT NOT NULL UNIQUE;

ALTER TABLE users
ALTER COLUMN username SET NOT NULL;

ALTER TABLE users
DROP COLUMN username


ALTER TABLE links
ADD CONSTRAINT fk_user
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE links
ADD CONSTRAINT unique_user_link
UNIQUE (user_id, original_url);
