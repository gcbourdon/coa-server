package game

// Phase represents which phase of a turn the game is currently in.
type Phase string

const (
	PhaseReady Phase = "ready"
	PhaseMain  Phase = "main"
	PhaseEnd   Phase = "end"
)

// GameState is the complete authoritative state of a single game session.
// This is serialised and broadcast to both clients after every action.
type GameState struct {
	GameID      string      `json:"gameId"`
	Players     [2]Player   `json:"players"` // index 0 = Player1, index 1 = Player2
	Board       Board       `json:"board"`
	CurrentTurn PlayerIndex `json:"currentTurn"` // 1 or 2
	TurnNumber  int         `json:"turnNumber"`
	Phase       Phase       `json:"phase"`
	Winner      PlayerIndex `json:"winner"` // 0 = no winner, 1 or 2 = winner
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
}

// Board holds the 3×4 grid. Grid[col][row], nil = empty cell.
// col: 0–2, row: 0–3 (row 0 = Player2 structure row, row 3 = Player1 structure row)
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
// Player1's structures are in row 3; Player2's structures are in row 0.
func StructureRow(p PlayerIndex) int {
	if p == Player1 {
		return 3
	}
	return 0
}

// Base returns the board row index of a player's base (deploy zone).
// Player1's base is row 2; Player2's base is row 1.
func Base(p PlayerIndex) int {
	if p == Player1 {
		return 2
	}
	return 1
}
