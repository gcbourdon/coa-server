package cards

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

var cardRegistry map[string]*CardDef

// LoadAll walks the cards data root and loads all card definitions.
// dataRoot should point to the top-level cards/ directory (e.g. "../cards").
// Subfolders: conquerors/, spells/, constants/, structures/
func LoadAll(dataRoot string) error {
	cardRegistry = make(map[string]*CardDef)

	folders := []string{"conquerors", "spells", "constants", "structures"}
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
		if def.ID == "" {
			return fmt.Errorf("%s: card definition missing required field 'id'", e.Name())
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
