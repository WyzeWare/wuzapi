-- Remove trigger
DROP TRIGGER IF EXISTS auto_unsuspend_user ON wuzapi.users;

-- Remove function
DROP FUNCTION IF EXISTS wuzapi.check_user_suspension();

-- Remove indexes
DROP INDEX IF EXISTS wuzapi.idx_organizations_suspended;
DROP INDEX IF EXISTS wuzapi.idx_users_suspended;

-- Remove columns from organizations table
ALTER TABLE wuzapi.organizations
DROP COLUMN IF EXISTS is_suspended;

-- Remove columns from users table
ALTER TABLE wuzapi.users
DROP COLUMN IF EXISTS suspension_end_date,
DROP COLUMN IF EXISTS is_suspended;
