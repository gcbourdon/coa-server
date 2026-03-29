package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter builds and returns the HTTP router.
func NewRouter(hub *Hub) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// WebSocket game endpoint
	r.Get("/game/{gameId}", hub.HandleWebSocket)

	// REST API — auth, users, decks, cards
	r.Route("/api/v1", func(r chi.Router) {
		// TODO: auth middleware (JWT) goes here

		r.Post("/games", hub.HandleCreateGame)

		r.Post("/auth/register", notImplemented)
		r.Post("/auth/login", notImplemented)
		r.Post("/auth/logout", notImplemented)

		r.Route("/users", func(r chi.Router) {
			r.Get("/me", notImplemented)
			r.Patch("/me", notImplemented)
		})

		r.Route("/decks", func(r chi.Router) {
			r.Get("/", HandleListDecks)
			r.Post("/", notImplemented)
			r.Get("/{deckId}", HandleGetDeck)
			r.Patch("/{deckId}", notImplemented)
			r.Delete("/{deckId}", notImplemented)
		})

		r.Route("/cards", func(r chi.Router) {
			r.Get("/", HandleListCards)
			r.Get("/{cardId}", HandleGetCard)
		})
	})

	return r
}

func notImplemented(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
