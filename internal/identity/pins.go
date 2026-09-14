package identity

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Pin is one entry of familiar's identities.yaml. At least one of Remote, Path,
// Project is set. Members are theme-scoped and ignored here.
type Pin struct {
	Remote  string `yaml:"remote"`
	Path    string `yaml:"path"`
	Project string `yaml:"project"`
	Slot    *int   `yaml:"slot"`
}

type Pins struct {
	Identities []Pin `yaml:"identities"`
}

// LoadPins reads ~/.config/familiar/identities.yaml. Missing is no pins; malformed is an error.
func LoadPins(path string) (Pins, error) {
	text, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Pins{}, nil
	}
	if err != nil {
		return Pins{}, err
	}
	return ParsePins(text)
}

func ParsePins(text []byte) (Pins, error) {
	var raw struct {
		Identities yaml.Node `yaml:"identities"`
	}
	if err := yaml.Unmarshal(text, &raw); err != nil {
		return Pins{}, fmt.Errorf("identities.yaml: %w", err)
	}
	if raw.Identities.Kind == 0 {
		return Pins{}, nil
	}
	if raw.Identities.Kind != yaml.SequenceNode {
		return Pins{}, errors.New("identities.yaml: `identities` must be a list")
	}
	var pins Pins
	if err := raw.Identities.Decode(&pins.Identities); err != nil {
		return Pins{}, fmt.Errorf("identities.yaml: %w", err)
	}
	for i, p := range pins.Identities {
		if p.Remote == "" && p.Path == "" && p.Project == "" {
			return Pins{}, fmt.Errorf("identities.yaml: pin %d needs one of: remote, path, project", i)
		}
		if p.Slot == nil || *p.Slot < 0 || *p.Slot >= SlotCount {
			return Pins{}, fmt.Errorf("identities.yaml: pin %d needs a slot in 0..%d", i, SlotCount-1)
		}
	}
	return pins, nil
}

// canonical expands ~, makes a path absolute, and resolves symlinks. A path that
// does not exist keeps its lexical form, so a not-yet-cloned pin is inert.
func canonical(path string, realpath func(string) (string, error)) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[2:])
		}
	}
	if absolute, err := filepath.Abs(path); err == nil {
		path = absolute
	}
	if real, err := realpath(path); err == nil {
		return real
	}
	return filepath.Clean(path)
}

// MatchPin applies familiar's precedence: remote > path > project. remote is the
// normalized key ("" when none); root is the registered root; project its basename.
func MatchPin(pins Pins, remote, root, project string, realpath func(string) (string, error)) (int, bool) {
	if remote != "" {
		for _, p := range pins.Identities {
			if p.Remote != "" && strings.EqualFold(p.Remote, remote) {
				return *p.Slot, true
			}
		}
	}
	if root != "" {
		canonRoot := canonical(root, realpath)
		for _, p := range pins.Identities {
			if p.Path != "" && canonical(p.Path, realpath) == canonRoot {
				return *p.Slot, true
			}
		}
	}
	if project != "" {
		for _, p := range pins.Identities {
			if p.Project == project {
				return *p.Slot, true
			}
		}
	}
	return 0, false
}
