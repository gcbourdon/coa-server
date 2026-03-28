package game

const (
	MaxAP       = 6
	GainPerTurn = 3
)

// StartReadyPhase transitions to the Main phase for the current player:
// clears Weary on all their conquerors, gains AP, triggers start-of-turn effects.
func StartReadyPhase(gs *GameState) {
	p := gs.CurrentTurn

	// Clear Weary and reset per-turn flags on all conquerors owned by the active player.
	for col := 0; col < 3; col++ {
		for row := 0; row < 4; row++ {
			c := gs.Board.Grid[col][row]
			if c != nil && c.Owner == p {
				c.IsWeary = false
				c.SteadfastUsed = false
			}
		}
	}

	// Gain AP, capped at MaxAP.
	player := gs.Player(p)
	player.AP += GainPerTurn
	if player.AP > MaxAP {
		player.AP = MaxAP
	}

	gs.Phase = PhaseMain
}

// EndTurn advances the game to the next player's turn.
func EndTurn(gs *GameState) {
	gs.Phase = PhaseEnd
	// TODO: resolve end-of-turn permanent/structure effects here.

	gs.CurrentTurn = gs.CurrentTurn.Opponent()
	gs.TurnNumber++
	StartReadyPhase(gs)
}

// SpendAP deducts cost from the active player's AP.
// Returns false if the player cannot afford it.
func SpendAP(gs *GameState, cost int) bool {
	p := gs.Player(gs.CurrentTurn)
	if p.AP < cost {
		return false
	}
	p.AP -= cost
	return true
}

// CheckWinCondition returns the winning PlayerIndex if the game is over,
// or 0 if play continues.
// Win condition: all three opponent structures destroyed.
func CheckWinCondition(gs *GameState) PlayerIndex {
	for _, p := range []PlayerIndex{Player1, Player2} {
		opponent := p.Opponent()
		allDestroyed := true
		for i := 0; i < 3; i++ {
			if !gs.Player(opponent).Structures[i].IsDestroyed {
				allDestroyed = false
				break
			}
		}
		if allDestroyed {
			return p
		}
	}
	return 0
}
