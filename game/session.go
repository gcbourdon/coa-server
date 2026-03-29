package game

import (
	"fmt"
	"math/rand/v2"

	"coa-server/cards"

	"github.com/google/uuid"
)

// NewGame initialises a fresh GameState for two players.
// deckP1/deckP2 are slices of card IDs representing each player's deck.
// structureIDsP1/structureIDsP2 are three structure card IDs in column order [col0, col1, col2].
func NewGame(
	gameID string,
	playerID1, playerID2 string,
	deckP1, deckP2 []string,
	structureIDsP1, structureIDsP2 [3]string,
) (*GameState, error) {
	// Coin flip: randomly decide who goes first.
	firstPlayer := PlayerIndex(rand.IntN(2) + 1) // 1 or 2

	gs := &GameState{
		GameID:      gameID,
		CurrentTurn: firstPlayer,
		FirstPlayer: firstPlayer,
		TurnNumber:  1,
		Phase:       PhaseMain,
		Winner:      0,
		Sequence:    []SequenceItem{},
	}

	p1, err := buildPlayer(playerID1, Player1, deckP1, structureIDsP1[:])
	if err != nil {
		return nil, fmt.Errorf("building player 1: %w", err)
	}
	p2, err := buildPlayer(playerID2, Player2, deckP2, structureIDsP2[:])
	if err != nil {
		return nil, fmt.Errorf("building player 2: %w", err)
	}

	gs.Players[Player1-1] = p1
	gs.Players[Player2-1] = p2

	// Shuffle both decks and deal 5-card opening hands.
	ShuffleDeck(gs.Player(Player1))
	ShuffleDeck(gs.Player(Player2))
	DrawCards(gs.Player(Player1), 5)
	DrawCards(gs.Player(Player2), 5)

	// The first player starts with 3 AP; the other gains AP on their first turn.
	gs.Player(firstPlayer).AP = GainPerTurn

	return gs, nil
}

func buildPlayer(playerID string, p PlayerIndex, deckCardIDs []string, structureIDs []string) (Player, error) {
	player := Player{
		ID:         playerID,
		AP:         0,
		Hand:       []CardInstance{},
		Deck:       make([]CardInstance, 0, len(deckCardIDs)),
		Discard:    []CardInstance{},
		Permanents: []PermanentInstance{},
	}

	for _, cardID := range deckCardIDs {
		player.Deck = append(player.Deck, CardInstance{
			InstanceID: uuid.NewString(),
			CardID:     cardID,
		})
	}

	for col, structID := range structureIDs {
		def, err := cards.GetCard(structID)
		if err != nil {
			return Player{}, fmt.Errorf("unknown structure %s: %w", structID, err)
		}
		if def.Type != cards.CardTypeStructure {
			return Player{}, fmt.Errorf("card %s is not a structure", structID)
		}
		player.Structures[col] = Structure{
			CardID:    structID,
			Owner:     p,
			Col:       col,
			HPMax:     def.BaseHP,
			HPCurrent: def.BaseHP,
			EffectID:  def.EffectID,
		}
	}

	return player, nil
}
