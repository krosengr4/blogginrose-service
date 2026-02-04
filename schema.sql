-- -------------------------------------------------------------------------------
-- target DBMS: PostgresSQL
-- Project name: rosenblog
-- -------------------------------------------------------------------------------
-- 
-- Note: Run this script while connected to the rosenblog_db database
-- Command: psql -U postgres -d rosenblog_db -f database.sql
-- Or: cat database.sql | docker exec -i rosenblog-db psql -U postgres -d rosenblog_db

-- Drop tables if they exist
DROP TABLE IF EXISTS posts;

-- Create Posts table
CREATE TABLE posts (
    post_id SERIAL PRIMARY KEY,
    title VARCHAR(225) NOT NULL,
    content TEXT NOT NULL,
    date_posted TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)
