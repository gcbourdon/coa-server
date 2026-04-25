package game

import (
	"fmt"

	"coa-server/cards"

	"github.com/google/uuid"
)

// QueuePlayCard validates available resources, removes the card from the player's hand,
// spends the AP cost, and pushes a SequenceItem onto the sequence.
// The card's effect does not execute until the item resolves via PassPriority.
func QueuePlayCard(gs *GameState, p PlayerIndex, cardInstanceID string, targetCol, targetRow int) error {
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

	def, err := cards.GetCard(ci.CardID)
	if err != nil {
		return err
	}

	// Commit resources immediately on play (before resolution).
	player.AP -= def.APCost
	player.Hand = append(player.Hand[:cardIdx], player.Hand[cardIdx+1:]...)

	item := SequenceItem{
		ID:         uuid.NewString(),
		Owner:      p,
		CardID:     ci.CardID,
		InstanceID: ci.InstanceID,
		ItemType:   SequenceItemPlayCard,
		TargetCol:  targetCol,
		TargetRow:  targetRow,
	}

	gs.Sequence = append(gs.Sequence, item)
	// The act of playing counts as the player's own pass — they've declared their intent.
	// Start at 1 so the opponent only needs to pass once to resolve, not twice.
	gs.PassCount = 1
	gs.PriorityPlayer = p.Opponent()

	return nil
}

// PassPriority records that p is yielding their priority window.
// When two consecutive passes occur without any intervening stack addition, the top
// item resolves. Priority then returns to the active player (CurrentTurn).
func PassPriority(gs *GameState, p PlayerIndex) error {
	if gs.PriorityPlayer != p {
		return &ValidationError{Reason: "NOT_YOUR_PRIORITY", Message: "it is not your priority to pass"}
	}
	if len(gs.Sequence) == 0 {
		return &ValidationError{Reason: "SEQUENCE_EMPTY", Message: "there are no cards on the sequence"}
	}

	gs.PassCount++

	if gs.PassCount >= 2 {
		if err := resolveTopSequenceItem(gs); err != nil {
			return err
		}
		gs.PassCount = 0
		if len(gs.Sequence) > 0 {
			// More items remain; active player gets priority to respond.
			gs.PriorityPlayer = gs.CurrentTurn
		} else {
			gs.PriorityPlayer = 0
		}
	} else {
		// One pass recorded; offer priority to the other player.
		gs.PriorityPlayer = p.Opponent()
	}

	return nil
}

// resolveTopSequenceItem pops the top item from the sequence and executes its effect.
func resolveTopSequenceItem(gs *GameState) error {
	top := gs.Sequence[len(gs.Sequence)-1]
	gs.Sequence = gs.Sequence[:len(gs.Sequence)-1]

	switch top.ItemType {
	case SequenceItemPlayCard:
		return resolvePlayCard(gs, top)
	default:
		return fmt.Errorf("unknown sequence item type: %s", top.ItemType)
	}
}

func resolvePlayCard(gs *GameState, item SequenceItem) error {
	def, err := cards.GetCard(item.CardID)
	if err != nil {
		return err
	}
	switch def.Type {
	case cards.CardTypeConqueror:
		return deployConqueror(gs, item, def)
	default:
		// Spell/constant/item resolution will be implemented per card type.
		return fmt.Errorf("resolution for card type %q not yet implemented", def.Type)
	}
}

func deployConqueror(gs *GameState, item SequenceItem, def *cards.CardDef) error {
	// If the target cell was occupied since the card was queued, the deploy fizzles.
	if gs.Board.Grid[item.TargetCol][item.TargetRow] != nil {
		return nil
	}

	keywords := make([]Keyword, len(def.Keywords))
	for i, kw := range def.Keywords {
		keywords[i] = Keyword(kw)
	}

	c := &ConquerorInstance{
		InstanceID: uuid.NewString(),
		CardID:     item.CardID,
		Owner:      item.Owner,
		Col:        item.TargetCol,
		Row:        item.TargetRow,
		IsWeary:    true,
		CurrentATK: def.BaseATK,
		CurrentDEF: def.BaseDEF,
		CurrentHP:  def.BaseHP,
		CurrentSPD: def.BaseSPD,
		CurrentRNG: def.BaseRNG,
		Keywords:   keywords,
		EffectID:   def.EffectID,
	}

	if c.HasKeyword(KeywordRush) {
		c.IsWeary = false
	}

	gs.Board.Grid[item.TargetCol][item.TargetRow] = c
	return nil
}
