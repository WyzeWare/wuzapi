package models

import (
    "context"
    "errors"
    "time"
)

type SuperOrganization struct {
    ID        int       `json:"id" db:"id"`
    Name      string    `json:"name" db:"name"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (so *SuperOrganization) Validate() error {
    if so.Name == "" {
        return errors.New("name cannot be empty")
    }
    return nil
}

func (so *SuperOrganization) Create(ctx context.Context, db DB) error {
    if err := so.Validate(); err != nil {
        return err
    }
    query := `INSERT INTO wuzapi.super_organizations (name) VALUES ($1) RETURNING id, created_at`
    return db.QueryRowContext(ctx, query, so.Name).Scan(&so.ID, &so.CreatedAt)
}

func (so *SuperOrganization) Update(ctx context.Context, db DB) error {
    if err := so.Validate(); err != nil {
        return err
    }
    query := `UPDATE wuzapi.super_organizations SET name = $1 WHERE id = $2`
    _, err := db.ExecContext(ctx, query, so.Name, so.ID)
    return err
}

func (so *SuperOrganization) Delete(ctx context.Context, db DB) error {
    query := `DELETE FROM wuzapi.super_organizations WHERE id = $1`
    _, err := db.ExecContext(ctx, query, so.ID)
    return err
}

func GetSuperOrganizations(ctx context.Context, db DB) ([]SuperOrganization, error) {
    query := `SELECT id, name, created_at FROM wuzapi.super_organizations`
    rows, err := db.QueryContext(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var orgs []SuperOrganization
    for rows.Next() {
        var org SuperOrganization
        if err := rows.Scan(&org.ID, &org.Name, &org.CreatedAt); err != nil {
            return nil, err
        }
        orgs = append(orgs, org)
    }
    return orgs, nil
}

func GetSuperOrganizationByID(ctx context.Context, db DB, id int) (*SuperOrganization, error) {
    query := `SELECT id, name, created_at FROM wuzapi.super_organizations WHERE id = $1`
    so := &SuperOrganization{}
    err := db.QueryRowContext(ctx, query, id).Scan(&so.ID, &so.Name, &so.CreatedAt)
    if err != nil {
        return nil, err
    }
    return so, nil
}