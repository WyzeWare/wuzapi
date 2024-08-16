package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

type SuperAdminRequest struct {
	Token            string `json:"token" binding:"required"`
	Name             string `json:"name" binding:"required"`
	OrganizationName string `json:"organization_name" binding:"required"`
}

type SetupResult struct {
	UserID     int    `json:"user_id"`
	SuperOrgID int    `json:"super_organization_id"`
	Message    string `json:"message"`
	Token      string `json:"token"`
}

var setupLock sync.Once

func (s *server) setupSuperAdminHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SuperAdminRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Failed to decode request: "+err.Error(), http.StatusBadRequest)
			return
		}

		if !s.validateSetupToken(req.Token) {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		var setupComplete bool
		if err := s.db.QueryRow("SELECT EXISTS (SELECT 1 FROM wuzapi.super_organizations LIMIT 1)").Scan(&setupComplete); err != nil {
			http.Error(w, "Failed to check setup status: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if setupComplete && !s.debugSetup {
			http.Error(w, "Setup has already been completed", http.StatusForbidden)
			return
		}

		var setupResult SetupResult
		var setupErr error
		setupLock.Do(func() {
			setupResult, setupErr = s.performSetup(req)
		})

		if setupErr != nil {
			http.Error(w, setupErr.Error(), http.StatusInternalServerError)
			return
		}

		s.respondWithJSON(w, http.StatusOK, setupResult)
	}
}

func (s *server) performSetup(req SuperAdminRequest) (SetupResult, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return SetupResult{}, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	var superOrgID int
	var userID int

	// Generate a token
	token := generateToken()
	hashedToken := HashToken(token, salt)

	// Create super organization
	err = tx.QueryRow(`
        INSERT INTO wuzapi.super_organizations (name) 
        VALUES ($1) 
        RETURNING id
    `, req.OrganizationName).Scan(&superOrgID)
	if err != nil {
		return SetupResult{}, fmt.Errorf("failed to create super organization: %v", err)
	}

	// Create super admin user with hashed token
	err = tx.QueryRow(`
        INSERT INTO wuzapi.users (name, token, is_super_admin, is_admin, super_organization_id) 
        VALUES ($1, $2, TRUE, TRUE, $3)
        RETURNING id
    `, req.Name, hashedToken, superOrgID).Scan(&userID)
	if err != nil {
		return SetupResult{}, fmt.Errorf("failed to create super admin: %v", err)
	}

	// Add super admin to super_organization_admins
	_, err = tx.Exec(`
        INSERT INTO wuzapi.super_organization_admins (user_id, super_organization_id)
        VALUES ($1, $2)
    `, userID, superOrgID)
	if err != nil {
		return SetupResult{}, fmt.Errorf("failed to add super admin to super organization: %v", err)
	}

	// Commit the transaction if all operations are successful
	if err := tx.Commit(); err != nil {
		return SetupResult{}, fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Delete the setup token from the database
	if err := s.deleteSetupToken(); err != nil {
		return SetupResult{}, fmt.Errorf("failed to delete setup token: %v", err)
	}

	// Return the result including the plain token
	return SetupResult{
		UserID:     userID,
		SuperOrgID: superOrgID,
		Token:      token, // Return the plain token for API usage
		Message:    "Super admin setup completed successfully",
	}, nil
}
func (s *server) validateSetupToken(providedToken string) bool {
	var dbToken string
	var expiresAt time.Time

	// Retrieve the latest setup token and its expiration time from the database
	err := s.db.QueryRow("SELECT token, expires_at FROM setup_token ORDER BY created_at DESC LIMIT 1").Scan(&dbToken, &expiresAt)
	if err != nil {
		// Use your preferred logging method here
		fmt.Printf("Error retrieving setup token from database: %v\n", err)
		return false
	}

	// Compare the provided token with the stored token and check expiration
	isValid, err := VerifyToken(providedToken, dbToken, salt, expiresAt)
	if err != nil {
		fmt.Printf("Error verifying token: %v\n", err)
		return false
	}

	return isValid
}

func (s *server) deleteSetupToken() error {
	_, err := s.db.Exec("DELETE FROM setup_token")
	return err
}

func generateToken() string {
	bytes := make([]byte, 32) // 32 bytes to get a 64-character hex string
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func (s *server) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal JSON response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(response)
}

// HashToken hashes a token with a salt using SHA-256
func HashToken(token, salt string) string {
	hasher := sha256.New()
	hasher.Write([]byte(salt + token))
	return hex.EncodeToString(hasher.Sum(nil))
}

// VerifyToken checks if the provided token is valid
func VerifyToken(providedToken, storedHashedToken, salt string, expiry time.Time) (bool, error) {
	// Decode the base64 encoded hashed token
	decodedHashedToken, err := base64.StdEncoding.DecodeString(storedHashedToken)
	if err != nil {
		return false, fmt.Errorf("error decoding hashed token: %v", err)
	}

	// Hash the provided token
	hashedProvidedToken := HashToken(providedToken, salt)
	// Compare the hashed provided token with the decoded stored hash
	if hashedProvidedToken != string(decodedHashedToken) {
		return false, nil
	}

	// Check if the token has expired
	if time.Now().After(expiry) {
		return false, fmt.Errorf("token has expired")
	}

	return true, nil
}
