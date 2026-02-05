-- -------------------------------------------------------------------------------
-- target DBMS: PostgreSQL
-- Project name: blogginrose
-- -------------------------------------------------------------------------------
--
-- Note: Run this script while connected to the blogginrose database
-- Command: psql -U postgres -d blogginrose -f schema.sql
-- Or: cat schema.sql | docker exec -i <postgres-pod> psql -U postgres -d blogginrose

-- Drop tables if they exist
DROP TABLE IF EXISTS posts;

-- Create Posts table
CREATE TABLE posts (
    post_id SERIAL PRIMARY KEY,
    title VARCHAR(225) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    content TEXT NOT NULL,
    author TEXT NOT NULL,
    published_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    view_count INTEGER NOT NULL DEFAULT 0
);

