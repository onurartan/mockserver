package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	mslogger "mockserver/logger"
)

// LoadConfig reads a JSON or YAML config file, applies defaults, and validates required fields.
// Supports .json, .yaml, .yml extensions.
// Returns a fully populated Config or an error if loading or validation fails.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file '%s': %w", path, err)
	}

	var cfg Config
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse JSON in '%s': %w", path, err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse YAML in '%s': %w", path, err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file extension '%s', must be .json, .yaml or .yml", ext)
	}

	// Apply defaults and validate
	if err := validateAndApplyDefaults(&cfg, path); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	mslogger.LogSuccess(fmt.Sprintf("Config loaded successfully from %s", path), 1, -1)
	return &cfg, nil
}
