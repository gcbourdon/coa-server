package game

// effects.go handles keyword-triggered and card effect resolution.
// Passive keywords (Taunt, Steadfast, Rend, Overwhelm) are handled inline in
// validator.go and combat.go respectively. This file handles effects that fire
// as discrete triggers: Intercept, and any future activated/triggered abilities.

// InterceptMove moves a conqueror with the Intercept keyword one position to
// join an adjacent defense during the opponent's combat step.
// This is called when the defending player spends 1 AP to activate Intercept.
func InterceptMove(gs *GameState, p PlayerIndex, conquerorID string, toCol, toRow int) error {
	if gs.Player(p).AP < 1 {
		return &ValidationError{
			Reason:  "INSUFFICIENT_AP",
			Message: "Intercept costs 1 AP.",
		}
	}

	c := findConqueror(gs, conquerorID)
	if c == nil {
		return &ValidationError{Reason: "CONQUEROR_NOT_FOUND", Message: "Conqueror not found."}
	}
	if !c.HasKeyword(KeywordIntercept) {
		return &ValidationError{Reason: "NO_INTERCEPT", Message: "This conqueror does not have Intercept."}
	}
	if !isAdjacent(c.Col, c.Row, toCol, toRow) {
		return &ValidationError{Reason: "NOT_ADJACENT", Message: "Intercept can only move one position."}
	}
	if gs.Board.Grid[toCol][toRow] != nil {
		return &ValidationError{Reason: "CELL_OCCUPIED", Message: "Target cell is occupied."}
	}

	gs.Board.Grid[c.Col][c.Row] = nil
	c.Col = toCol
	c.Row = toRow
	gs.Board.Grid[toCol][toRow] = c
	gs.Player(p).AP--

	return nil
}
