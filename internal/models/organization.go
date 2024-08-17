package models

import (
    "context"
    "errors"
    "time"
)

type Organization struct {
    ID        int       `json:"id" db:"id"`
    Name      string    `json:"name" db:"name"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (o *Organization) Validate() error {
    if o.Name == "" {
        return errors.New("name cannot be empty")
    }
    return nil
}

func (o *Organization) Create(ctx context.Context, db DB) error {
    if err := o.Validate(); err != nil {
        return err
    }
    query := `INSERT INTO wuzapi.organizations (name) VALUES ($1) RETURNING id, created_at`
    return db.QueryRowContext(ctx, query, o.Name).Scan(&o.ID, &o.CreatedAt)
}

func (o *Organization) Update(ctx context.Context, db DB) error {
    if err := o.Validate(); err != nil {
        return err
    }
    query := `UPDATE wuzapi.organizations SET name = $1 WHERE id = $2`
    _, err := db.ExecContext(ctx, query, o.Name, o.ID)
    return err
}

func (o *Organization) Delete(ctx context.Context, db DB) error {
    query := `DELETE FROM wuzapi.organizations WHERE id = $1`
    _, err := db.ExecContext(ctx, query, o.ID)
    return err
}

func GetOrganizations(ctx context.Context, db DB) ([]Organization, error) {
    query := `SELECT id, name, created_at FROM wuzapi.organizations`
    rows, err := db.QueryContext(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var orgs []Organization
    for rows.Next() {
        var org Organization
        if err := rows.Scan(&org.ID, &org.Name, &org.CreatedAt); err != nil {
            return nil, err
        }
        orgs = append(orgs, org)
    }
    return orgs, nil
}

func GetOrganizationByID(ctx context.Context, db DB, id int) (*Organization, error) {
    query := `SELECT id, name, created_at FROM wuzapi.organizations WHERE id = $1`
    o := &Organization{}
    err := db.QueryRowContext(ctx, query, id).Scan(&o.ID, &o.Name, &o.CreatedAt)
    if err != nil {
        return nil, err
    }
    return o, nil
}