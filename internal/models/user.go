package models

import (
	"context"
	"errors"
	"regexp"
	"time"
)

type User struct {
	ID                      int       `json:"id" db:"id"`
	Name                    string    `json:"name" db:"name"`
	Token                   string    `json:"token" db:"token"`
	PhoneNumber             string    `json:"phone_number" db:"phone_number"`
	IsPhoneNumberOnWhatsApp bool      `json:"is_phone_number_on_whatsapp" db:"is_phone_number_on_whatsapp"`
	Webhook                 string    `json:"webhook" db:"webhook"`
	JID                     string    `json:"jid" db:"jid"`
	QRCode                  string    `json:"qrcode" db:"qrcode"`
	Connected               int       `json:"connected" db:"connected"`
	Expiration              int       `json:"expiration" db:"expiration"`
	Events                  string    `json:"events" db:"events"`
	CreatedAt               time.Time `json:"created_at" db:"created_at"`
	IsSuperAdmin            bool      `json:"is_super_admin" db:"is_super_admin"`
	IsAdmin                 bool      `json:"is_admin" db:"is_admin"`
	SuperOrganizationID     *int      `json:"super_organization_id" db:"super_organization_id"`
}

func (u *User) Validate() error {
	if u.Name == "" {
		return errors.New("name cannot be empty")
	}
	if len(u.Token) < 59 || len(u.Token) > 100 {
		return errors.New("token length must be between 59 and 100 characters")
	}
	if u.Webhook != "" {
		if ok, _ := regexp.MatchString(`^https?://`, u.Webhook); !ok {
			return errors.New("webhook must be a valid URL")
		}
	}
	if u.PhoneNumber == "" {
		return errors.New("phone number cannot be empty")
	}
	// You might want to add more specific phone number validation here
	return nil
}

func (u *User) Create(ctx context.Context, db DB) error {
	if err := u.Validate(); err != nil {
		return err
	}
	query := `INSERT INTO wuzapi.users (name, token, phone_number, is_phone_number_on_whatsapp, webhook, jid, qrcode, connected, expiration, events, is_super_admin, is_admin, super_organization_id) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) 
              RETURNING id, created_at`
	return db.QueryRowContext(ctx, query, u.Name, u.Token, u.PhoneNumber, u.IsPhoneNumberOnWhatsApp, u.Webhook, u.JID, u.QRCode, u.Connected, u.Expiration, u.Events, u.IsSuperAdmin, u.IsAdmin, u.SuperOrganizationID).
		Scan(&u.ID, &u.CreatedAt)
}

func (u *User) Update(ctx context.Context, db DB) error {
	if err := u.Validate(); err != nil {
		return err
	}
	query := `UPDATE wuzapi.users SET name = $1, token = $2, phone_number = $3, is_phone_number_on_whatsapp = $4, webhook = $5, jid = $6, qrcode = $7, connected = $8, 
              expiration = $9, events = $10, is_super_admin = $11, is_admin = $12, super_organization_id = $13 
              WHERE id = $14`
	_, err := db.ExecContext(ctx, query, u.Name, u.Token, u.PhoneNumber, u.IsPhoneNumberOnWhatsApp, u.Webhook, u.JID, u.QRCode, u.Connected, u.Expiration, u.Events, u.IsSuperAdmin, u.IsAdmin, u.SuperOrganizationID, u.ID)
	return err
}

func GetUserByID(ctx context.Context, db DB, id int) (*User, error) {
	query := `SELECT id, name, token, phone_number, is_phone_number_on_whatsapp, webhook, jid, qrcode, connected, expiration, events, created_at, is_super_admin, is_admin, super_organization_id 
              FROM wuzapi.users WHERE id = $1`
	var u User
	err := db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Name, &u.Token, &u.PhoneNumber, &u.IsPhoneNumberOnWhatsApp, &u.Webhook, &u.JID, &u.QRCode, &u.Connected, &u.Expiration, &u.Events,
		&u.CreatedAt, &u.IsSuperAdmin, &u.IsAdmin, &u.SuperOrganizationID)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func GetUserByPhoneNumber(ctx context.Context, db DB, phoneNumber string) (*User, error) {
	query := `SELECT id, name, token, phone_number, is_phone_number_on_whatsapp, webhook, jid, qrcode, connected, expiration, events, created_at, is_super_admin, is_admin, super_organization_id 
              FROM wuzapi.users WHERE phone_number = $1`
	var u User
	err := db.QueryRowContext(ctx, query, phoneNumber).Scan(
		&u.ID, &u.Name, &u.Token, &u.PhoneNumber, &u.IsPhoneNumberOnWhatsApp, &u.Webhook, &u.JID, &u.QRCode, &u.Connected, &u.Expiration, &u.Events,
		&u.CreatedAt, &u.IsSuperAdmin, &u.IsAdmin, &u.SuperOrganizationID)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetUsersByOrganization retrieves all users for a given organization
func GetUsersByOrganization(ctx context.Context, db DB, organizationID int) ([]*User, error) {
	query := `SELECT u.id, u.name, u.token, u.webhook, u.jid, u.qrcode, u.connected, u.expiration, u.events, u.created_at, u.is_super_admin, u.is_admin, u.super_organization_id 
              FROM wuzapi.users u
              JOIN wuzapi.user_organizations uo ON u.id = uo.user_id
              WHERE uo.organization_id = $1`

	rows, err := db.QueryContext(ctx, query, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var u User
		if err := rows.Scan(
			&u.ID, &u.Name, &u.Token, &u.Webhook, &u.JID, &u.QRCode, &u.Connected, &u.Expiration, &u.Events,
			&u.CreatedAt, &u.IsSuperAdmin, &u.IsAdmin, &u.SuperOrganizationID); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
