package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"pierakladnia/internal/auth"
	"pierakladnia/internal/config"
	"pierakladnia/internal/db"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "create-admin":
		createAdmin(os.Args[2:])
	case "delete-user":
		deleteUser(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: admin <command> [flags]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  create-admin  Create a new admin user")
	fmt.Fprintln(os.Stderr, "  delete-user   Delete a user by email")
}

func createAdmin(args []string) {
	fs := flag.NewFlagSet("create-admin", flag.ExitOnError)
	email := fs.String("email", "", "Admin email address (required)")
	password := fs.String("password", "", "Admin password (required)")
	fs.Parse(args)

	if *email == "" || *password == "" {
		fmt.Fprintln(os.Stderr, "Error: --email and --password are required")
		fs.Usage()
		os.Exit(1)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	database, err := db.NewDB(cfg.DB.DSN)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer database.Close()

	// Check if user already exists
	existing, err := db.GetUserByEmail(database, *email)
	if err != nil {
		log.Fatalf("Failed to check existing user: %v", err)
	}
	if existing != nil {
		log.Fatalf("User with email %s already exists (id=%d, role=%s)", *email, existing.ID, existing.Role)
	}

	hash, err := auth.HashPassword(*password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	id, err := db.CreateAdminUser(database, *email, hash)
	if err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	fmt.Printf("Admin user created successfully: id=%d, email=%s\n", id, *email)
}

func deleteUser(args []string) {
	fs := flag.NewFlagSet("delete-user", flag.ExitOnError)
	email := fs.String("email", "", "User email address (required)")
	fs.Parse(args)

	if *email == "" {
		fmt.Fprintln(os.Stderr, "Error: --email is required")
		fs.Usage()
		os.Exit(1)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	database, err := db.NewDB(cfg.DB.DSN)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer database.Close()

	err = db.DeleteUserByEmail(database, *email)
	if err != nil {
		log.Fatalf("Failed to delete user: %v", err)
	}

	fmt.Printf("User %s deleted successfully\n", *email)
}
