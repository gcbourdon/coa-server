package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"coa-server/cards"
	"coa-server/db"
	"coa-server/server"
)

//go:embed migrations/*.up.sql
var migrationsFS embed.FS

func main() {
	addr := envOr("ADDR", ":8080")
	dbDSN := envOr("DATABASE_URL", "")
	cardsRoot := envOr("CARDS_DIR", "../cards")

	// Load card definitions from the cards/ data directory.
	if err := cards.LoadAll(cardsRoot); err != nil {
		log.Fatalf("loading cards: %v", err)
	}
	if err := cards.LoadDecks(cardsRoot); err != nil {
		log.Fatalf("loading decks: %v", err)
	}
	log.Printf("cards and decks loaded from %s", cardsRoot)

	// Connect to PostgreSQL if a DSN is provided.
	// In Phase 1 (engine-only testing) it's fine to omit DATABASE_URL.
	if dbDSN != "" {
		if err := db.Connect(context.Background(), dbDSN); err != nil {
			log.Fatalf("db connect: %v", err)
		}
		defer db.Close()
		log.Println("connected to postgres")

		migFS, _ := fs.Sub(migrationsFS, "migrations")
		if err := db.Migrate(context.Background(), migFS); err != nil {
			log.Fatalf("db migrate: %v", err)
		}
		log.Println("migrations applied")
	} else {
		log.Println("DATABASE_URL not set — running without database (game engine only)")
	}

	hub := server.NewHub()
	router := server.NewRouter(hub)

	log.Printf("coa-server listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
