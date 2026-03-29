package server

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"coa-server/cards"

	"github.com/go-chi/chi/v5"
)

// cardResponse wraps CardDef for API output, adding the derived ID field.
type cardResponse struct {
	ID string `json:"id"`
	*cards.CardDef
}

func toCardResponse(def *cards.CardDef) cardResponse {
	return cardResponse{ID: def.ID, CardDef: def}
}

// sortedCards returns all cards sorted deterministically by set_id then number.
func sortedCards() []*cards.CardDef {
	all := cards.AllCards()
	sort.Slice(all, func(i, j int) bool {
		if all[i].SetID != all[j].SetID {
			return all[i].SetID < all[j].SetID
		}
		return all[i].Number < all[j].Number
	})
	return all
}

// HandleListCards handles GET /api/v1/cards
//
// When the `cardIDs` query parameter is present (comma-separated list of card IDs),
// it returns those specific cards without pagination.
//
// Otherwise it returns a paginated list of all cards.
// Query params: page (1-indexed, default 1), pageSize (default 20, max 100).
func HandleListCards(w http.ResponseWriter, r *http.Request) {
	if raw := r.URL.Query().Get("cardIDs"); raw != "" {
		handleCardsByIDs(w, raw)
		return
	}

	page, pageSize := parsePagination(r)

	all := sortedCards()
	total := len(all)

	start := (page - 1) * pageSize
	if start >= total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	slice := all[start:end]

	resp := make([]cardResponse, len(slice))
	for i, c := range slice {
		resp[i] = toCardResponse(c)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"cards":    resp,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// handleCardsByIDs is the sub-handler for GET /api/v1/cards?cardIDs=OGN_1,OGN_2,...
func handleCardsByIDs(w http.ResponseWriter, raw string) {
	ids := strings.Split(raw, ",")
	resp := make([]cardResponse, 0, len(ids))
	var missing []string

	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		def, err := cards.GetCard(id)
		if err != nil {
			missing = append(missing, id)
			continue
		}
		resp = append(resp, toCardResponse(def))
	}

	if len(missing) > 0 {
		http.Error(w, "unknown card IDs: "+strings.Join(missing, ", "), http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"cards": resp})
}

// HandleGetCard handles GET /api/v1/cards/{cardId}
func HandleGetCard(w http.ResponseWriter, r *http.Request) {
	cardID := chi.URLParam(r, "cardId")
	def, err := cards.GetCard(cardID)
	if err != nil {
		http.Error(w, "card not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, toCardResponse(def))
}

// deckSummary is the list-view representation of a deck (no cards array).
type deckSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Archetype   string `json:"archetype"`
	Description string `json:"description"`
}

// HandleListDecks handles GET /api/v1/decks
// Returns a summary list (id, name, archetype, description) of all loaded starter decks.
func HandleListDecks(w http.ResponseWriter, r *http.Request) {
	all := cards.AllDecks()
	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })
	summaries := make([]deckSummary, len(all))
	for i, d := range all {
		summaries[i] = deckSummary{
			ID:          d.ID,
			Name:        d.Name,
			Archetype:   d.Archetype,
			Description: d.Description,
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"decks": summaries})
}

// HandleGetDeck handles GET /api/v1/decks/{deckId}
// Returns the deck metadata and its card list (cardId + qty pairs).
func HandleGetDeck(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckId")
	def, err := cards.GetDeck(deckID)
	if err != nil {
		http.Error(w, "deck not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, def)
}

// parsePagination extracts and validates page/pageSize query params.
func parsePagination(r *http.Request) (page, pageSize int) {
	const defaultPageSize = 20
	const maxPageSize = 100

	page = 1
	pageSize = defaultPageSize

	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := r.URL.Query().Get("pageSize"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			if n > maxPageSize {
				n = maxPageSize
			}
			pageSize = n
		}
	}
	return
}

// writeJSON serialises v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
