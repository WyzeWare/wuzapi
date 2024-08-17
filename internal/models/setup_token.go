package models

import (
    "context"
    "errors"
    "time"
)

type SetupToken struct {
    ID        int       `json:"id" db:"id"`
    Token     string    `json:"token" db:"token"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
}

func (st *SetupToken) Validate() error {
    if len(st.Token) < 59 || len(st.Token) > 100 {
        return errors.New("token length must be between 59 and 100 characters")
    }
    return nil
}

func (st *SetupToken) Create(ctx context.Context, db DB) error {
    if err := st.Validate(); err != nil {
        return err
    }
    query := `INSERT INTO setup_token (token, expires_at) VALUES ($1, $2) RETURNING id, created_at`
    return db.QueryRowContext(ctx, query, st.Token, st.ExpiresAt).Scan(&st.ID, &st.CreatedAt)
}

func (st *SetupToken) Update(ctx context.Context, db DB) error {
    if err := st.Validate(); err != nil {
        return err
    }
    query := `UPDATE setup_token SET token = $1, expires_at = $2 WHERE id = $3`
    _, err := db.ExecContext(ctx, query, st.Token, st.ExpiresAt, st.ID)
    return err
}

func (st *SetupToken) Delete(ctx context.Context, db DB) error {
    query := `DELETE FROM setup_token WHERE id = $1`
    _, err := db.ExecContext(ctx, query, st.ID)
    return err
}

func GetSetupTokenByToken(ctx context.Context, db DB, token string) (*SetupToken, error) {
    query := `SELECT id, token, created_at, expires_at FROM setup_token WHERE token = $1`
    st := &SetupToken{}
    err := db.QueryRowContext(ctx, query, token).Scan(&st.ID, &st.Token, &st.CreatedAt, &st.ExpiresAt)
    if err != nil {
        return nil, err
    }
    return st, nil
}

func DeleteExpiredSetupTokens(ctx context.Context, db DB) (int64, error) {
    query := `DELETE FROM setup_token WHERE expires_at < CURRENT_TIMESTAMP`
    result, err := db.ExecContext(ctx, query)
    if err != nil {
        return 0, err
    }
    return result.RowsAffected()
}