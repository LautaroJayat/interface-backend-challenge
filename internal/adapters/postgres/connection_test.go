package postgres_test

import (
	"context"
	"database/sql"
	"os"
	"strconv"
	"testing"
	"time"

	"messaging-app/internal/adapters/postgres"
	"messaging-app/internal/testutils"
)


// --- helpers ---
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	cfg := postgres.DefaultConfig()

	// Allow overriding config via env
	if host := os.Getenv("PGHOST"); host != "" {
		cfg.Host = host
	}
	if portString := os.Getenv("PGPORT"); portString != "" {
		port, err := strconv.Atoi(portString)
		if err != nil {
			t.Logf("could not convert PGPORT into integer, %s. err=%s", portString, err)
			t.FailNow()
		}
		cfg.Port = port
	}
	if user := os.Getenv("PGUSER"); user != "" {
		cfg.User = user
	}
	if pass := os.Getenv("PGPASSWORD"); pass != "" {
		cfg.Password = pass
	}
	if db := os.Getenv("PGDATABASE"); db != "" {
		cfg.Database = db
	}
	db, err := postgres.NewConnection(cfg, testutils.NewTestLogger(t))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	return db

}

func TestNewConnectionIntegration(t *testing.T) {

	db := setupTestDB(t)

	defer db.Close()

	// Basic query test
	var one int
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := db.QueryRowContext(ctx, "SELECT 1").Scan(&one)
	if err != nil {
		t.Fatalf("failed to run test query: %v", err)
	}
	if one != 1 {
		t.Fatalf("expected 1 from SELECT 1, got %d", one)
	}
}
