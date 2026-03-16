package tests

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func TestShipmentScheduling(t *testing.T) {
	_ = godotenv.Load("../.env")
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	// Setup clean schema for testing
	_, _ = db.ExecContext(ctx, "DROP SCHEMA IF EXISTS test_scheduling CASCADE")
	_, err = db.ExecContext(ctx, "CREATE SCHEMA test_scheduling")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = db.ExecContext(ctx, "DROP SCHEMA test_scheduling CASCADE") }()

	_, err = db.ExecContext(ctx, "SET search_path TO test_scheduling, public")
	if err != nil {
		t.Fatal(err)
	}

	// Read and execute migration
	migration, err := os.ReadFile("c:/Users/USER/Desktop/webTracker/database/migrations/001_initial_schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(ctx, string(migration))
	if err != nil {
		t.Fatal(err)
	}

	// Insert country for timezone tests
	_, err = db.ExecContext(ctx, "INSERT INTO country_timezones (country_name, zone_name) VALUES ('usa', 'America/New_York') ON CONFLICT DO NOTHING")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Create Shipment - Midday", func(t *testing.T) {
		// We can't easily mock CURRENT_TIMESTAMP in PG without extensions,
		// but we can check the relative logic.
		// For a live test, we just check if it's currently midday or not.

		var sched, exp time.Time
		err := db.QueryRowContext(ctx, `
			INSERT INTO Shipment (user_jid, destination, origin) 
			VALUES ($1, $2, $3) 
			RETURNING scheduled_transit_time, expected_delivery_time`,
			"test@user", "usa", "nigeria").Scan(&sched, &exp)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("Created Midday - Departure: %v, Arrival: %v", sched, exp)

		if exp.Sub(sched) < 12*time.Hour {
			t.Errorf("Arrival should be at least 12h after departure (next day snap), got diff: %v", exp.Sub(sched))
		}
	})

	t.Run("Manual Update Arrival Sync", func(t *testing.T) {
		// Test if updating departure shifts arrival
		newDeparture := time.Now().AddDate(0, 0, 5).Truncate(time.Hour)

		var trackingID string
		err := db.QueryRowContext(ctx, "INSERT INTO Shipment (user_jid, destination) VALUES ($1, $2) RETURNING tracking_id", "test@user", "usa").Scan(&trackingID)
		if err != nil {
			t.Fatal(err)
		}

		var updatedArrival time.Time
		err = db.QueryRowContext(ctx, `
			UPDATE Shipment SET scheduled_transit_time = $1 WHERE tracking_id = $2 
			RETURNING expected_delivery_time`,
			newDeparture, trackingID).Scan(&updatedArrival)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("Manual Update - New Departure: %v, Updated Arrival: %v", newDeparture, updatedArrival)

		expectedArrivalDate := newDeparture.AddDate(0, 0, 1)
		if updatedArrival.Year() != expectedArrivalDate.Year() || updatedArrival.Month() != expectedArrivalDate.Month() || updatedArrival.Day() != expectedArrivalDate.Day() {
			t.Errorf("Expected arrival on %v, got %v", expectedArrivalDate, updatedArrival)
		}
	})
}
