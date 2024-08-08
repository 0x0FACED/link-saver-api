CREATE TABLE links (
    id BIGSERIAL PRIMARY KEY,
    original_link UNIQUE TEXT NOT NULL,
    username VARCHAR(32) NOT NULL
    description VARCHAR(32) UNIQUE NOT NULL
    link_path VARCHAR(128) NOT NULL
    date_added TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
