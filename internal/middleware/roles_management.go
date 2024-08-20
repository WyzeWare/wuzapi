package dbmiddleware

import (
	"database/sql"
	"errors"
	"fmt"
)

type DBMiddleware struct {
	db *sql.DB
}

func NewDBMiddleware(db *sql.DB) *DBMiddleware {
	return &DBMiddleware{db: db}
}

func (m *DBMiddleware) CheckPermission(userID int, permission string) (bool, error) {
	var hasPermission bool
	err := m.db.QueryRow("SELECT wuzapi.user_has_permission($1, $2)", userID, permission).Scan(&hasPermission)
	if err != nil {
		return false, fmt.Errorf("error checking permission: %w", err)
	}
	return hasPermission, nil
}

func (m *DBMiddleware) ManageUser(actorID, targetUserID int, action string) error {
	hasPermission, err := m.CheckPermission(actorID, "manage_all_users")
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("permission denied: cannot manage user")
	}

	// Perform the user management action here
	// For example: UPDATE wuzapi.users SET ... WHERE id = $1
	return nil
}

func (m *DBMiddleware) ManageOrganization(actorID, orgID int, action string) error {
	hasPermission, err := m.CheckPermission(actorID, "manage_all_orgs")
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("permission denied: cannot manage organization")
	}

	// Perform the organization management action here
	// For example: UPDATE wuzapi.organizations SET ... WHERE id = $1
	return nil
}

func (m *DBMiddleware) ManageOrgUser(actorID, targetUserID, orgID int, action string) error {
	hasPermission, err := m.CheckPermission(actorID, "manage_org_users")
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("permission denied: cannot manage organization user")
	}

	// Check if the actor is an admin of the specific organization
	var isOrgAdmin bool
	err = m.db.QueryRow("SELECT EXISTS(SELECT 1 FROM wuzapi.organization_admins WHERE user_id = $1 AND organization_id = $2)", actorID, orgID).Scan(&isOrgAdmin)
	if err != nil {
		return fmt.Errorf("error checking org admin status: %w", err)
	}
	if !isOrgAdmin {
		return errors.New("permission denied: not an admin of this organization")
	}

	// Perform the organization user management action here
	// For example: INSERT INTO wuzapi.user_organizations (user_id, organization_id) VALUES ($1, $2)
	return nil
}

func (m *DBMiddleware) ManageSelf(userID int, action string) error {
	hasPermission, err := m.CheckPermission(userID, "manage_self")
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("permission denied: cannot manage self")
	}

	// Perform the self-management action here
	// For example: UPDATE wuzapi.users SET ... WHERE id = $1
	return nil
}

func (m *DBMiddleware) DeleteSelf(userID int) error {
	hasPermission, err := m.CheckPermission(userID, "delete_self")
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("permission denied: cannot delete self")
	}

	// Perform the user deletion action here
	// For example: DELETE FROM wuzapi.users WHERE id = $1
	return nil
}
