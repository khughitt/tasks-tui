// Package config reads ~/.config/tasks-tui/config.toml. Absent is fine; present
// and malformed, or carrying a key this version does not know, is an error.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/khughitt/tasks-tui/internal/identity"
	"github.com/khughitt/tasks-tui/internal/launch"
)

type Identity struct {
	Slot map[string]int `toml:"slot"`
}

type Config struct {
	Tasks          string        `toml:"tasks"`
	RefreshSeconds int           `toml:"refresh_seconds"`
	Launch         launch.Config `toml:"launch"`
	Identity       Identity      `toml:"identity"`
}

func Default() Config {
	return Config{
		Tasks:          "tasks",
		RefreshSeconds: 30,
		Launch:         launch.Default(),
		Identity:       Identity{Slot: map[string]int{}},
	}
}

func DefaultPath() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "tasks-tui", "config.toml")
}

// Load decodes path over Default(): scalars and lists replace, harness tables merge by
// name. path == "" means DefaultPath().
func Load(path string) (Config, error) {
	if path == "" {
		path = DefaultPath()
	}
	cfg := Default()
	text, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return Config{}, err
	}
	md, err := toml.Decode(string(text), &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("%s: %w", path, err)
	}
	if undecoded := md.Undecoded(); len(undecoded) > 0 {
		keys := make([]string, len(undecoded))
		for i, key := range undecoded {
			keys[i] = key.String()
		}
		return Config{}, fmt.Errorf("%s: unknown key(s): %s", path, strings.Join(keys, ", "))
	}
	mergeHarness(&cfg, md)
	return cfg, validate(path, cfg)
}

func mergeHarness(cfg *Config, md toml.MetaData) {
	defaults := launch.Default().Harness
	for name, defaultHarness := range defaults {
		if !md.IsDefined("launch", "harness", name) {
			cfg.Launch.Harness[name] = defaultHarness
			continue
		}
		if !md.IsDefined("launch", "harness", name, "command") {
			h := cfg.Launch.Harness[name]
			h.Command = defaultHarness.Command
			cfg.Launch.Harness[name] = h
		}
	}
}

func validate(path string, cfg Config) error {
	switch {
	case cfg.Tasks == "":
		return fmt.Errorf("%s: tasks must name the binary", path)
	case cfg.RefreshSeconds < 0:
		return fmt.Errorf("%s: refresh_seconds must be >= 0", path)
	case len(cfg.Launch.Terminal) == 0:
		return fmt.Errorf("%s: launch.terminal must not be empty", path)
	}
	for name, h := range cfg.Launch.Harness {
		if len(h.Command) == 0 {
			return fmt.Errorf("%s: launch.harness.%s.command must not be empty", path, name)
		}
	}
	for prefix, slot := range cfg.Identity.Slot {
		if slot < 0 || slot >= identity.SlotCount {
			return fmt.Errorf("%s: identity.slot.%s must be 0..%d", path, prefix, identity.SlotCount-1)
		}
	}
	return nil
}
