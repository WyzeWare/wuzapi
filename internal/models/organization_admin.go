package models

import (
	"context"
	"time"
)

type OrganizationAdmin struct {
	UserID         int       `json:"user_id" db:"user_id"`
	OrganizationID int       `json:"organization_id" db:"organization_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

func (oa *OrganizationAdmin) Create(ctx context.Context, db DB) error {
	query := `INSERT INTO wuzapi.organization_admins (user_id, organization_id) VALUES ($1, $2) RETURNING created_at`
	return db.QueryRowContext(ctx, query, oa.UserID, oa.OrganizationID).Scan(&oa.CreatedAt)
}

func (oa *OrganizationAdmin) Delete(ctx context.Context, db DB) error {
	query := `DELETE FROM wuzapi.organization_admins WHERE user_id = $1 AND organization_id = $2`
	_, err := db.ExecContext(ctx, query, oa.UserID, oa.OrganizationID)
	return err
}

func GetOrganizationAdmins(ctx context.Context, db DB, organizationID int) ([]OrganizationAdmin, error) {
	query := `SELECT user_id, organization_id, created_at FROM wuzapi.organization_admins WHERE organization_id = $1`
	rows, err := db.QueryContext(ctx, query, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var oas []OrganizationAdmin
	for rows.Next() {
		var oa OrganizationAdmin
		if err := rows.Scan(&oa.UserID, &oa.OrganizationID, &oa.CreatedAt); err != nil {
			return nil, err
		}
		oas = append(oas, oa)
	}
	return oas, nil
}

func IsOrganizationAdmin(ctx context.Context, db DB, userID, organizationID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM wuzapi.organization_admins WHERE user_id = $1 AND organization_id = $2)`
	var exists bool
	err := db.QueryRowContext(ctx, query, userID, organizationID).Scan(&exists)
	return exists, err
}
