CREATE TABLE resource_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(6) UNIQUE NOT NULL 
);

CREATE TABLE resources (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    type_id INTEGER REFERENCES resource_types(id),
    content BYTEA
);

-- Insert standard resource types into table 
INSERT INTO resource_types (name) VALUES
    ('script'),
    ('css'),
    ('image')
ON CONFLICT (name) DO NOTHING;