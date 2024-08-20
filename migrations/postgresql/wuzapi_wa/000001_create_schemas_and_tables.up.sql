-- Create the schema if it doesn't exist
CREATE SCHEMA IF NOT EXISTS whatsmeow;

-- Grant privileges on schema
GRANT USAGE ON SCHEMA whatsmeow TO wuzapi;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA whatsmeow TO wuzapi;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA whatsmeow TO wuzapi;

-- Set default privileges for future tables and sequences
ALTER DEFAULT PRIVILEGES IN SCHEMA whatsmeow 
GRANT ALL PRIVILEGES ON TABLES TO wuzapi;

ALTER DEFAULT PRIVILEGES IN SCHEMA whatsmeow
GRANT ALL PRIVILEGES ON SEQUENCES TO wuzapi;
