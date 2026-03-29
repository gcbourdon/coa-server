package cards

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

var cardRegistry map[string]*CardDef
var deckRegistry map[string]*DeckDef

// LoadAll walks the cards data root and loads all card definitions.
// dataRoot should point to the top-level cards/ directory (e.g. "../cards").
// Subfolders: conquerors/, spells/, constants/, structures/, items/
func LoadAll(dataRoot string) error {
	cardRegistry = make(map[string]*CardDef)

	folders := []string{"conquerors", "spells", "constants", "structures", "items"}
	for _, folder := range folders {
		if err := loadFolder(filepath.Join(dataRoot, folder)); err != nil {
			return fmt.Errorf("loading %s: %w", folder, err)
		}
	}

	return nil
}

func loadFolder(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // empty or missing folder is fine
		}
		return err
	}

	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return err
		}
		var def CardDef
		if err := json.Unmarshal(data, &def); err != nil {
			return fmt.Errorf("parsing %s: %w", e.Name(), err)
		}
		if def.SetID == "" || def.Number == 0 {
			return fmt.Errorf("%s: card definition missing required fields 'set_id' or 'number'", e.Name())
		}
		def.ID = def.SetID + "_" + strconv.Itoa(def.Number)
		if _, exists := cardRegistry[def.ID]; exists {
			return fmt.Errorf("duplicate card id %s (from %s)", def.ID, e.Name())
		}
		cardRegistry[def.ID] = &def
	}
	return nil
}

// GetCard returns the CardDef for a given card ID.
func GetCard(id string) (*CardDef, error) {
	def, ok := cardRegistry[id]
	if !ok {
		return nil, fmt.Errorf("unknown card id: %s", id)
	}
	return def, nil
}

// AllCards returns all loaded card definitions.
func AllCards() []*CardDef {
	out := make([]*CardDef, 0, len(cardRegistry))
	for _, d := range cardRegistry {
		out = append(out, d)
	}
	return out
}

// AllByType returns all loaded card definitions of a given type.
func AllByType(t CardType) []*CardDef {
	var out []*CardDef
	for _, d := range cardRegistry {
		if d.Type == t {
			out = append(out, d)
		}
	}
	return out
}

// LoadDecks walks the cards/decks/ directory and loads all deck definitions.
func LoadDecks(dataRoot string) error {
	deckRegistry = make(map[string]*DeckDef)

	entries, err := os.ReadDir(filepath.Join(dataRoot, "decks"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dataRoot, "decks", e.Name()))
		if err != nil {
			return err
		}
		var def DeckDef
		if err := json.Unmarshal(data, &def); err != nil {
			return fmt.Errorf("parsing deck %s: %w", e.Name(), err)
		}
		if def.ID == "" {
			return fmt.Errorf("%s: deck definition missing required field 'id'", e.Name())
		}
		deckRegistry[def.ID] = &def
	}
	return nil
}

// GetDeck returns the DeckDef for a given deck ID.
func GetDeck(id string) (*DeckDef, error) {
	def, ok := deckRegistry[id]
	if !ok {
		return nil, fmt.Errorf("unknown deck id: %s", id)
	}
	return def, nil
}

// AllDecks returns all loaded deck definitions.
func AllDecks() []*DeckDef {
	out := make([]*DeckDef, 0, len(deckRegistry))
	for _, d := range deckRegistry {
		out = append(out, d)
	}
	return out
}
