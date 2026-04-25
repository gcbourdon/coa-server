package game

// GameStatus represents the overall lifecycle state of a game session.
type GameStatus string

const (
	// GameStatusWaiting means the session exists but the second player has not yet connected.
	// No game initialisation (coin flip, shuffle, deal) has occurred yet.
	GameStatusWaiting    GameStatus = "waiting"
	GameStatusInProgress GameStatus = "in_progress"
	// GameStatusFinished means the game has ended and is pending persistence to the database.
	GameStatusFinished GameStatus = "finished"
)

// Phase represents which phase of a turn the game is currently in.
type Phase string

const (
	PhaseReady  Phase = "ready"
	PhaseMain   Phase = "main"
	PhaseCombat Phase = "combat"
	PhaseEnd    Phase = "end"
)

// SequenceItemType identifies what kind of action is sitting on the sequence.
type SequenceItemType string

const (
	SequenceItemPlayCard SequenceItemType = "PLAY_CARD"
)

// SequenceItem represents one entry on the sequence (stack).
// Resources (AP, hand card) are committed when an item is pushed; the effect
// only executes when the item resolves after both players pass priority.
type SequenceItem struct {
	ID         string           `json:"id"`
	Owner      PlayerIndex      `json:"owner"`
	CardID     string           `json:"cardId"`
	InstanceID string           `json:"instanceId"` // original hand CardInstance.InstanceID
	ItemType   SequenceItemType `json:"itemType"`
	// Conqueror deploy target (used when ItemType == SequenceItemPlayCard for a conqueror).
	TargetCol int `json:"targetCol"`
	TargetRow int `json:"targetRow"`
}

// GameState is the complete authoritative state of a single game session.
// This is serialised and broadcast to both clients after every action.
type GameState struct {
	GameID      string      `json:"gameId"`
	Status      GameStatus  `json:"status"`
	Players     [2]Player   `json:"players"` // index 0 = Player1, index 1 = Player2
	Board       Board       `json:"board"`
	CurrentTurn PlayerIndex `json:"currentTurn"` // 1 or 2
	FirstPlayer PlayerIndex `json:"firstPlayer"` // who won the coin flip; set once at game start
	TurnNumber  int         `json:"turnNumber"`
	Phase       Phase       `json:"phase"`
	Winner      PlayerIndex `json:"winner"` // 0 = no winner, 1 or 2 = winner

	// Sequence (stack) — FILO resolution.
	// Items are appended when played and resolved from the end (top of stack).
	Sequence       []SequenceItem `json:"sequence"`
	PriorityPlayer PlayerIndex    `json:"priorityPlayer"` // who currently holds priority; 0 = nobody (between turns)
	PassCount      int            `json:"passCount"`      // consecutive passes since the last stack addition; 2 → resolve top
}

// Player returns a pointer to the Player struct for the given PlayerIndex.
func (gs *GameState) Player(p PlayerIndex) *Player {
	return &gs.Players[p-1]
}

// Player holds all per-player state.
type Player struct {
	ID         string              `json:"id"`
	AP         int                 `json:"ap"` // 0–6
	Hand       []CardInstance      `json:"hand"`
	Deck       []CardInstance      `json:"deck"`
	Discard    []CardInstance      `json:"discard"`
	Structures [3]Structure        `json:"structures"` // index = column (0, 1, 2)
	Permanents []PermanentInstance `json:"permanents"`
	Items      []ItemInstance      `json:"items"` // equipment items in play (attached or unattached)
}

// Board holds the 3×4 grid. Grid[col][row], nil = empty cell.
// col: 0–2, row: 0–3 bottom-to-top:
//
//	row 0 = Player1 base (deploy zone, bottom)
//	row 1 = Player1 structure row
//	row 2 = Player2 structure row
//	row 3 = Player2 base (deploy zone, top)
type Board struct {
	Grid [3][4]*ConquerorInstance `json:"grid"`
}

// Structure is a player's structure occupying one column of their structure row.
type Structure struct {
	CardID      string      `json:"cardId"`
	Owner       PlayerIndex `json:"owner"`
	Col         int         `json:"col"`
	HPMax       int         `json:"hpMax"`
	HPCurrent   int         `json:"hpCurrent"`
	IsDestroyed bool        `json:"isDestroyed"`
	EffectID    string      `json:"effectId,omitempty"`
}

// ConquerorInstance is a deployed conqueror on the board.
// Current* fields reflect stat modifications from effects; they start equal to base values.
type ConquerorInstance struct {
	InstanceID    string      `json:"instanceId"`
	CardID        string      `json:"cardId"`
	Owner         PlayerIndex `json:"owner"`
	Col           int         `json:"col"`
	Row           int         `json:"row"`
	IsWeary       bool        `json:"isWeary"`
	MovesUsed     int         `json:"movesUsed"` // resets each turn; capped at CurrentSPD
	CurrentATK    int         `json:"currentAtk"`
	CurrentDEF    int         `json:"currentDef"`
	CurrentHP     int         `json:"currentHp"`
	CurrentSPD    int         `json:"currentSpd"`
	CurrentRNG    int         `json:"currentRng"`
	Keywords      []Keyword   `json:"keywords"`
	EffectID      string      `json:"effectId,omitempty"`
	SteadfastUsed bool        `json:"steadfastUsed"`
}

// CardInstance is a card in a player's hand, deck, or discard pile.
type CardInstance struct {
	InstanceID string `json:"instanceId"`
	CardID     string `json:"cardId"`
}

// PermanentInstance is a permanent effect card in play.
type PermanentInstance struct {
	InstanceID string      `json:"instanceId"`
	CardID     string      `json:"cardId"`
	Owner      PlayerIndex `json:"owner"`
	EffectID   string      `json:"effectId"`
}

// ItemInstance is an item card in play.
// Consumable items are never represented here — they resolve immediately and go to discard.
// Equipment items enter play and may be attached to a conqueror.
// AttachedTo holds the InstanceID of the equipped conqueror; empty means unattached.
type ItemInstance struct {
	InstanceID string      `json:"instanceId"`
	CardID     string      `json:"cardId"`
	Owner      PlayerIndex `json:"owner"`
	EffectID   string      `json:"effectId"`
	AttachedTo string      `json:"attachedTo,omitempty"`
}

// HasKeyword reports whether a conqueror has the given keyword.
func (c *ConquerorInstance) HasKeyword(kw Keyword) bool {
	for _, k := range c.Keywords {
		if k == kw {
			return true
		}
	}
	return false
}

// StructureRow returns the board row index of a player's structure row.
// Rows are numbered 0–3 bottom-to-top: Player1 structures are in row 1; Player2's in row 2.
func StructureRow(p PlayerIndex) int {
	if p == Player1 {
		return 1
	}
	return 2
}

// Base returns the board row index of a player's base (deploy zone).
// Rows are numbered 0–3 bottom-to-top: Player1's base is row 0; Player2's base is row 3.
func Base(p PlayerIndex) int {
	if p == Player1 {
		return 0
	}
	return 3
}
