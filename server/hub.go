package server

import (
	"coa-server/game"
	"coa-server/shared"
	"sync"
)

// Hub manages all active game sessions and their connected clients.
type Hub struct {
	mu       sync.RWMutex
	sessions map[string]*Session // gameID → session
}

// PendingCombat holds an unresolved combat step waiting for defender assignments.
type PendingCombat struct {
	Attackers []shared.AttackDeclaration
}

// Session holds the live game state and the two connected clients for one game.
type Session struct {
	mu            sync.Mutex
	State         *game.GameState
	Clients       [2]*Client      // index 0 = Player1, index 1 = Player2
	PendingCombat *PendingCombat  // non-nil when waiting for defenders
}

func NewHub() *Hub {
	return &Hub{
		sessions: make(map[string]*Session),
	}
}

// GetSession returns the session for a game ID, or nil if not found.
func (h *Hub) GetSession(gameID string) *Session {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.sessions[gameID]
}

// CreateSession creates a new game session and registers it.
func (h *Hub) CreateSession(gs *game.GameState) *Session {
	s := &Session{State: gs}
	h.mu.Lock()
	h.sessions[gs.GameID] = s
	h.mu.Unlock()
	return s
}

// RemoveSession removes a game session from the hub.
func (h *Hub) RemoveSession(gameID string) {
	h.mu.Lock()
	delete(h.sessions, gameID)
	h.mu.Unlock()
}

// Broadcast sends a message to both clients in a session.
func (s *Session) Broadcast(msg []byte) {
	for _, c := range s.Clients {
		if c != nil {
			c.Send(msg)
		}
	}
}

// SendTo sends a message to one specific player in the session.
func (s *Session) SendTo(p game.PlayerIndex, msg []byte) {
	if c := s.Clients[p-1]; c != nil {
		c.Send(msg)
	}
}
