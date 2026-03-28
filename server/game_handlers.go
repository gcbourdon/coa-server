package server

import (
	"encoding/json"
	"net/http"

	"coa-server/game"

	"github.com/google/uuid"
)

type createGameRequest struct {
	Player1ID    string    `json:"player1Id"`
	Player2ID    string    `json:"player2Id"`
	DeckP1       []string  `json:"deckP1"`
	DeckP2       []string  `json:"deckP2"`
	StructuresP1 [3]string `json:"structuresP1"`
	StructuresP2 [3]string `json:"structuresP2"`
}

type createGameResponse struct {
	GameID    string `json:"gameId"`
	Player1ID string `json:"player1Id"`
	Player2ID string `json:"player2Id"`
}

// HandleCreateGame creates a new game session and returns the game ID and player IDs.
func (h *Hub) HandleCreateGame(w http.ResponseWriter, r *http.Request) {
	var req createGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Player1ID == "" {
		req.Player1ID = uuid.NewString()
	}
	if req.Player2ID == "" {
		req.Player2ID = uuid.NewString()
	}

	gameID := uuid.NewString()

	gs, err := game.NewGame(gameID, req.Player1ID, req.Player2ID, req.DeckP1, req.DeckP2, req.StructuresP1, req.StructuresP2)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.CreateSession(gs)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(createGameResponse{
		GameID:    gameID,
		Player1ID: req.Player1ID,
		Player2ID: req.Player2ID,
	})
}
