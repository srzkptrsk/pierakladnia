package auth

import (
	"database/sql"
	"log"

	"pierakladnia/internal/db"
)

// SessionResolver looks up the authenticated user from a session ID.
type SessionResolver interface {
	GetUserFromSession(sessionID string) (*db.User, error)
}

// dbSessionResolver is the concrete implementation backed by *sql.DB.
type dbSessionResolver struct {
	dbConn *sql.DB
}

// NewSessionResolver creates a SessionResolver backed by the given database.
func NewSessionResolver(dbConn *sql.DB) SessionResolver {
	return &dbSessionResolver{dbConn: dbConn}
}

func (s *dbSessionResolver) GetUserFromSession(sessionID string) (*db.User, error) {
	if sessionID == "" {
		return nil, nil
	}

	sess, err := db.GetSession(s.dbConn, sessionID)
	if err != nil {
		return nil, err
	}
	if sess == nil {
		return nil, nil // not found
	}

	u, err := db.GetUserByID(s.dbConn, sess.UserID)
	if err != nil {
		return nil, err
	}

	if u == nil {
		// Session exists but user deleted?
		log.Printf("Session %s points to non-existent user %d", sessionID, sess.UserID)
		return nil, nil
	}

	return u, nil
}
