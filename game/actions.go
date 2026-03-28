package game

import (
	"fmt"

	"coa-server/cards"

	"github.com/google/uuid"
)

// PlayConqueror plays a conqueror card from a player's hand onto the board.
// Assumes validation has already passed.
func PlayConqueror(gs *GameState, p PlayerIndex, cardInstanceID string, targetCol, targetRow int) error {
	player := gs.Player(p)

	cardIdx := -1
	for i, c := range player.Hand {
		if c.InstanceID == cardInstanceID {
			cardIdx = i
			break
		}
	}
	if cardIdx == -1 {
		return fmt.Errorf("card not in hand: %s", cardInstanceID)
	}
	ci := player.Hand[cardIdx]
	player.Hand = append(player.Hand[:cardIdx], player.Hand[cardIdx+1:]...)

	def, err := cards.GetCard(ci.CardID)
	if err != nil {
		return err
	}

	player.AP -= def.APCost

	// Convert []string keywords from card def to []Keyword.
	keywords := make([]Keyword, len(def.Keywords))
	for i, kw := range def.Keywords {
		keywords[i] = Keyword(kw)
	}

	conqueror := &ConquerorInstance{
		InstanceID: uuid.NewString(),
		CardID:     ci.CardID,
		Owner:      p,
		Col:        targetCol,
		Row:        targetRow,
		IsWeary:    true, // all conquerors enter Weary unless Rush
		CurrentATK: def.BaseATK,
		CurrentDEF: def.BaseDEF,
		CurrentHP:  def.BaseHP,
		CurrentSPD: def.BaseSPD,
		CurrentRNG: def.BaseRNG,
		Keywords: keywords,
		EffectID: def.EffectID,
	}

	if conqueror.HasKeyword(KeywordRush) {
		conqueror.IsWeary = false
	}

	gs.Board.Grid[targetCol][targetRow] = conqueror
	return nil
}

// MoveConqueror moves a conqueror to a new position.
// Assumes validation has already passed.
func MoveConqueror(gs *GameState, p PlayerIndex, conquerorID string, toCol, toRow int) {
	c := findConqueror(gs, conquerorID)
	gs.Board.Grid[c.Col][c.Row] = nil
	c.Col = toCol
	c.Row = toRow
	gs.Board.Grid[toCol][toRow] = c
	gs.Player(p).AP--
}

// ExecuteCombat resolves a single attacker vs a position.
// defenderIDs are the defenders assigned by the defending player.
// outgoingDistribution maps defender InstanceID → damage assigned.
// overflowTargetID is the InstanceID of a conqueror to receive non-Overwhelm overflow (may be empty).
// Returns the InstanceIDs of destroyed units.
func ExecuteCombat(
	gs *GameState,
	p PlayerIndex,
	attackerID string,
	targetCol, targetRow int,
	defenderIDs []string,
	outgoingDistribution map[string]int,
	overflowTargetID string,
) ([]string, error) {
	attacker := findConqueror(gs, attackerID)
	if attacker == nil {
		return nil, fmt.Errorf("attacker not found: %s", attackerID)
	}

	defenders := make([]*ConquerorInstance, 0, len(defenderIDs))
	for _, id := range defenderIDs {
		if d := findConqueror(gs, id); d != nil {
			defenders = append(defenders, d)
		}
	}

	result := ResolveCombat(attacker, defenders)
	opponent := p.Opponent()

	var destroyed []string

	if len(defenders) == 0 {
		// Undefended attack: hits a structure or an exposed conqueror directly.
		if targetRow == StructureRow(opponent) {
			structure := &gs.Player(opponent).Structures[targetCol]
			if !structure.IsDestroyed {
				DamageStructure(structure, attacker.CurrentATK)
			}
		} else {
			target := gs.Board.Grid[targetCol][targetRow]
			if target != nil {
				target.CurrentHP -= attacker.CurrentATK
				if target.CurrentHP <= 0 {
					if target.HasKeyword(KeywordSteadfast) && !target.SteadfastUsed {
						target.CurrentHP = 1
						target.SteadfastUsed = true
					} else {
						destroyed = append(destroyed, target.InstanceID)
					}
				}
			}
		}
	} else {
		// Defended combat.
		destroyed = ApplyCombatDamage(attacker, defenders, result, outgoingDistribution)

		// Overwhelm: overflow goes to the attacked structure.
		if attacker.HasKeyword(KeywordOverwhelm) && result.Overflow > 0 && targetRow == StructureRow(opponent) {
			structure := &gs.Player(opponent).Structures[targetCol]
			if !structure.IsDestroyed {
				DamageStructure(structure, result.Overflow)
			}
		} else if result.Overflow > 0 && overflowTargetID != "" {
			// Non-Overwhelm overflow goes to another opponent conqueror — never the attacked structure.
			overflowTarget := findConqueror(gs, overflowTargetID)
			if overflowTarget != nil && overflowTarget.Owner == opponent {
				overflowTarget.CurrentHP -= result.Overflow
				if overflowTarget.CurrentHP <= 0 {
					if overflowTarget.HasKeyword(KeywordSteadfast) && !overflowTarget.SteadfastUsed {
						overflowTarget.CurrentHP = 1
						overflowTarget.SteadfastUsed = true
					} else {
						destroyed = append(destroyed, overflowTarget.InstanceID)
					}
				}
			}
		}
	}

	attacker.IsWeary = true

	RemoveDestroyedConquerors(gs, destroyed)
	return destroyed, nil
}

// HandleFuryAttack processes the second Fury attack. Costs 1 AP; attacker becomes Weary after.
func HandleFuryAttack(gs *GameState, p PlayerIndex, attackerID string) error {
	if gs.Player(p).AP < 1 {
		return fmt.Errorf("fury second attack costs 1 AP")
	}
	attacker := findConqueror(gs, attackerID)
	if attacker == nil {
		return fmt.Errorf("attacker not found: %s", attackerID)
	}
	if !attacker.HasKeyword(KeywordFury) {
		return fmt.Errorf("conqueror does not have Fury")
	}
	gs.Player(p).AP--
	return nil
}
