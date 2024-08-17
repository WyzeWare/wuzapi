package models

import (
    "context"
    "time"
)

type UserOrganization struct {
    UserID         int       `json:"user_id" db:"user_id"`
    OrganizationID int       `json:"organization_id" db:"organization_id"`
    CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

func (uo *UserOrganization) Create(ctx context.Context, db DB) error {
    query := `INSERT INTO wuzapi.user_organizations (user_id, organization_id) VALUES ($1, $2) RETURNING created_at`
    return db.QueryRowContext(ctx, query, uo.UserID, uo.OrganizationID).Scan(&uo.CreatedAt)
}

func (uo *UserOrganization) Delete(ctx context.Context, db DB) error {
    query := `DELETE FROM wuzapi.user_organizations WHERE user_id = $1 AND organization_id = $2`
    _, err := db.ExecContext(ctx, query, uo.UserID, uo.OrganizationID)
    return err
}

func GetUserOrganizations(ctx context.Context, db DB, userID int) ([]UserOrganization, error) {
    query := `SELECT user_id, organization_id, created_at FROM wuzapi.user_organizations WHERE user_id = $1`
    rows, err := db.QueryContext(ctx, query, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var uos []UserOrganization
    for rows.Next() {
        var uo UserOrganization
        if err := rows.Scan(&uo.UserID, &uo.OrganizationID, &uo.CreatedAt); err != nil {
            return nil, err
        }
        uos = append(uos, uo)
    }
    return uos, nil
}