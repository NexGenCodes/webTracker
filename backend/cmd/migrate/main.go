package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"webtracker-bot/internal/config"
	"webtracker-bot/internal/database"
)

func main() {
	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is not set in environment or .env file")
	}

	ctx := context.Background()

	// Connect to the database using the unified connection logic
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer db.Close()

	// Define the SQL directory
	sqlDir := filepath.Join("sql", "migrations")

	// Read all files in the sql directory
	files, err := os.ReadDir(sqlDir)
	if err != nil {
		log.Fatalf("Failed to read sql directory: %v\n", err)
	}

	// Filter for migration files and sort them (e.g. 001_multitenant.sql)
	var sqlFiles []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".sql" {
			// Skip sqlc generated files or schema files that aren't migrations
			if file.Name() == "schema.sql" || file.Name() == "queries.sql" {
				continue
			}
			sqlFiles = append(sqlFiles, file.Name())
		}
	}
	sort.Strings(sqlFiles)

	if len(sqlFiles) == 0 {
		fmt.Println("No SQL files found in the 'sql/migrations' directory.")
		return
	}

	// Create schema_migrations table if it doesn't exist
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create schema_migrations table: %v", err)
	}

	// Execute each SQL file if not already applied
	for _, fileName := range sqlFiles {
		// Check if already applied
		var exists bool
		err = db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", fileName).Scan(&exists)
		if err != nil {
			log.Fatalf("Failed to check migration status for %s: %v", fileName, err)
		}

		if exists {
			fmt.Printf("Skipping already applied migration: %s\n", fileName)
			continue
		}

		filePath := filepath.Join(sqlDir, fileName)
		fmt.Printf("Applying migration: %s...\n", filePath)

		sqlBytes, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to read sql file %s: %v\n", filePath, err)
		}

		// Execute the migration and record it in a transaction
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			log.Fatalf("Failed to begin transaction for %s: %v", fileName, err)
		}

		_, err = tx.ExecContext(ctx, string(sqlBytes))
		if err != nil {
			tx.Rollback()
			log.Fatalf("Failed to execute migration %s: %v\n", filePath, err)
		}

		_, err = tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", fileName)
		if err != nil {
			tx.Rollback()
			log.Fatalf("Failed to record migration %s: %v\n", filePath, err)
		}

		if err := tx.Commit(); err != nil {
			log.Fatalf("Failed to commit migration %s: %v\n", filePath, err)
		}

		fmt.Printf("Successfully applied %s\n", filePath)
	}

	fmt.Println("All pending migrations applied successfully!")

	// Auto-run generate to ensure Go code is in sync
	fmt.Println("Auto-running code generation...")
	root := findRoot()
	if root != "" {
		cmd := exec.Command("go", "run", "cmd/generate/main.go")
		cmd.Dir = root
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Printf("Warning: Auto-generate failed: %v", err)
		}
	} else {
		log.Println("Warning: Could not find backend root for auto-generation")
	}
}

func findRoot() string {
	// Try CWD first
	if _, err := os.Stat("sqlc.yaml"); err == nil {
		abs, _ := filepath.Abs(".")
		return abs
	}
	// Try parent dirs
	dir, _ := os.Getwd()
	for dir != "" && dir != filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "sqlc.yaml")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	return ""
}
