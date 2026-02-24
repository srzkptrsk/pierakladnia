package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"pierakladnia/internal/app"
	"pierakladnia/internal/auth"
	"pierakladnia/internal/db"
	"pierakladnia/internal/render"
)

func Register(deps *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			render.HTML(w, http.StatusOK, "register.html", nil)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		r.ParseForm()
		email := strings.TrimSpace(r.FormValue("email"))
		password := r.FormValue("password")

		if email == "" || len(password) < 6 {
			render.HTML(w, http.StatusBadRequest, "register.html", map[string]string{
				"Error": "Invalid email or password (min 6 chars)",
			})
			return
		}

		hash, err := auth.HashPassword(password)
		if err != nil {
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}

		userID, err := db.CreateUser(deps.DB, email, hash)
		if err != nil {
			render.HTML(w, http.StatusBadRequest, "register.html", map[string]string{
				"Error": "Email already exists or DB error",
			})
			return
		}

		// Generate verification token
		rawToken, _ := auth.GenerateRawToken(32)
		tokenHash := auth.HashToken(rawToken)

		err = db.CreateEmailVerificationToken(deps.DB, tokenHash, userID, time.Now().Add(24*time.Hour))
		if err != nil {
			log.Printf("Failed to create verification token: %v", err)
		} else {
			deps.Mailer.SendVerificationEmail(email, rawToken, deps.Config.HTTP.BaseURL)
		}

		render.HTML(w, http.StatusOK, "login.html", map[string]string{
			"Message": "Registration successful! Please check your email to verify your account.",
		})
	}
}

func VerifyEmail(deps *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "Missing token", http.StatusBadRequest)
			return
		}

		tokenHash := auth.HashToken(token)
		user, err := db.GetUserByVerificationToken(deps.DB, tokenHash)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		if user == nil {
			http.Error(w, "Invalid or expired token", http.StatusBadRequest)
			return
		}

		if err := db.VerifyUser(deps.DB, user.ID, tokenHash); err != nil {
			http.Error(w, "Verification failed", http.StatusInternalServerError)
			return
		}

		render.HTML(w, http.StatusOK, "login.html", map[string]string{
			"Message": "Email verified successfully! You can now log in.",
		})
	}
}

func Login(deps *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// If already logged in
		if user := GetUserFromContext(r.Context()); user != nil {
			http.Redirect(w, r, "/strings", http.StatusFound)
			return
		}

		if r.Method == http.MethodGet {
			render.HTML(w, http.StatusOK, "login.html", nil)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		r.ParseForm()
		email := strings.TrimSpace(r.FormValue("email"))
		password := r.FormValue("password")

		user, err := db.GetUserByEmail(deps.DB, email)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		if user == nil || !auth.CheckPasswordHash(password, user.PasswordHash) {
			render.HTML(w, http.StatusUnauthorized, "login.html", map[string]string{
				"Error": "Invalid credentials",
			})
			return
		}

		if user.EmailVerifiedAt == nil {
			render.HTML(w, http.StatusUnauthorized, "login.html", map[string]string{
				"Error": "Please verify your email address first",
			})
			return
		}

		// Create Session
		sessionID, _ := auth.GenerateRawToken(64)
		expiresAt := time.Now().Add(time.Duration(deps.Config.Auth.SessionTTLHours) * time.Hour)

		err = db.CreateSession(deps.DB, sessionID, user.ID, expiresAt)
		if err != nil {
			http.Error(w, "Internal error creating session", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     deps.Config.Auth.CookieName,
			Value:    sessionID,
			Expires:  expiresAt,
			HttpOnly: true,
			Path:     "/",
		})

		http.Redirect(w, r, "/strings", http.StatusFound)
	}
}

func Logout(deps *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		cookie, err := r.Cookie(deps.Config.Auth.CookieName)
		if err == nil {
			db.DeleteSession(deps.DB, cookie.Value)
		}

		http.SetCookie(w, &http.Cookie{
			Name:     deps.Config.Auth.CookieName,
			Value:    "",
			Expires:  time.Now().Add(-1 * time.Hour),
			HttpOnly: true,
			Path:     "/",
		})

		http.Redirect(w, r, "/login", http.StatusFound)
	}
}
