package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	// Theme settings
	Theme Theme `json:"theme"`

	// Keybindings can override defaults
	Keybindings map[string]string `json:"keybindings,omitempty"`

	// Performance settings
	LargeDiffThreshold int `json:"large_diff_threshold"` // Lines before showing warning
	MaxContextLines    int `json:"max_context_lines"`    // Context lines in diff
}

// Theme defines color settings
type Theme struct {
	Added    string `json:"added"`
	Removed  string `json:"removed"`
	Context  string `json:"context"`
	Hunk     string `json:"hunk"`
	LineNum  string `json:"line_num"`
	Selected string `json:"selected"`
	Border   string `json:"border"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		Theme: Theme{
			Added:    "#40ff40",
			Removed:  "#ff4040",
			Context:  "#fafafa",
			Hunk:     "#8888ff",
			LineNum:  "#888888",
			Selected: "#4444ff",
			Border:   "#404040",
		},
		LargeDiffThreshold: 5000,
		MaxContextLines:    3,
	}
}

// Load loads configuration from standard paths
func Load() (Config, error) {
	cfg := DefaultConfig()

	// Try paths in order
	paths := []string{
		".gdiff.json",
		"gdiff.json",
	}

	// Add home config path
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".config", "gdiff", "gdiff.json"))
	}

	for _, path := range paths {
		if data, err := os.ReadFile(path); err == nil {
			if err := json.Unmarshal(data, &cfg); err != nil {
				return cfg, err
			}
			return cfg, nil
		}
	}

	return cfg, nil
}

// Save saves configuration to the given path
func Save(cfg Config, path string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
