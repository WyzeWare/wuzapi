-- Drop triggers first to avoid dependency issues
DROP TRIGGER IF EXISTS enforce_super_org_admin ON wuzapi.super_organizations;
DROP TRIGGER IF EXISTS prevent_delete_last_super_admin ON wuzapi.users;
DROP TRIGGER IF EXISTS ensure_super_org_exists ON wuzapi.super_organizations;
DROP TRIGGER IF EXISTS enforce_super_admin_super_org ON wuzapi.users;

-- Drop functions
DROP FUNCTION IF EXISTS wuzapi.ensure_super_org_has_admin();
DROP FUNCTION IF EXISTS wuzapi.ensure_super_admin_exists();
DROP FUNCTION IF EXISTS wuzapi.ensure_super_organization_exists();
DROP FUNCTION IF EXISTS wuzapi.ensure_super_admin_belongs_to_super_org();
DROP FUNCTION IF EXISTS wuzapi.users_exist();

-- Drop indexes
DROP INDEX IF EXISTS idx_super_org_admins_org;
DROP INDEX IF EXISTS idx_users_phone_number;
DROP INDEX IF EXISTS idx_users_is_whatsapp;
DROP INDEX IF EXISTS idx_super_org_admins_user;
DROP INDEX IF EXISTS idx_users_admin;
DROP INDEX IF EXISTS idx_users_super_admin;
DROP INDEX IF EXISTS idx_user_organizations_org;
DROP INDEX IF EXISTS idx_user_organizations_user;
DROP INDEX IF EXISTS idx_users_token;

-- Drop tables
DROP TABLE IF EXISTS wuzapi.super_organization_admins;
DROP TABLE IF EXISTS wuzapi.organization_admins;
DROP TABLE IF EXISTS wuzapi.user_organizations;
DROP TABLE IF EXISTS wuzapi.users;
DROP TABLE IF EXISTS wuzapi.organizations;
DROP TABLE IF EXISTS wuzapi.super_organizations;

-- Drop setup_token table
DROP TABLE IF EXISTS setup_token;

-- Drop the schema
DROP SCHEMA IF EXISTS wuzapi CASCADE;
