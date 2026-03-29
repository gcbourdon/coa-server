package cards

import "encoding/json"

// CardType represents the type of a card.
type CardType string

const (
	CardTypeConqueror CardType = "conqueror"
	CardTypeSpell     CardType = "spell"
	CardTypeConstant  CardType = "constant"
	CardTypeStructure CardType = "structure"
	CardTypeItem      CardType = "item"
)

// CardSubtype further classifies item cards.
type CardSubtype string

const (
	CardSubtypeConsumable CardSubtype = "consumable"
	CardSubtypeEquipment  CardSubtype = "equipment"
)

// CardRarity represents the rarity of a card.
type CardRarity string

const (
	CardRarityCommon    CardRarity = "common"
	CardRarityUncommon  CardRarity = "uncommon"
	CardRarityRare      CardRarity = "rare"
	CardRarityUltraRare CardRarity = "ultra_rare"
)

// DeckDef is the static definition of a starter deck loaded from the cards/decks/ directory.
type DeckDef struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Archetype   string          `json:"archetype"`
	Description string          `json:"description"`
	Cards       []DeckCardEntry `json:"cards"`
	Structures  [3]string       `json:"structures"` // three structure card IDs in default column order
}

// DeckCardEntry is a single card entry in a deck definition.
type DeckCardEntry struct {
	CardID string `json:"cardId"`
	Qty    int    `json:"qty"`
}

// CardDef is the static definition of a card loaded from the cards/ data directory.
// All card types share this struct — fields not relevant to a type are omitted/zero.
//
// The unique ID is derived at load time as "<SetID>_<Number>" (e.g. "OGN_1").
// It is not stored in the JSON file.
type CardDef struct {
	ID     string   `json:"-"` // derived: SetID + "_" + strconv.Itoa(Number)
	SetID  string   `json:"set_id"`
	Number int      `json:"number"`
	Name   string   `json:"name"`
	Type   CardType   `json:"type"`
	Rarity CardRarity `json:"rarity"`

	// AP cost to play. Structures have no AP cost (placed at game start).
	APCost int `json:"ap_cost,omitempty"`

	// Conqueror & structure stats (structures use BaseHP only).
	BaseATK  int      `json:"base_atk,omitempty"`
	BaseDEF  int      `json:"base_def,omitempty"`
	BaseHP   int      `json:"base_hp,omitempty"`
	BaseSPD  int      `json:"base_spd,omitempty"`
	BaseRNG  int      `json:"base_rng,omitempty"`
	Keywords []string `json:"keywords,omitempty"`

	// Item subtype — "consumable" or "equipment". Only set for CardTypeItem.
	Subtype CardSubtype `json:"subtype,omitempty"`

	// ArmCost is the AP cost to attach an equipment item to a conqueror.
	// Only set for equipment items (Subtype == CardSubtypeEquipment).
	ArmCost int `json:"arm_cost,omitempty"`

	// Immediate keyword — if true, this card can be added to the sequence whenever
	// its controller has priority, including during the opponent's turn.
	// Applies to any card type, not just spells.
	Immediate bool `json:"immediate,omitempty"`

	// Effect definition. EffectID names the handler; EffectParams supplies its inputs.
	// Used by conqueror abilities, spells, constants, and structure abilities.
	EffectID     string          `json:"effect_id,omitempty"`
	EffectParams json.RawMessage `json:"effect_params,omitempty"`

	// Card text.
	RulesText  string `json:"rules_text,omitempty"`
	FlavorText string `json:"flavor_text,omitempty"`
}
