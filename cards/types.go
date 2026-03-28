package cards

import "encoding/json"

// CardType represents the type of a card.
type CardType string

const (
	CardTypeConqueror CardType = "conqueror"
	CardTypeSpell     CardType = "spell"
	CardTypeConstant  CardType = "constant"
	CardTypeStructure CardType = "structure"
)

// CardDef is the static definition of a card loaded from the cards/ data directory.
// All card types share this struct — fields not relevant to a type are omitted/zero.
type CardDef struct {
	ID   string   `json:"id"`
	Name string   `json:"name"`
	Type CardType `json:"type"`

	// AP cost to play. Structures have no AP cost (placed at game start).
	APCost int `json:"ap_cost,omitempty"`

	// Conqueror & structure stats (structures use BaseHP only).
	BaseATK int      `json:"base_atk,omitempty"`
	BaseDEF int      `json:"base_def,omitempty"`
	BaseHP  int      `json:"base_hp,omitempty"`
	BaseSPD int      `json:"base_spd,omitempty"`
	BaseRNG int      `json:"base_rng,omitempty"`
	Keywords []string `json:"keywords,omitempty"`

	// Spell tag — if true, can be played during the opponent's turn.
	Immediate bool `json:"immediate,omitempty"`

	// Effect definition. EffectID names the handler; EffectParams supplies its inputs.
	// Used by conqueror abilities, spells, constants, and structure abilities.
	EffectID     string          `json:"effect_id,omitempty"`
	EffectParams json.RawMessage `json:"effect_params,omitempty"`

	// Card text.
	RulesText  string `json:"rules_text,omitempty"`
	FlavorText string `json:"flavor_text,omitempty"`
}
