CREATE SCHEMA IF NOT EXISTS wuzapi;
ALTER SCHEMA wuzapi OWNER TO wuzapi;
GRANT ALL PRIVILEGES ON SCHEMA wuzapi TO wuzapi;

-- Grant privileges on existing tables (if any)
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA wuzapi TO wuzapi;

-- Grant privileges on existing sequences (if any)
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA wuzapi TO wuzapi;

-- Set default privileges for future tables
ALTER DEFAULT PRIVILEGES IN SCHEMA wuzapi 
GRANT ALL PRIVILEGES ON TABLES TO wuzapi;

-- Set default privileges for future sequences
ALTER DEFAULT PRIVILEGES IN SCHEMA wuzapi 
GRANT ALL PRIVILEGES ON SEQUENCES TO wuzapi;

-- Create the super_organizations table
CREATE TABLE IF NOT EXISTS wuzapi.super_organizations (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create the organizations table
CREATE TABLE IF NOT EXISTS wuzapi.organizations (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create the users table
CREATE TABLE IF NOT EXISTS wuzapi.users (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    token TEXT NOT NULL UNIQUE CHECK(length(token) >= 59 AND length(token) <= 100),
    phone_number TEXT NOT NULL,
    is_phone_number_on_whatsapp BOOLEAN DEFAULT FALSE,
    webhook TEXT DEFAULT '',
    jid TEXT DEFAULT '',
    qrcode TEXT DEFAULT '',
    connected INTEGER,
    expiration INTEGER,
    events TEXT DEFAULT 'All',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_super_admin BOOLEAN DEFAULT FALSE,
    is_admin BOOLEAN DEFAULT FALSE,
    super_organization_id INTEGER,
    CONSTRAINT fk_super_organization FOREIGN KEY (super_organization_id)
        REFERENCES wuzapi.super_organizations(id) ON DELETE RESTRICT
);

-- Create the user_organizations table
CREATE TABLE IF NOT EXISTS wuzapi.user_organizations (
    user_id INTEGER NOT NULL,
    organization_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, organization_id),
    FOREIGN KEY (user_id) REFERENCES wuzapi.users(id) ON DELETE CASCADE,
    FOREIGN KEY (organization_id) REFERENCES wuzapi.organizations(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create the organization_admins table
CREATE TABLE IF NOT EXISTS wuzapi.organization_admins (
    user_id INTEGER NOT NULL,
    organization_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, organization_id),
    FOREIGN KEY (user_id) REFERENCES wuzapi.users(id) ON DELETE CASCADE,
    FOREIGN KEY (organization_id) REFERENCES wuzapi.organizations(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create the super_organization_admins table
CREATE TABLE IF NOT EXISTS wuzapi.super_organization_admins (
    user_id INTEGER NOT NULL,
    super_organization_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, super_organization_id),
    FOREIGN KEY (user_id) REFERENCES wuzapi.users(id) ON DELETE CASCADE,
    FOREIGN KEY (super_organization_id) REFERENCES wuzapi.super_organizations(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for optimization
CREATE INDEX IF NOT EXISTS idx_users_token ON wuzapi.users(token);
CREATE INDEX IF NOT EXISTS idx_users_phone_number ON wuzapi.users(phone_number);
CREATE INDEX IF NOT EXISTS idx_users_is_whatsapp ON wuzapi.users(is_phone_number_on_whatsapp);
CREATE INDEX IF NOT EXISTS idx_user_organizations_user ON wuzapi.user_organizations(user_id);
CREATE INDEX IF NOT EXISTS idx_user_organizations_org ON wuzapi.user_organizations(organization_id);
CREATE INDEX IF NOT EXISTS idx_users_super_admin ON wuzapi.users(is_super_admin);
CREATE INDEX IF NOT EXISTS idx_users_admin ON wuzapi.users(is_admin);
CREATE INDEX IF NOT EXISTS idx_super_org_admins_user ON wuzapi.super_organization_admins(user_id);
CREATE INDEX IF NOT EXISTS idx_super_org_admins_org ON wuzapi.super_organization_admins(super_organization_id);

-- Create the setup_token table
CREATE TABLE IF NOT EXISTS setup_token (
    id SERIAL PRIMARY KEY,
    token TEXT NOT NULL UNIQUE CHECK(length(token) >= 59 AND length(token) <= 100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);

GRANT ALL PRIVILEGES ON TABLE setup_token TO wuzapi;

-- Functions and triggers for data integrity

-- Function to check if users exist
CREATE OR REPLACE FUNCTION wuzapi.users_exist() RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (SELECT 1 FROM wuzapi.users);
END;
$$ LANGUAGE plpgsql;

-- Ensure super admins belong to a super organization
CREATE OR REPLACE FUNCTION wuzapi.ensure_super_admin_belongs_to_super_org() RETURNS TRIGGER AS $$
BEGIN
    IF wuzapi.users_exist() THEN
        IF NEW.is_super_admin AND NEW.super_organization_id IS NULL THEN
            RAISE EXCEPTION 'Super admin must belong to a super organization when users exist';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER enforce_super_admin_super_org
BEFORE INSERT OR UPDATE ON wuzapi.users
FOR EACH ROW EXECUTE FUNCTION wuzapi.ensure_super_admin_belongs_to_super_org();

-- Ensure at least one super organization exists
CREATE OR REPLACE FUNCTION wuzapi.ensure_super_organization_exists() RETURNS TRIGGER AS $$
BEGIN
    IF wuzapi.users_exist() THEN
        IF NOT EXISTS (SELECT 1 FROM wuzapi.super_organizations) THEN
            RAISE EXCEPTION 'At least one super organization must exist when users are present';
        END IF;
    END IF;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER ensure_super_org_exists
AFTER DELETE ON wuzapi.super_organizations
FOR EACH STATEMENT EXECUTE FUNCTION wuzapi.ensure_super_organization_exists();

-- Ensure at least one super admin exists
CREATE OR REPLACE FUNCTION wuzapi.ensure_super_admin_exists() RETURNS TRIGGER AS $$
BEGIN
    IF wuzapi.users_exist() THEN
        IF NOT EXISTS (SELECT 1 FROM wuzapi.users WHERE is_super_admin = TRUE) THEN
            RAISE EXCEPTION 'At least one super admin must exist when users are present';
        END IF;
    END IF;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER prevent_delete_last_super_admin
AFTER DELETE OR UPDATE ON wuzapi.users
FOR EACH STATEMENT EXECUTE FUNCTION wuzapi.ensure_super_admin_exists();

-- Ensure each super organization has an admin
CREATE OR REPLACE FUNCTION wuzapi.ensure_super_org_has_admin() RETURNS TRIGGER AS $$
BEGIN
    IF wuzapi.users_exist() THEN
        IF NOT EXISTS (
            SELECT 1 FROM wuzapi.super_organization_admins
            WHERE super_organization_id = NEW.id
        ) THEN
            RAISE EXCEPTION 'Each super organization must have at least one admin when users exist';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER enforce_super_org_admin
AFTER INSERT OR UPDATE ON wuzapi.super_organizations
FOR EACH ROW EXECUTE FUNCTION wuzapi.ensure_super_org_has_admin();
