package app

import (
	"database/sql"
	"pierakladnia/internal/auth"
	"pierakladnia/internal/config"
)

type MailSender interface {
	SendVerificationEmail(toEmail, token, baseURL string) error
}

type App struct {
	Config   *config.Config
	DB       *sql.DB
	Mailer   MailSender
	Sessions auth.SessionResolver
}
