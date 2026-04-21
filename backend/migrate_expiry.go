package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load(".env")
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer db.Close()

	_, err = db.Exec("ALTER TABLE companies ADD COLUMN IF NOT EXISTS plan_type TEXT DEFAULT 'pro';")
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}

	fmt.Println("Migration successful: added plan_type column")
}
