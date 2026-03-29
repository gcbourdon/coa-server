package shared

import "encoding/json"

// Message is the envelope for all WebSocket messages in both directions.
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Client → Server action types
const (
	ActionJoinGame        = "JOIN_GAME"
	ActionEndTurn         = "END_TURN"
	ActionPlayCard        = "PLAY_CARD"
	ActionPassPriority    = "PASS_PRIORITY"
	ActionMoveConqueror   = "MOVE_CONQUEROR"
	ActionInitiateCombat  = "INITIATE_COMBAT"
	ActionAssignDefenders = "ASSIGN_DEFENDERS"
	ActionUseAbility      = "USE_ABILITY"
	ActionPlaySpell       = "PLAY_SPELL"
)

// Server → Client event types
const (
	EventGameState          = "GAME_STATE"
	EventActionRejected     = "ACTION_REJECTED"
	EventCombatResult       = "COMBAT_RESULT"
	EventGameOver           = "GAME_OVER"
	EventWaitingForDefenders = "WAITING_FOR_DEFENDERS"
)

// --- Client → Server payloads ---

type JoinGamePayload struct {
	GameID   string `json:"gameId"`
	PlayerID string `json:"playerId"`
}

type PlayCardPayload struct {
	CardID    string `json:"cardId"`
	TargetCol int    `json:"targetCol"`
	TargetRow int    `json:"targetRow"`
}

type MoveConquerorPayload struct {
	ConquerorID string `json:"conquerorId"`
	ToCol       int    `json:"toCol"`
	ToRow       int    `json:"toRow"`
}

type AttackDeclaration struct {
	ConquerorID string `json:"conquerorId"`
	TargetCol   int    `json:"targetCol"`
	TargetRow   int    `json:"targetRow"`
}

type InitiateCombatPayload struct {
	Attackers []AttackDeclaration `json:"attackers"`
}

type DefenderAssignment struct {
	ConquerorID      string `json:"conquerorId"`
	DefendsAgainst   string `json:"defendsAgainst"`
}

type AssignDefendersPayload struct {
	Defenders []DefenderAssignment `json:"defenders"`
}

type UseAbilityPayload struct {
	ConquerorID string  `json:"conquerorId"`
	AbilityID   string  `json:"abilityId"`
	TargetID    *string `json:"targetId"`
}

type PlaySpellPayload struct {
	CardID  string   `json:"cardId"`
	Targets []string `json:"targets"`
}

// --- Server → Client payloads ---

type ActionRejectedPayload struct {
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

type OverflowResult struct {
	Damage     int    `json:"damage"`
	AssignedTo string `json:"assignedTo"`
}

type CombatResultPayload struct {
	AttackerID     string         `json:"attackerId"`
	Defenders      []string       `json:"defenders"`
	OutgoingDamage int            `json:"outgoingDamage"`
	ReturnDamage   int            `json:"returnDamage"`
	Destroyed      []string       `json:"destroyed"`
	Overflow       *OverflowResult `json:"overflow,omitempty"`
}

type GameOverPayload struct {
	Winner int    `json:"winner"` // 0 or 1
	Reason string `json:"reason"`
}

type WaitingForDefendersPayload struct {
	Attackers []AttackDeclaration `json:"attackers"`
	TimeoutMs int                 `json:"timeoutMs"`
}
