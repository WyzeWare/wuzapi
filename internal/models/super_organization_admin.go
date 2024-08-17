package models

import (
    "context"
    "time"
)

type SuperOrganizationAdmin struct {
    UserID              int       `json:"user_id" db:"user_id"`
    SuperOrganizationID int       `json:"super_organization_id" db:"super_organization_id"`
    CreatedAt           time.Time `json:"created_at" db:"created_at"`
}

func (soa *SuperOrganizationAdmin) Create(ctx context.Context, db DB) error {
    query := `INSERT INTO wuzapi.super_organization_admins (user_id, super_organization_id) VALUES ($1, $2) RETURNING created_at`
    return db.QueryRowContext(ctx, query, soa.UserID, soa.SuperOrganizationID).Scan(&soa.CreatedAt)
}

func (soa *SuperOrganizationAdmin) Delete(ctx context.Context, db DB) error {
    query := `DELETE FROM wuzapi.super_organization_admins WHERE user_id = $1 AND super_organization_id = $2`
    _, err := db.ExecContext(ctx, query, soa.UserID, soa.SuperOrganizationID)
    return err
}

func GetSuperOrganizationAdmins(ctx context.Context, db DB, superOrganizationID int) ([]SuperOrganizationAdmin, error) {
    query := `SELECT user_id, super_organization_id, created_at FROM wuzapi.super_organization_admins WHERE super_organization_id = $1`
    rows, err := db.QueryContext(ctx, query, superOrganizationID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var soas []SuperOrganizationAdmin
    for rows.Next() {
        var soa SuperOrganizationAdmin
        if err := rows.Scan(&soa.UserID, &soa.SuperOrganizationID, &soa.CreatedAt); err != nil {
            return nil, err
        }
        soas = append(soas, soa)
    }
    return soas, nil
}

func IsSuperOrganizationAdmin(ctx context.Context, db DB, userID, superOrganizationID int) (bool, error) {
    query := `SELECT EXISTS(SELECT 1 FROM wuzapi.super_organization_admins WHERE user_id = $1 AND super_organization_id = $2)`
    var exists bool
    err := db.QueryRowContext(ctx, query, userID, superOrganizationID).Scan(&exists)
    return exists, err
}