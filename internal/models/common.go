package models

import (
    "context"
    "database/sql"
    "fmt"

    _ "github.com/lib/pq" // PostgreSQL driver
)

// DB is an interface that abstracts database operations
type DB interface {
    QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
    QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
    ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// PostgresDB implements the DB interface for PostgreSQL
type PostgresDB struct {
    *sql.DB
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(host, port, user, password, dbname string) (*PostgresDB, error) {
    connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        host, port, user, password, dbname)
    
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    // Test the connection
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    return &PostgresDB{DB: db}, nil
}

// Close closes the database connection
func (pdb *PostgresDB) Close() error {
    return pdb.DB.Close()
}
