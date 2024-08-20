-- Drop function
DROP FUNCTION IF EXISTS wuzapi.user_has_permission(INTEGER, TEXT);

-- Drop tables
DROP TABLE IF EXISTS wuzapi.role_permissions;
DROP TABLE IF EXISTS wuzapi.permissions;
DROP TABLE IF EXISTS wuzapi.roles;

-- Remove role_id from users table
ALTER TABLE wuzapi.users
DROP CONSTRAINT IF EXISTS fk_user_role,
DROP COLUMN IF EXISTS role_id;
