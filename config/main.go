package config

import (
	"encoding/json"
	"fmt"
	"os"
)

import (
	mslogger "mockserver/logger"
)

// [IMP_FUNC]
// LoadConfig reads a JSON config file, applies defaults, and validates required fields.
// Returns a fully populated Config or an error if loading or validation fails.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file '%s': %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse JSON in '%s': %w", path, err)
	}


	// Apply defaults and validate
	if err := validateAndApplyDefaults(&cfg, path); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	mslogger.LogSuccess(fmt.Sprintf("Config loaded successfully from %s", path), 1, -1)
	return &cfg, nil
}
