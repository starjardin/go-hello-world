// main.go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	db "github.com/starjardin/hello-world/db/generated"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Connect to PostgreSQL

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable is required")
	}

	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal("Cannot connect to DB:", err)
	}
	defer dbpool.Close()

	// Wrap pgxpool to satisfy sqlc's interface
	queries := db.New(dbpool)

	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		// Get message with id=1
		content, err := queries.GetMessage(ctx, 1)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Message not found", http.StatusNotFound)
				return
			}
			http.Error(w, fmt.Errorf("DB error was found here: %w", err).Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(content))
	})

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
