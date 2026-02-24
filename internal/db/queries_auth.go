package db

import (
	"database/sql"
	"fmt"
	"time"
)

func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	var u User
	err := db.QueryRow(`
		SELECT id, email, password_hash, role, can_translate, email_verified_at, created_at, updated_at
		FROM users WHERE email = ?
	`, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CanTranslate,
		&u.EmailVerifiedAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func CreateUser(db *sql.DB, email, passwordHash string) (int, error) {
	res, err := db.Exec(`
		INSERT INTO users (email, password_hash) VALUES (?, ?)
	`, email, passwordHash)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func CreateAdminUser(db *sql.DB, email, passwordHash string) (int, error) {
	res, err := db.Exec(`
		INSERT INTO users (email, password_hash, role, can_translate, email_verified_at)
		VALUES (?, ?, 'admin', 1, NOW())
	`, email, passwordHash)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func GetSession(db *sql.DB, sessionID string) (*Session, error) {
	var s Session
	err := db.QueryRow(`
		SELECT id, user_id, expires_at, created_at
		FROM sessions WHERE id = ?
	`, sessionID).Scan(&s.ID, &s.UserID, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func CreateSession(db *sql.DB, sessionID string, userID int, expiresAt time.Time) error {
	_, err := db.Exec(`
		INSERT INTO sessions (id, user_id, expires_at)
		VALUES (?, ?, ?)
	`, sessionID, userID, expiresAt)
	return err
}

func DeleteSession(db *sql.DB, sessionID string) error {
	_, err := db.Exec(`DELETE FROM sessions WHERE id = ?`, sessionID)
	return err
}

func CreateEmailVerificationToken(db *sql.DB, tokenHash string, userID int, expiresAt time.Time) error {
	_, err := db.Exec(`
		INSERT INTO email_verification_tokens (token_hash, user_id, expires_at)
		VALUES (?, ?, ?)
	`, tokenHash, userID, expiresAt)
	return err
}

func GetUserByVerificationToken(db *sql.DB, tokenHash string) (*User, error) {
	var u User
	err := db.QueryRow(`
		SELECT u.id, u.email, u.password_hash, u.role, u.can_translate, u.email_verified_at, u.created_at, u.updated_at
		FROM users u
		JOIN email_verification_tokens t ON t.user_id = u.id
		WHERE t.token_hash = ? AND t.expires_at > NOW()
	`, tokenHash).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CanTranslate,
		&u.EmailVerifiedAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil, nil if token is not found or expired
		}
		return nil, err
	}
	return &u, nil
}

func VerifyUser(db *sql.DB, userID int, tokenHash string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`UPDATE users SET email_verified_at = NOW() WHERE id = ?`, userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM email_verification_tokens WHERE token_hash = ?`, tokenHash)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func GetAllUsers(db *sql.DB) ([]User, error) {
	rows, err := db.Query(`
		SELECT id, email, role, can_translate, email_verified_at, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(
			&u.ID, &u.Email, &u.Role, &u.CanTranslate, &u.EmailVerifiedAt, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func ToggleCanTranslate(db *sql.DB, userID int) error {
	_, err := db.Exec(`UPDATE users SET can_translate = NOT can_translate WHERE id = ?`, userID)
	return err
}

func GetUserByID(db *sql.DB, userID int) (*User, error) {
	var u User
	err := db.QueryRow(`
		SELECT id, email, password_hash, role, can_translate, email_verified_at, created_at, updated_at
		FROM users WHERE id = ?
	`, userID).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CanTranslate,
		&u.EmailVerifiedAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func DeleteUserByEmail(d *sql.DB, email string) error {
	tx, err := d.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var userID int
	err = tx.QueryRow(`SELECT id FROM users WHERE email = ?`, email).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user with email %s not found", email)
		}
		return err
	}

	// Delete related data first
	tx.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID)
	tx.Exec(`DELETE FROM email_verification_tokens WHERE user_id = ?`, userID)

	_, err = tx.Exec(`DELETE FROM users WHERE id = ?`, userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
