-- migration_001_create_users_table.sql
-- This migration script creates the 'users' table and adds necessary constraints.

BEGIN TRANSACTION;

-- Enable foreign key support in SQLite
PRAGMA foreign_keys = ON;

-- Create the 'users' table if it doesn't already exist
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    token TEXT NOT NULL UNIQUE CHECK(length(token) >= 59 AND length(token) <= 100),
    webhook TEXT DEFAULT '',
    jid TEXT DEFAULT '',
    qrcode TEXT DEFAULT '',
    connected INTEGER,
    expiration INTEGER,
    events TEXT DEFAULT 'All'
);

-- Create a unique index on the 'token' column to enforce uniqueness
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_token ON users(token);

COMMIT;
