package game

import "fmt"

// ValidationError is returned when an action fails validation.
type ValidationError struct {
	Reason  string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Reason, e.Message)
}

func validationErr(reason, message string) *ValidationError {
	return &ValidationError{Reason: reason, Message: message}
}

// ValidatePlayCard checks that a player can legally play a conqueror card onto the board.
func ValidatePlayCard(gs *GameState, p PlayerIndex, cardInstanceID string, targetCol, targetRow int) error {
	if gs.CurrentTurn != p {
		return validationErr("NOT_YOUR_TURN", "It is not your turn.")
	}
	if gs.Phase != PhaseMain {
		return validationErr("WRONG_PHASE", "Cards can only be played during the Main phase.")
	}

	cardIdx := -1
	for i, c := range gs.Player(p).Hand {
		if c.InstanceID == cardInstanceID {
			cardIdx = i
			break
		}
	}
	if cardIdx == -1 {
		return validationErr("CARD_NOT_IN_HAND", "That card is not in your hand.")
	}

	if targetCol < 0 || targetCol > 2 || targetRow < 0 || targetRow > 3 {
		return validationErr("INVALID_POSITION", "Target position is out of bounds.")
	}
	if targetRow != Base(p) {
		return validationErr("WRONG_ROW", fmt.Sprintf("Conquerors must be deployed to your base (row %d).", Base(p)))
	}
	if gs.Board.Grid[targetCol][targetRow] != nil {
		return validationErr("CELL_OCCUPIED", "That cell is already occupied.")
	}

	return nil
}

// ValidateMoveConqueror checks that a move action is legal.
func ValidateMoveConqueror(gs *GameState, p PlayerIndex, conquerorID string, toCol, toRow int) error {
	if gs.CurrentTurn != p {
		return validationErr("NOT_YOUR_TURN", "It is not your turn.")
	}
	if gs.Phase != PhaseMain {
		return validationErr("WRONG_PHASE", "Conquerors can only be moved during the Main phase.")
	}
	if gs.Player(p).AP < 1 {
		return validationErr("INSUFFICIENT_AP", "Moving costs 1 AP.")
	}

	c := findConqueror(gs, conquerorID)
	if c == nil {
		return validationErr("CONQUEROR_NOT_FOUND", "Conqueror not found on the board.")
	}
	if c.Owner != p {
		return validationErr("NOT_YOUR_CONQUEROR", "You do not control that conqueror.")
	}
	if c.IsWeary {
		return validationErr("CONQUEROR_WEARY", "Weary conquerors cannot move.")
	}

	if toCol < 0 || toCol > 2 || toRow < 0 || toRow > 3 {
		return validationErr("INVALID_POSITION", "Target position is out of bounds.")
	}

	colDelta := abs(toCol - c.Col)
	rowDelta := abs(toRow - c.Row)
	totalDelta := colDelta + rowDelta

	if colDelta > 0 && rowDelta > 0 {
		return validationErr("DIAGONAL_MOVE", "Conquerors cannot move diagonally.")
	}
	if totalDelta == 0 {
		return validationErr("NO_MOVEMENT", "Conqueror must move at least one position.")
	}
	if totalDelta > c.CurrentSPD {
		return validationErr("EXCEEDS_SPEED", fmt.Sprintf("This conqueror can move at most %d position(s) per action.", c.CurrentSPD))
	}
	// Conquerors cannot enter the opponent's base.
	if toRow == Base(c.Owner.Opponent()) {
		return validationErr("FORBIDDEN_ROW", "Conquerors cannot move into the opponent's base.")
	}

	if gs.Board.Grid[toCol][toRow] != nil {
		return validationErr("CELL_OCCUPIED", "That cell is already occupied by another conqueror.")
	}

	return nil
}

// ValidateInitiateCombat checks that a combat step can be initiated.
func ValidateInitiateCombat(gs *GameState, p PlayerIndex) error {
	if gs.CurrentTurn != p {
		return validationErr("NOT_YOUR_TURN", "It is not your turn.")
	}
	if gs.Phase != PhaseMain {
		return validationErr("WRONG_PHASE", "Combat can only be initiated during the Main phase.")
	}
	if gs.Player(p).AP < 1 {
		return validationErr("INSUFFICIENT_AP", "Initiating combat costs 1 AP.")
	}
	return nil
}

// ValidateAttackDeclaration checks that a specific attacker targeting a specific position is legal.
func ValidateAttackDeclaration(gs *GameState, p PlayerIndex, conquerorID string, targetCol, targetRow int) error {
	c := findConqueror(gs, conquerorID)
	if c == nil {
		return validationErr("CONQUEROR_NOT_FOUND", "Attacker not found on the board.")
	}
	if c.Owner != p {
		return validationErr("NOT_YOUR_CONQUEROR", "You do not control that attacker.")
	}
	if c.IsWeary {
		return validationErr("CONQUEROR_WEARY", "Weary conquerors cannot attack.")
	}

	opponent := p.Opponent()

	if !withinRange(c, targetCol, targetRow) {
		return validationErr("OUT_OF_RANGE", "Target is not within this conqueror's attack range.")
	}

	targetIsOpponentTerritory := targetRow == Base(opponent) || targetRow == StructureRow(opponent)
	targetHasOpponentConqueror := gs.Board.Grid[targetCol][targetRow] != nil &&
		gs.Board.Grid[targetCol][targetRow].Owner == opponent
	targetIsStructureRow := targetRow == StructureRow(opponent)

	if !targetIsOpponentTerritory && !targetHasOpponentConqueror {
		return validationErr("INVALID_TARGET", "Attackers must target the opponent's side of the board.")
	}

	// RNG 0 conquerors must occupy the structure's cell to attack it.
	if targetIsStructureRow && c.CurrentRNG == 0 {
		if c.Col != targetCol || c.Row != targetRow {
			return validationErr("MUST_OCCUPY_STRUCTURE_CELL", "RNG 0 conquerors must occupy the structure's cell to attack it.")
		}
	}

	// RNG 1 conquerors attack a structure from the opponent's base, same column.
	if targetIsStructureRow && c.CurrentRNG == 1 {
		if c.Col != targetCol || c.Row != Base(opponent) {
			return validationErr("OUT_OF_RANGE", "RNG 1 conquerors must be in the opponent's base and same column to attack a structure.")
		}
	}

	return nil
}

// ValidateAssignDefenders checks that defender assignments are legal.
func ValidateAssignDefenders(gs *GameState, p PlayerIndex, defenders []string, attackerID string) error {
	for _, defID := range defenders {
		d := findConqueror(gs, defID)
		if d == nil {
			return validationErr("CONQUEROR_NOT_FOUND", fmt.Sprintf("Defender %s not found.", defID))
		}
		if d.Owner != p {
			return validationErr("NOT_YOUR_CONQUEROR", "You can only assign your own conquerors as defenders.")
		}
		// Weary conquerors CAN defend — no Weary check here.
	}

	attacker := findConqueror(gs, attackerID)
	if attacker == nil {
		return validationErr("ATTACKER_NOT_FOUND", "Attacker not found.")
	}

	defenderSet := make(map[string]bool, len(defenders))
	for _, id := range defenders {
		defenderSet[id] = true
	}

	// Taunt: any adjacent opponent conqueror with Taunt must be included.
	for col := 0; col < 3; col++ {
		for row := 0; row < 4; row++ {
			c := gs.Board.Grid[col][row]
			if c == nil || c.Owner == attacker.Owner {
				continue
			}
			if !c.HasKeyword(KeywordTaunt) {
				continue
			}
			if !isAdjacent(attacker.Col, attacker.Row, col, row) {
				continue
			}
			if !defenderSet[c.InstanceID] {
				return validationErr("MUST_INCLUDE_TAUNT", fmt.Sprintf("Conqueror %s has Taunt and must be included in this defense.", c.InstanceID))
			}
		}
	}

	return nil
}

// ValidateUseAbility checks that an ability activation is legal.
func ValidateUseAbility(gs *GameState, p PlayerIndex, conquerorID, abilityID string, cost int) error {
	if gs.CurrentTurn != p {
		return validationErr("NOT_YOUR_TURN", "It is not your turn.")
	}
	if gs.Phase != PhaseMain {
		return validationErr("WRONG_PHASE", "Abilities can only be used during the Main phase.")
	}

	c := findConqueror(gs, conquerorID)
	if c == nil {
		return validationErr("CONQUEROR_NOT_FOUND", "Conqueror not found.")
	}
	if c.Owner != p {
		return validationErr("NOT_YOUR_CONQUEROR", "You do not control that conqueror.")
	}
	if c.IsWeary {
		return validationErr("CONQUEROR_WEARY", "Weary conquerors cannot use abilities.")
	}
	if gs.Player(p).AP < cost {
		return validationErr("INSUFFICIENT_AP", fmt.Sprintf("This ability costs %d AP but you only have %d.", cost, gs.Player(p).AP))
	}

	return nil
}

// --- helpers ---

// FindConquerorByID returns the conqueror with the given instance ID, or nil.
func FindConquerorByID(gs *GameState, instanceID string) *ConquerorInstance {
	return findConqueror(gs, instanceID)
}

// FindConquerorsByIDs returns a slice of conquerors matching the given instance IDs.
func FindConquerorsByIDs(gs *GameState, ids []string) []*ConquerorInstance {
	out := make([]*ConquerorInstance, 0, len(ids))
	for _, id := range ids {
		if c := findConqueror(gs, id); c != nil {
			out = append(out, c)
		}
	}
	return out
}

func findConqueror(gs *GameState, instanceID string) *ConquerorInstance {
	for col := 0; col < 3; col++ {
		for row := 0; row < 4; row++ {
			c := gs.Board.Grid[col][row]
			if c != nil && c.InstanceID == instanceID {
				return c
			}
		}
	}
	return nil
}

func withinRange(c *ConquerorInstance, targetCol, targetRow int) bool {
	colDelta := abs(targetCol - c.Col)
	rowDelta := abs(targetRow - c.Row)

	if c.CurrentRNG == 0 {
		return (colDelta == 0 && rowDelta == 0) || (colDelta+rowDelta == 1)
	}
	return (colDelta == 0 && rowDelta <= 1) || (colDelta <= 1 && rowDelta == 0)
}

func isAdjacent(col1, row1, col2, row2 int) bool {
	return abs(col1-col2)+abs(row1-row2) == 1
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
