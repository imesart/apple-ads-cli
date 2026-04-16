package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Profile holds all configuration for the aads CLI.
type Profile struct {
	ClientID         string `yaml:"client_id"`
	TeamID           string `yaml:"team_id"`
	KeyID            string `yaml:"key_id"`
	OrgID            string `yaml:"org_id"`
	PrivateKeyPath   string `yaml:"private_key_path"`
	DefaultCurrency  string `yaml:"default_currency"`
	DefaultTimezone  string `yaml:"default_timezone"`
	DefaultTimeOfDay string `yaml:"default_time_of_day"`

	// Safety limits as decimal text in default_currency.
	// The empty string means disabled.
	MaxDailyBudget  DecimalText `yaml:"max_daily_budget"`
	MaxBid          DecimalText `yaml:"max_bid"`
	MaxCPAGoal      DecimalText `yaml:"max_cpa_goal"`
	MaxBudgetAmount DecimalText `yaml:"max_budget"`
}

// ConfigFile represents the on-disk YAML structure.
type ConfigFile struct {
	DefaultProfile string             `yaml:"default_profile"`
	Profiles       map[string]Profile `yaml:"profiles"`
}

var configDirOverride string

// SelectedProfile resolves the active profile name using documented precedence:
// explicit CLI/profile input first, then AADS_PROFILE, then empty string to
// allow Load("") to fall back to default_profile from the config file.
func SelectedProfile(profile string) string {
	if profile = strings.TrimSpace(profile); profile != "" {
		return profile
	}
	if profile = strings.TrimSpace(os.Getenv("AADS_PROFILE")); profile != "" {
		return profile
	}
	return ""
}

// SelectedTimeout resolves the active request timeout using documented
// precedence: AADS_TIMEOUT when set, otherwise the provided default.
// AADS_TIMEOUT is interpreted as a whole number of seconds and must be > 0.
func SelectedTimeout(defaultTimeout time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv("AADS_TIMEOUT"))
	if value == "" {
		return defaultTimeout, nil
	}

	seconds, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid AADS_TIMEOUT %q: must be a positive integer number of seconds", value)
	}
	if seconds <= 0 {
		return 0, fmt.Errorf("invalid AADS_TIMEOUT %q: must be greater than 0 seconds", value)
	}
	return time.Duration(seconds) * time.Second, nil
}

// Load reads configuration from ~/.aads/config.yaml for the given profile.
// If profile is empty, the default profile from the file is used (or "default").
// Supports the structured profiles format (default_profile + profiles map) and
// the flat format (single config). After loading from file, environment
// variable overrides are applied.
func Load(profile string) (*Profile, error) {
	configPath := DefaultConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No config file; return empty config with env overrides
			cfg := &Profile{}
			cfg.ApplyEnv()
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	cfg, err := parseConfig(data, profile)
	if err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	cfg.ApplyEnv()
	return cfg, nil
}

// parseConfig tries the structured profiles format first, then the flat format.
func parseConfig(data []byte, profile string) (*Profile, error) {
	// Try structured profiles format: default_profile + profiles map
	var cf ConfigFile
	if err := yaml.Unmarshal(data, &cf); err == nil && len(cf.Profiles) > 0 {
		if profile == "" {
			profile = cf.DefaultProfile
		}
		if profile == "" {
			profile = "default"
		}
		if cfg, ok := cf.Profiles[profile]; ok {
			cfg.PrivateKeyPath = expandPath(cfg.PrivateKeyPath)
			return &cfg, nil
		}
		return nil, fmt.Errorf("profile %q not found in config file", profile)
	}

	if looksLikeLegacyProfileMap(data) {
		return nil, fmt.Errorf("unsupported legacy config format: use default_profile + profiles")
	}

	// Try flat format: single config at top level
	var cfg Profile
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}
	cfg.PrivateKeyPath = expandPath(cfg.PrivateKeyPath)
	return &cfg, nil
}

// LoadFile reads the full ConfigFile from disk.
// Returns an empty ConfigFile if the file does not exist or cannot be parsed.
func LoadFile() *ConfigFile {
	configPath := DefaultConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		return &ConfigFile{Profiles: make(map[string]Profile)}
	}

	// Try new format
	var cf ConfigFile
	if err := yaml.Unmarshal(data, &cf); err == nil && len(cf.Profiles) > 0 {
		return &cf
	}

	if looksLikeLegacyProfileMap(data) {
		return &ConfigFile{Profiles: make(map[string]Profile)}
	}

	// Try flat format
	var cfg Profile
	if err := yaml.Unmarshal(data, &cfg); err == nil && cfg.ClientID != "" {
		return &ConfigFile{
			DefaultProfile: "default",
			Profiles:       map[string]Profile{"default": cfg},
		}
	}

	return &ConfigFile{Profiles: make(map[string]Profile)}
}

// SaveFile writes the ConfigFile to disk.
func SaveFile(cf *ConfigFile) error {
	configPath := DefaultConfigPath()
	configDir := filepath.Dir(configPath)

	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(cf)
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// ApplyEnv applies environment variable overrides to the config.
// Environment variables take precedence over file-based configuration.
//
// Supported variables:
//   - AADS_CLIENT_ID
//   - AADS_TEAM_ID
//   - AADS_KEY_ID
//   - AADS_ORG_ID
//   - AADS_PRIVATE_KEY_PATH
func (c *Profile) ApplyEnv() {
	if v := os.Getenv("AADS_CLIENT_ID"); v != "" {
		c.ClientID = v
	}
	if v := os.Getenv("AADS_TEAM_ID"); v != "" {
		c.TeamID = v
	}
	if v := os.Getenv("AADS_KEY_ID"); v != "" {
		c.KeyID = v
	}
	if v := os.Getenv("AADS_ORG_ID"); v != "" {
		c.OrgID = v
	}
	if v := os.Getenv("AADS_PRIVATE_KEY_PATH"); v != "" {
		c.PrivateKeyPath = expandPath(v)
	}
}

// Validate checks that all required fields are populated.
func (c *Profile) Validate() error {
	var missing []string
	if c.ClientID == "" {
		missing = append(missing, "client_id")
	}
	if c.TeamID == "" {
		missing = append(missing, "team_id")
	}
	if c.KeyID == "" {
		missing = append(missing, "key_id")
	}
	if c.PrivateKeyPath == "" {
		missing = append(missing, "private_key_path")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required config fields: %s", strings.Join(missing, ", "))
	}
	return nil
}

// DefaultConfigDir returns the default configuration directory (~/.aads).
func DefaultConfigDir() string {
	if override := strings.TrimSpace(configDirOverride); override != "" {
		return expandPath(override)
	}
	if override := strings.TrimSpace(os.Getenv("AADS_CONFIG_DIR")); override != "" {
		return expandPath(override)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".aads")
	}
	return filepath.Join(home, ".aads")
}

// SetConfigDir overrides the configuration directory for the current process.
// An empty path clears the override and falls back to AADS_CONFIG_DIR or the
// default ~/.aads path.
func SetConfigDir(path string) {
	configDirOverride = strings.TrimSpace(path)
}

// DefaultConfigPath returns the default config file path (~/.aads/config.yaml).
func DefaultConfigPath() string {
	return filepath.Join(DefaultConfigDir(), "config.yaml")
}

// DefaultTokenCachePath returns the default token cache path (~/.aads/token.json).
func DefaultTokenCachePath() string {
	return filepath.Join(DefaultConfigDir(), "token.json")
}

// expandPath replaces a leading ~ with the user's home directory.
func expandPath(path string) string {
	if path == "" {
		return path
	}
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}

func looksLikeLegacyProfileMap(data []byte) bool {
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil || len(raw) == 0 {
		return false
	}

	if _, ok := raw["profiles"]; ok {
		return false
	}

	if hasFlatProfileKeys(raw) {
		return false
	}

	for _, v := range raw {
		if _, ok := v.(map[string]any); ok {
			return true
		}
	}

	return false
}

func hasFlatProfileKeys(raw map[string]any) bool {
	for _, key := range []string{
		"client_id",
		"team_id",
		"key_id",
		"org_id",
		"private_key_path",
		"default_currency",
		"default_timezone",
		"default_time_of_day",
		"max_daily_budget",
		"max_bid",
		"max_cpa_goal",
		"max_budget",
	} {
		if _, ok := raw[key]; ok {
			return true
		}
	}

	return false
}
