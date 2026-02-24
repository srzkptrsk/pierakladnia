package auth

import (
	"database/sql"
	"log"

	"pierakladnia/internal/db"
)

// In a real app we'd inject this via an interface, but for simple MVP
// we use a singleton or pass dependencies explicitly to handlers.
func GetUserFromSession(dbConn *sql.DB, sessionID string) (*db.User, error) {
	if sessionID == "" {
		return nil, nil
	}

	s, err := db.GetSession(dbConn, sessionID)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, nil // not found
	}

	u, err := db.GetUserByID(dbConn, s.UserID)
	if err != nil {
		return nil, err
	}

	if u == nil {
		// Session exists but user deleted?
		log.Printf("Session %s point to non-existent user %d", sessionID, s.UserID)
		return nil, nil
	}

	return u, nil
}
