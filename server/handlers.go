package server

import (
	"encoding/json"
	"log"
	"net/http"

	"coa-server/game"
	"coa-server/shared"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// TODO: restrict to known origins in production.
		return true
	},
}

// HandleWebSocket upgrades an HTTP connection to WebSocket and begins the read loop.
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	gameID := r.PathValue("gameId")
	if gameID == "" {
		http.Error(w, "missing gameId", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}

	h.readLoop(conn, gameID)
}

// readLoop processes incoming messages from a WebSocket connection.
func (h *Hub) readLoop(conn *websocket.Conn, gameID string) {
	var client *Client

	defer func() {
		conn.Close()
		if client != nil {
			client.Close()
		}
	}()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("unexpected close: %v", err)
			}
			return
		}

		var msg shared.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("malformed message: %v", err)
			continue
		}

		if msg.Type == shared.ActionJoinGame {
			client = h.handleJoin(conn, gameID, msg.Payload)
			continue
		}

		if client == nil {
			sendRejection(conn, "NOT_JOINED", "Send JOIN_GAME first.")
			continue
		}

		session := h.GetSession(gameID)
		if session == nil {
			sendRejection(conn, "SESSION_NOT_FOUND", "Game session not found.")
			continue
		}

		h.dispatch(session, client, msg)
	}
}

// handleJoin processes a JOIN_GAME message and registers the client in the session.
func (h *Hub) handleJoin(conn *websocket.Conn, gameID string, payload json.RawMessage) *Client {
	var p shared.JoinGamePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		sendRejection(conn, "INVALID_PAYLOAD", "Invalid JOIN_GAME payload.")
		return nil
	}

	session := h.GetSession(gameID)
	if session == nil {
		sendRejection(conn, "SESSION_NOT_FOUND", "Game session not found.")
		return nil
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	// Determine PlayerIndex (1 or 2) from player ID.
	var playerIndex game.PlayerIndex
	for _, pi := range []game.PlayerIndex{game.Player1, game.Player2} {
		if session.State.Player(pi).ID == p.PlayerID {
			playerIndex = pi
			break
		}
	}
	if playerIndex == 0 {
		sendRejection(conn, "PLAYER_NOT_IN_GAME", "Player ID not found in this game.")
		return nil
	}

	client := NewClient(conn, playerIndex, p.PlayerID)
	session.Clients[playerIndex-1] = client
	go client.WritePump()

	// Both players are now connected — run the coin flip, shuffle, and opening deal.
	if session.Clients[0] != nil && session.Clients[1] != nil {
		game.StartGame(session.State)
	}

	broadcastState(session)
	return client
}

// dispatch routes a validated action message to the appropriate handler.
func (h *Hub) dispatch(session *Session, client *Client, msg shared.Message) {
	session.mu.Lock()
	defer session.mu.Unlock()

	gs := session.State
	pi := client.PlayerIndex

	switch msg.Type {
	case shared.ActionEndTurn:
		if gs.CurrentTurn != pi {
			session.SendTo(pi, rejectMsg("NOT_YOUR_TURN", "It is not your turn."))
			return
		}
		if len(gs.Sequence) > 0 {
			session.SendTo(pi, rejectMsg("SEQUENCE_NOT_EMPTY", "Resolve all cards on the sequence before ending your turn."))
			return
		}
		game.EndTurn(gs)
		broadcastState(session)

	case shared.ActionPassPriority:
		if err := game.PassPriority(gs, pi); err != nil {
			session.SendTo(pi, rejectFromErr(err))
			return
		}
		checkAndBroadcast(session)

	case shared.ActionPlayCard:
		var p shared.PlayCardPayload
		if err := json.Unmarshal(msg.Payload, &p); err != nil {
			session.SendTo(pi, rejectMsg("INVALID_PAYLOAD", "Invalid PLAY_CARD payload."))
			return
		}
		if err := game.ValidatePlayCard(gs, pi, p.CardID, p.TargetCol, p.TargetRow); err != nil {
			session.SendTo(pi, rejectFromErr(err))
			return
		}
		if err := game.QueuePlayCard(gs, pi, p.CardID, p.TargetCol, p.TargetRow); err != nil {
			session.SendTo(pi, rejectMsg("INTERNAL_ERROR", err.Error()))
			return
		}
		broadcastState(session)

	case shared.ActionMoveConqueror:
		var p shared.MoveConquerorPayload
		if err := json.Unmarshal(msg.Payload, &p); err != nil {
			session.SendTo(pi, rejectMsg("INVALID_PAYLOAD", "Invalid MOVE_CONQUEROR payload."))
			return
		}
		if err := game.ValidateMoveConqueror(gs, pi, p.ConquerorID, p.ToCol, p.ToRow); err != nil {
			session.SendTo(pi, rejectFromErr(err))
			return
		}
		game.MoveConqueror(gs, pi, p.ConquerorID, p.ToCol, p.ToRow)
		broadcastState(session)

	case shared.ActionInitiateCombat:
		var p shared.InitiateCombatPayload
		if err := json.Unmarshal(msg.Payload, &p); err != nil {
			session.SendTo(pi, rejectMsg("INVALID_PAYLOAD", "Invalid INITIATE_COMBAT payload."))
			return
		}
		if err := game.ValidateInitiateCombat(gs, pi); err != nil {
			session.SendTo(pi, rejectFromErr(err))
			return
		}
		for _, atk := range p.Attackers {
			if err := game.ValidateAttackDeclaration(gs, pi, atk.ConquerorID, atk.TargetCol, atk.TargetRow); err != nil {
				session.SendTo(pi, rejectFromErr(err))
				return
			}
		}
		if !game.SpendAP(gs, 1) {
			session.SendTo(pi, rejectMsg("INSUFFICIENT_AP", "Not enough AP to initiate combat."))
			return
		}

		session.PendingCombat = &PendingCombat{Attackers: p.Attackers}
		session.State.Phase = game.PhaseCombat

		session.SendTo(pi.Opponent(), mustMarshal(shared.Message{
			Type: shared.EventWaitingForDefenders,
			Payload: mustMarshalRaw(shared.WaitingForDefendersPayload{
				Attackers: p.Attackers,
				TimeoutMs: 15000,
			}),
		}))
		broadcastState(session)

	case shared.ActionAssignDefenders:
		var p shared.AssignDefendersPayload
		if err := json.Unmarshal(msg.Payload, &p); err != nil {
			session.SendTo(pi, rejectMsg("INVALID_PAYLOAD", "Invalid ASSIGN_DEFENDERS payload."))
			return
		}
		if session.PendingCombat == nil {
			session.SendTo(pi, rejectMsg("NO_PENDING_COMBAT", "No combat is waiting for defenders."))
			return
		}
		// Defender assignments come from the defending player (opponent of the attacker).
		attacker := gs.CurrentTurn // combat was initiated on CurrentTurn's action, but EndTurn hasn't been called
		_ = attacker

		// Resolve each attacker independently.
		pending := session.PendingCombat
		session.PendingCombat = nil

		for _, atk := range pending.Attackers {
			// Collect defenders assigned to this attacker.
			var defenderIDs []string
			for _, def := range p.Defenders {
				if def.DefendsAgainst == atk.ConquerorID {
					defenderIDs = append(defenderIDs, def.ConquerorID)
				}
			}

			// Validate defenders.
			if err := game.ValidateAssignDefenders(gs, pi, defenderIDs, atk.ConquerorID); err != nil {
				session.SendTo(pi, rejectFromErr(err))
				return
			}

			// Build equal distribution of outgoing damage across defenders (attacker chooses — default: equal split).
			// TODO: let the attacker specify damage distribution for multi-defender combats.
			distribution := map[string]int{}
			if len(defenderIDs) > 0 {
				atker := game.FindConquerorByID(gs, atk.ConquerorID)
				defenders := game.FindConquerorsByIDs(gs, defenderIDs)
				result := game.ResolveCombat(atker, defenders)
				distribute(distribution, defenders, result.OutgoingDamage)
			}

			destroyed, err := game.ExecuteCombat(gs, pi.Opponent(), atk.ConquerorID, atk.TargetCol, atk.TargetRow, defenderIDs, distribution, "")
			if err != nil {
				log.Printf("combat error: %v", err)
				continue
			}

			session.Broadcast(mustMarshal(shared.Message{
				Type: shared.EventCombatResult,
				Payload: mustMarshalRaw(shared.CombatResultPayload{
					AttackerID:     atk.ConquerorID,
					Defenders:      defenderIDs,
					Destroyed:      destroyed,
				}),
			}))
		}

		// Combat is fully resolved — return to Main phase so the attacker can continue their turn.
		session.State.Phase = game.PhaseMain
		checkAndBroadcast(session)

	default:
		session.SendTo(pi, rejectMsg("UNKNOWN_ACTION", "Unknown action type: "+msg.Type))
	}
}

// distribute spreads outgoing damage across defenders proportional to their HP.
// This is a simple greedy allocation — fill each defender to death, leftover is overflow.
func distribute(dist map[string]int, defenders []*game.ConquerorInstance, outgoing int) {
	remaining := outgoing
	for _, d := range defenders {
		if remaining <= 0 {
			break
		}
		dmg := d.CurrentHP
		if dmg > remaining {
			dmg = remaining
		}
		dist[d.InstanceID] = dmg
		remaining -= dmg
	}
}

// --- helpers ---

func broadcastState(session *Session) {
	session.Broadcast(mustMarshal(shared.Message{
		Type:    shared.EventGameState,
		Payload: mustMarshalRaw(session.State),
	}))
}

func checkAndBroadcast(session *Session) {
	winner := game.CheckWinCondition(session.State)
	if winner != 0 {
		session.State.Winner = winner
		session.State.Status = game.GameStatusFinished
		session.Broadcast(mustMarshal(shared.Message{
			Type: shared.EventGameOver,
			Payload: mustMarshalRaw(shared.GameOverPayload{
				Winner: int(winner),
				Reason: "ALL_STRUCTURES_DESTROYED",
			}),
		}))
		return
	}
	broadcastState(session)
}

func sendRejection(conn *websocket.Conn, reason, message string) {
	_ = conn.WriteMessage(websocket.TextMessage, mustMarshal(shared.Message{
		Type:    shared.EventActionRejected,
		Payload: mustMarshalRaw(shared.ActionRejectedPayload{Reason: reason, Message: message}),
	}))
}

func rejectMsg(reason, message string) []byte {
	return mustMarshal(shared.Message{
		Type:    shared.EventActionRejected,
		Payload: mustMarshalRaw(shared.ActionRejectedPayload{Reason: reason, Message: message}),
	})
}

func rejectFromErr(err error) []byte {
	if ve, ok := err.(*game.ValidationError); ok {
		return rejectMsg(ve.Reason, ve.Message)
	}
	return rejectMsg("VALIDATION_ERROR", err.Error())
}

func mustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func mustMarshalRaw(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
