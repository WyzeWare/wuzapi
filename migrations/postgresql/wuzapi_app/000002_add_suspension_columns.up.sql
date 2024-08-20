-- Add suspension columns to users table
ALTER TABLE wuzapi.users
ADD COLUMN is_suspended BOOLEAN DEFAULT FALSE,
ADD COLUMN suspension_end_date TIMESTAMP WITH TIME ZONE;

-- Add suspension column to organizations table
ALTER TABLE wuzapi.organizations
ADD COLUMN is_suspended BOOLEAN DEFAULT FALSE;

-- Create index for faster queries on suspended users
CREATE INDEX idx_users_suspended ON wuzapi.users(is_suspended);

-- Create index for faster queries on suspended organizations
CREATE INDEX idx_organizations_suspended ON wuzapi.organizations(is_suspended);

-- Grant privileges on the new columns
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA wuzapi TO wuzapi;

-- Function to automatically unsuspend users when suspension period ends
CREATE OR REPLACE FUNCTION wuzapi.check_user_suspension() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.is_suspended AND NEW.suspension_end_date IS NOT NULL AND NEW.suspension_end_date <= CURRENT_TIMESTAMP THEN
        NEW.is_suspended := FALSE;
        NEW.suspension_end_date := NULL;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically unsuspend users
CREATE TRIGGER auto_unsuspend_user
BEFORE UPDATE ON wuzapi.users
FOR EACH ROW
EXECUTE FUNCTION wuzapi.check_user_suspension();
