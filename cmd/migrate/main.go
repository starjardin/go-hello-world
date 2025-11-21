package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/migrate/main.go [up|down|status]")
		fmt.Println("Environment variables required:")
		fmt.Println("  DB_URL - PostgreSQL connection string")
		os.Exit(1)
	}

	command := os.Args[1]

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable is required")
	}

	migrator, err := NewMigrator(dbURL)
	if err != nil {
		log.Fatal("Failed to create migrator:", err)
	}
	defer migrator.Close()

	switch command {
	case "up":
		if err := migrator.Up(); err != nil {
			log.Fatal("Migration up failed:", err)
		}
		fmt.Println("Migration completed successfully!")
	case "down":
		if err := migrator.Down(); err != nil {
			log.Fatal("Migration down failed:", err)
		}
		fmt.Println("Migration rollback completed!")
	case "status":
		if err := migrator.Status(); err != nil {
			log.Fatal("Failed to get migration status:", err)
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available commands: up, down, status")
		os.Exit(1)
	}
}

type Migrator struct {
	db *sql.DB
}

func NewMigrator(dbURL string) (*Migrator, error) {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Migrator{db: db}, nil
}

func (m *Migrator) Close() {
	if m.db != nil {
		m.db.Close()
	}
}

func (m *Migrator) Up() error {
	ctx := context.Background()

	// Create migrations table if it doesn't exist
	if err := m.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Read schema.sql file
	schemaPath := filepath.Join("db", "schema.sql")
	schemaContent, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema.sql: %w", err)
	}

	// Check if migration has already been applied
	migrationName := "001_initial_schema"
	applied, err := m.isMigrationApplied(ctx, migrationName)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if applied {
		fmt.Printf("Migration %s already applied, skipping\n", migrationName)
		return nil
	}

	// Start transaction
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute the schema
	_, err = tx.ExecContext(ctx, string(schemaContent))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	// Record the migration as applied
	_, err = tx.ExecContext(ctx,
		"INSERT INTO schema_migrations (version, applied_at) VALUES ($1, NOW())",
		migrationName)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	fmt.Printf("Applied migration: %s\n", migrationName)
	return nil
}

func (m *Migrator) Down() error {
	ctx := context.Background()
	migrationName := "001_initial_schema"

	// Check if migration has been applied
	applied, err := m.isMigrationApplied(ctx, migrationName)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if !applied {
		fmt.Printf("Migration %s not applied, nothing to rollback\n", migrationName)
		return nil
	}

	// Start transaction
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Drop tables (rollback)
	_, err = tx.ExecContext(ctx, "DROP TABLE IF EXISTS messages")
	if err != nil {
		return fmt.Errorf("failed to drop messages table: %w", err)
	}

	// Remove migration record
	_, err = tx.ExecContext(ctx,
		"DELETE FROM schema_migrations WHERE version = $1",
		migrationName)
	if err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	fmt.Printf("Rolled back migration: %s\n", migrationName)
	return nil
}

func (m *Migrator) Status() error {
	ctx := context.Background()

	// Check if migrations table exists
	var exists bool
	err := m.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = 'schema_migrations'
		)`).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check migrations table: %w", err)
	}

	if !exists {
		fmt.Println("Migrations table does not exist. No migrations have been run.")
		return nil
	}

	// Get applied migrations
	rows, err := m.db.QueryContext(ctx, `
		SELECT version, applied_at 
		FROM schema_migrations 
		ORDER BY applied_at`)
	if err != nil {
		return fmt.Errorf("failed to query migrations: %w", err)
	}
	defer rows.Close()

	fmt.Println("Applied migrations:")
	hasRows := false
	for rows.Next() {
		hasRows = true
		var version string
		var appliedAt string
		if err := rows.Scan(&version, &appliedAt); err != nil {
			return fmt.Errorf("failed to scan migration row: %w", err)
		}
		fmt.Printf("  %s (applied at: %s)\n", version, appliedAt)
	}

	if !hasRows {
		fmt.Println("  No migrations have been applied.")
	}

	return nil
}

func (m *Migrator) createMigrationsTable(ctx context.Context) error {
	_, err := m.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL
		)`)
	return err
}

func (m *Migrator) isMigrationApplied(ctx context.Context, version string) (bool, error) {
	var count int
	err := m.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM schema_migrations WHERE version = $1",
		version).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
