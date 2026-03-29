package game

// CombatResult holds the computed outcome of a single attacker vs. defenders engagement.
// All values are computed before any HP mutations are applied (simultaneous resolution).
type CombatResult struct {
	OutgoingDamage int
	ReturnDamage   int
	Overflow       int
}

// ResolveCombat computes damage for one attacker against zero or more defenders.
// Implements the simultaneous damage formula from the core rules spec.
// HP mutations are NOT applied here — the caller applies them after this returns.
func ResolveCombat(attacker *ConquerorInstance, defenders []*ConquerorInstance) CombatResult {
	totalDefDEF := 0
	totalDefATK := 0
	totalDefHP := 0

	for _, d := range defenders {
		totalDefDEF += d.CurrentDEF
		totalDefATK += d.CurrentATK
		totalDefHP += d.CurrentHP
	}

	// Rend bypasses all defender DEF — damage goes straight to HP.
	outgoing := attacker.CurrentATK - totalDefDEF
	if attacker.HasKeyword(KeywordRend) {
		outgoing = attacker.CurrentATK
	}
	if outgoing < 0 {
		outgoing = 0
	}

	returning := totalDefATK - attacker.CurrentDEF
	if returning < 0 {
		returning = 0
	}

	overflow := outgoing - totalDefHP
	if overflow < 0 {
		overflow = 0
	}

	return CombatResult{
		OutgoingDamage: outgoing,
		ReturnDamage:   returning,
		Overflow:       overflow,
	}
}

// ApplyCombatDamage mutates HP on the attacker and defenders based on a resolved CombatResult.
// outgoingDistribution maps defender InstanceID → damage assigned to that defender.
// The caller is responsible for ensuring the distribution sums to <= OutgoingDamage.
// Returns a slice of destroyed conqueror InstanceIDs.
func ApplyCombatDamage(
	attacker *ConquerorInstance,
	defenders []*ConquerorInstance,
	result CombatResult,
	outgoingDistribution map[string]int,
) []string {
	// Apply return damage to attacker.
	attacker.CurrentHP -= result.ReturnDamage

	// Apply outgoing damage to each defender per the distribution.
	for _, d := range defenders {
		dmg, ok := outgoingDistribution[d.InstanceID]
		if !ok {
			continue
		}
		d.CurrentHP -= dmg
	}

	var destroyed []string

	// Check attacker destruction.
	if attacker.CurrentHP <= 0 {
		if attacker.HasKeyword(KeywordSteadfast) && !attacker.SteadfastUsed {
			attacker.CurrentHP = 1
			attacker.SteadfastUsed = true
		} else {
			destroyed = append(destroyed, attacker.InstanceID)
		}
	}

	// Check defender destructions.
	for _, d := range defenders {
		if d.CurrentHP <= 0 {
			if d.HasKeyword(KeywordSteadfast) && !d.SteadfastUsed {
				d.CurrentHP = 1
				d.SteadfastUsed = true
			} else {
				destroyed = append(destroyed, d.InstanceID)
			}
		}
	}

	return destroyed
}

// RemoveDestroyedConquerors removes all conquerors in the destroyed list from the board.
func RemoveDestroyedConquerors(gs *GameState, destroyed []string) {
	destroyedSet := make(map[string]bool, len(destroyed))
	for _, id := range destroyed {
		destroyedSet[id] = true
	}

	for col := 0; col < 3; col++ {
		for row := 0; row < 4; row++ {
			c := gs.Board.Grid[col][row]
			if c != nil && destroyedSet[c.InstanceID] {
				gs.Board.Grid[col][row] = nil
			}
		}
	}
}

// DamageStructure applies direct damage to a structure (undefended attack or Overwhelm overflow).
// Returns true if the structure was destroyed by this damage.
func DamageStructure(s *Structure, damage int) bool {
	s.HPCurrent -= damage
	if s.HPCurrent <= 0 {
		s.HPCurrent = 0
		s.IsDestroyed = true
		return true
	}
	return false
}
