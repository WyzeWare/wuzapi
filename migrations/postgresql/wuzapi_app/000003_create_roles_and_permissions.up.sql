-- Create roles table
CREATE TABLE IF NOT EXISTS wuzapi.roles (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Insert roles
INSERT INTO wuzapi.roles (name) VALUES
('super_org_admin'),
('super_org_member'),
('org_admin'),
('ordinary_user');

-- Add role_id to users table
ALTER TABLE wuzapi.users
ADD COLUMN role_id INTEGER,
ADD CONSTRAINT fk_user_role FOREIGN KEY (role_id) REFERENCES wuzapi.roles(id);

-- Create permissions table
CREATE TABLE IF NOT EXISTS wuzapi.permissions (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Insert permissions
INSERT INTO wuzapi.permissions (name) VALUES
('manage_all_users'),
('manage_all_orgs'),
('manage_org_users'),
('manage_self'),
('delete_self');

-- Create role_permissions table
CREATE TABLE IF NOT EXISTS wuzapi.role_permissions (
    role_id INTEGER NOT NULL,
    permission_id INTEGER NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES wuzapi.roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES wuzapi.permissions(id) ON DELETE CASCADE
);

-- Assign permissions to roles
INSERT INTO wuzapi.role_permissions (role_id, permission_id)
VALUES
-- super_org_admin
((SELECT id FROM wuzapi.roles WHERE name = 'super_org_admin'), (SELECT id FROM wuzapi.permissions WHERE name = 'manage_all_users')),
((SELECT id FROM wuzapi.roles WHERE name = 'super_org_admin'), (SELECT id FROM wuzapi.permissions WHERE name = 'manage_all_orgs')),
((SELECT id FROM wuzapi.roles WHERE name = 'super_org_admin'), (SELECT id FROM wuzapi.permissions WHERE name = 'manage_org_users')),
((SELECT id FROM wuzapi.roles WHERE name = 'super_org_admin'), (SELECT id FROM wuzapi.permissions WHERE name = 'manage_self')),
((SELECT id FROM wuzapi.roles WHERE name = 'super_org_admin'), (SELECT id FROM wuzapi.permissions WHERE name = 'delete_self')),

-- super_org_member
((SELECT id FROM wuzapi.roles WHERE name = 'super_org_member'), (SELECT id FROM wuzapi.permissions WHERE name = 'manage_all_users')),
((SELECT id FROM wuzapi.roles WHERE name = 'super_org_member'), (SELECT id FROM wuzapi.permissions WHERE name = 'manage_all_orgs')),
((SELECT id FROM wuzapi.roles WHERE name = 'super_org_member'), (SELECT id FROM wuzapi.permissions WHERE name = 'manage_org_users')),
((SELECT id FROM wuzapi.roles WHERE name = 'super_org_member'), (SELECT id FROM wuzapi.permissions WHERE name = 'manage_self')),
((SELECT id FROM wuzapi.roles WHERE name = 'super_org_member'), (SELECT id FROM wuzapi.permissions WHERE name = 'delete_self')),

-- org_admin
((SELECT id FROM wuzapi.roles WHERE name = 'org_admin'), (SELECT id FROM wuzapi.permissions WHERE name = 'manage_org_users')),
((SELECT id FROM wuzapi.roles WHERE name = 'org_admin'), (SELECT id FROM wuzapi.permissions WHERE name = 'manage_self')),
((SELECT id FROM wuzapi.roles WHERE name = 'org_admin'), (SELECT id FROM wuzapi.permissions WHERE name = 'delete_self')),

-- ordinary_user
((SELECT id FROM wuzapi.roles WHERE name = 'ordinary_user'), (SELECT id FROM wuzapi.permissions WHERE name = 'manage_self')),
((SELECT id FROM wuzapi.roles WHERE name = 'ordinary_user'), (SELECT id FROM wuzapi.permissions WHERE name = 'delete_self'));

-- Create function to check user permissions
CREATE OR REPLACE FUNCTION wuzapi.user_has_permission(user_id INTEGER, permission_name TEXT) 
RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1
        FROM wuzapi.users u
        JOIN wuzapi.role_permissions rp ON u.role_id = rp.role_id
        JOIN wuzapi.permissions p ON rp.permission_id = p.id
        WHERE u.id = user_id AND p.name = permission_name
    );
END;
$$ LANGUAGE plpgsql;

-- Grant necessary privileges
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA wuzapi TO wuzapi;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA wuzapi TO wuzapi;
