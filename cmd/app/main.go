package main

import (
	"log"
	"net/http"

	"pierakladnia/internal/app"
	"pierakladnia/internal/config"
	"pierakladnia/internal/db"
	myhttp "pierakladnia/internal/http"
	"pierakladnia/internal/mail"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	database, err := db.NewDB(cfg.DB.DSN)
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}

	mailer, err := mail.NewSESSender(cfg.SES.Region, cfg.SES.FromEmail, cfg.SES.AccessKeyID, cfg.SES.SecretAccessKey)
	if err != nil {
		log.Fatalf("Failed to init SES mailer: %v", err)
	}

	appDeps := &app.App{
		Config: cfg,
		DB:     database,
		Mailer: mailer,
	}

	router := myhttp.NewRouter(appDeps)

	log.Printf("Starting server on %s", cfg.HTTP.Addr)
	if err := http.ListenAndServe(cfg.HTTP.Addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
