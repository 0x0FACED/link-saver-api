
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(32) NOT NULL UNIQUE
);

CREATE TABLE links (
    id BIGSERIAL PRIMARY KEY,
    original_url TEXT NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    description VARCHAR(32) NOT NULL,
    content BYTEA NOT NULL,
    date_added TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, original_url)
);
