-- db/schema.sql
CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    content TEXT NOT NULL
);

-- Use ON CONFLICT to avoid duplicate insertion
INSERT INTO messages (id, content) 
VALUES (1, 'Hello, World!') 
ON CONFLICT (id) DO NOTHING;