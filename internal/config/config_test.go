package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func clearEnv(t *testing.T) {
	t.Helper()
	t.Setenv("AADS_CLIENT_ID", "")
	t.Setenv("AADS_TEAM_ID", "")
	t.Setenv("AADS_KEY_ID", "")
	t.Setenv("AADS_ORG_ID", "")
	t.Setenv("AADS_PRIVATE_KEY_PATH", "")
	t.Setenv("AADS_CONFIG_DIR", "")
	t.Setenv("AADS_TIMEOUT", "")
	SetConfigDir("")
}

func writeConfig(t *testing.T, dir, content string) {
	t.Helper()
	aadsDir := filepath.Join(dir, ".aads")
	if err := os.MkdirAll(aadsDir, 0o700); err != nil {
		t.Fatalf("MkdirAll(%q): %v", aadsDir, err)
	}
	configPath := filepath.Join(aadsDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(%q): %v", configPath, err)
	}
	t.Setenv("HOME", dir)
}

func writeConfigToDir(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("MkdirAll(%q): %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(%q): %v", filepath.Join(dir, "config.yaml"), err)
	}
}

// ============================================================
// New format: default_profile + profiles map
// ============================================================

func TestLoad_NewFormat(t *testing.T) {
	dir := t.TempDir()
	clearEnv(t)
	writeConfig(t, dir, `
default_profile: work
profiles:
  personal:
    client_id: personal-client
    team_id: personal-team
    key_id: personal-key
    private_key_path: /keys/personal.pem
  work:
    client_id: work-client
    team_id: work-team
    key_id: work-key
    org_id: "123"
    private_key_path: /keys/work.pem
    default_timezone: Europe/Luxembourg
    default_time_of_day: "09:30"
    max_daily_budget: 500.0
`)

	// Empty profile uses default_profile from file ("work")
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load(\"\") error: %v", err)
	}
	if cfg.ClientID != "work-client" {
		t.Errorf("ClientID = %q, want %q", cfg.ClientID, "work-client")
	}
	if cfg.OrgID != "123" {
		t.Errorf("OrgID = %q, want %q", cfg.OrgID, "123")
	}
	if cfg.MaxDailyBudget != DecimalText("500") {
		t.Errorf("MaxDailyBudget = %q, want %q", cfg.MaxDailyBudget, DecimalText("500"))
	}
	if cfg.DefaultTimezone != "Europe/Luxembourg" {
		t.Errorf("DefaultTimezone = %q, want %q", cfg.DefaultTimezone, "Europe/Luxembourg")
	}
	if cfg.DefaultTimeOfDay != "09:30" {
		t.Errorf("DefaultTimeOfDay = %q, want %q", cfg.DefaultTimeOfDay, "09:30")
	}

	// Explicit profile
	cfg2, err := Load("personal")
	if err != nil {
		t.Fatalf("Load(personal) error: %v", err)
	}
	if cfg2.ClientID != "personal-client" {
		t.Errorf("ClientID = %q, want %q", cfg2.ClientID, "personal-client")
	}
}

func TestLoad_NewFormat_NonexistentProfile(t *testing.T) {
	dir := t.TempDir()
	clearEnv(t)
	writeConfig(t, dir, `
default_profile: default
profiles:
  default:
    client_id: default-client
    team_id: default-team
    key_id: default-key
    private_key_path: /keys/default.pem
`)

	_, err := Load("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent profile")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error should mention profile name, got: %v", err)
	}
}

func TestLoad_NewFormat_NoDefaultProfile(t *testing.T) {
	dir := t.TempDir()
	clearEnv(t)
	writeConfig(t, dir, `
profiles:
  default:
    client_id: default-client
    team_id: default-team
    key_id: default-key
    private_key_path: /keys/default.pem
`)

	// No default_profile set, empty profile => falls back to "default"
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load(\"\") error: %v", err)
	}
	if cfg.ClientID != "default-client" {
		t.Errorf("ClientID = %q, want %q", cfg.ClientID, "default-client")
	}
}

func TestSelectedProfile(t *testing.T) {
	clearEnv(t)
	t.Setenv("AADS_PROFILE", "work")

	if got := SelectedProfile(""); got != "work" {
		t.Fatalf("SelectedProfile(\"\") = %q, want %q", got, "work")
	}
	if got := SelectedProfile("personal"); got != "personal" {
		t.Fatalf("SelectedProfile(\"personal\") = %q, want %q", got, "personal")
	}
	if got := SelectedProfile("  personal  "); got != "personal" {
		t.Fatalf("SelectedProfile(\"  personal  \") = %q, want %q", got, "personal")
	}

	t.Setenv("AADS_PROFILE", "")
	if got := SelectedProfile(""); got != "" {
		t.Fatalf("SelectedProfile(\"\") without env = %q, want empty", got)
	}
}

func TestSelectedTimeout(t *testing.T) {
	clearEnv(t)

	got, err := SelectedTimeout(30 * time.Second)
	if err != nil {
		t.Fatalf("SelectedTimeout(default) error: %v", err)
	}
	if got != 30*time.Second {
		t.Fatalf("SelectedTimeout(default) = %v, want %v", got, 30*time.Second)
	}

	t.Setenv("AADS_TIMEOUT", "120")
	got, err = SelectedTimeout(30 * time.Second)
	if err != nil {
		t.Fatalf("SelectedTimeout(env) error: %v", err)
	}
	if got != 120*time.Second {
		t.Fatalf("SelectedTimeout(env) = %v, want %v", got, 120*time.Second)
	}
}

func TestSelectedTimeout_Invalid(t *testing.T) {
	clearEnv(t)

	for _, value := range []string{"abc", "0", "-5"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("AADS_TIMEOUT", value)
			if _, err := SelectedTimeout(30 * time.Second); err == nil {
				t.Fatalf("SelectedTimeout(%q) should fail", value)
			}
		})
	}
}

func TestLoad_LegacyProfiledConfigUnsupported(t *testing.T) {
	dir := t.TempDir()
	clearEnv(t)
	writeConfig(t, dir, `
default:
  client_id: default-client
  team_id: default-team
  key_id: default-key
  private_key_path: /keys/default.pem
`)

	_, err := Load("default")
	if err == nil {
		t.Fatal("Load(default) should reject legacy profiled config")
	}
	if !strings.Contains(err.Error(), "unsupported legacy config format") {
		t.Errorf("error should mention unsupported legacy format, got: %v", err)
	}
}

// ============================================================
// Flat format (backward compatibility)
// ============================================================

func TestLoad_FlatConfig(t *testing.T) {
	dir := t.TempDir()
	clearEnv(t)
	writeConfig(t, dir, `
client_id: flat-client
team_id: flat-team
key_id: flat-key
org_id: "12345"
private_key_path: /path/to/key.pem
default_currency: USD
default_timezone: Europe/Luxembourg
default_time_of_day: "08:15"
max_daily_budget: 100.0
max_bid: 5.50
max_cpa_goal: 10.0
max_budget: 50000.0
`)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.ClientID != "flat-client" {
		t.Errorf("ClientID = %q, want %q", cfg.ClientID, "flat-client")
	}
	if cfg.TeamID != "flat-team" {
		t.Errorf("TeamID = %q, want %q", cfg.TeamID, "flat-team")
	}
	if cfg.KeyID != "flat-key" {
		t.Errorf("KeyID = %q, want %q", cfg.KeyID, "flat-key")
	}
	if cfg.OrgID != "12345" {
		t.Errorf("OrgID = %q, want %q", cfg.OrgID, "12345")
	}
	if cfg.PrivateKeyPath != "/path/to/key.pem" {
		t.Errorf("PrivateKeyPath = %q, want %q", cfg.PrivateKeyPath, "/path/to/key.pem")
	}
	if cfg.DefaultCurrency != "USD" {
		t.Errorf("DefaultCurrency = %q, want %q", cfg.DefaultCurrency, "USD")
	}
	if cfg.DefaultTimezone != "Europe/Luxembourg" {
		t.Errorf("DefaultTimezone = %q, want %q", cfg.DefaultTimezone, "Europe/Luxembourg")
	}
	if cfg.DefaultTimeOfDay != "08:15" {
		t.Errorf("DefaultTimeOfDay = %q, want %q", cfg.DefaultTimeOfDay, "08:15")
	}
	if cfg.MaxDailyBudget != DecimalText("100") {
		t.Errorf("MaxDailyBudget = %q, want %q", cfg.MaxDailyBudget, DecimalText("100"))
	}
	if cfg.MaxBid != DecimalText("5.5") {
		t.Errorf("MaxBid = %q, want %q", cfg.MaxBid, DecimalText("5.5"))
	}
	if cfg.MaxCPAGoal != DecimalText("10") {
		t.Errorf("MaxCPAGoal = %q, want %q", cfg.MaxCPAGoal, DecimalText("10"))
	}
	if cfg.MaxBudgetAmount != DecimalText("50000") {
		t.Errorf("MaxBudgetAmount = %q, want %q", cfg.MaxBudgetAmount, DecimalText("50000"))
	}
}

func TestLoad_FlatConfig_QuotedDecimalStrings(t *testing.T) {
	dir := t.TempDir()
	clearEnv(t)
	writeConfig(t, dir, `
client_id: flat-client
team_id: flat-team
key_id: flat-key
private_key_path: /path/to/key.pem
default_currency: USD
max_daily_budget: "100.00"
max_bid: "5.50"
max_cpa_goal: ""
max_budget: "50000.00"
`)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.MaxDailyBudget != DecimalText("100") {
		t.Errorf("MaxDailyBudget = %q, want %q", cfg.MaxDailyBudget, DecimalText("100"))
	}
	if cfg.MaxBid != DecimalText("5.5") {
		t.Errorf("MaxBid = %q, want %q", cfg.MaxBid, DecimalText("5.5"))
	}
	if cfg.MaxCPAGoal != DecimalText("") {
		t.Errorf("MaxCPAGoal = %q, want empty", cfg.MaxCPAGoal)
	}
	if cfg.MaxBudgetAmount != DecimalText("50000") {
		t.Errorf("MaxBudgetAmount = %q, want %q", cfg.MaxBudgetAmount, DecimalText("50000"))
	}
}

// ============================================================
// Edge cases
// ============================================================

func TestLoad_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	clearEnv(t)
	writeConfig(t, dir, "")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() with empty file error: %v", err)
	}
	if cfg.ClientID != "" {
		t.Errorf("ClientID = %q, want empty", cfg.ClientID)
	}
}

func TestLoad_NoConfigFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	clearEnv(t)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() with no config file error: %v", err)
	}
	if cfg == nil {
		t.Fatal("cfg should not be nil even without config file")
	}
}

// ============================================================
// LoadFile and SaveFile
// ============================================================

func TestLoadFile_NewFormat(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, `
default_profile: work
profiles:
  work:
    client_id: work-client
    team_id: work-team
    key_id: work-key
    private_key_path: /keys/work.pem
`)

	cf := LoadFile()
	if cf.DefaultProfile != "work" {
		t.Errorf("DefaultProfile = %q, want %q", cf.DefaultProfile, "work")
	}
	if len(cf.Profiles) != 1 {
		t.Fatalf("Profiles len = %d, want 1", len(cf.Profiles))
	}
	if cf.Profiles["work"].ClientID != "work-client" {
		t.Errorf("work ClientID = %q, want %q", cf.Profiles["work"].ClientID, "work-client")
	}
}

func TestLoadFile_LegacyFormatIgnored(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, `
default:
  client_id: default-client
  team_id: default-team
  key_id: default-key
  private_key_path: /keys/default.pem
`)

	cf := LoadFile()
	if cf.DefaultProfile != "" {
		t.Errorf("DefaultProfile = %q, want empty", cf.DefaultProfile)
	}
	if len(cf.Profiles) != 0 {
		t.Errorf("Profiles len = %d, want 0", len(cf.Profiles))
	}
}

func TestLoadFile_FlatFormat(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, `
client_id: flat-client
team_id: flat-team
key_id: flat-key
private_key_path: /keys/flat.pem
`)

	cf := LoadFile()
	if cf.DefaultProfile != "default" {
		t.Errorf("DefaultProfile = %q, want %q", cf.DefaultProfile, "default")
	}
	if cf.Profiles["default"].ClientID != "flat-client" {
		t.Errorf("default ClientID = %q, want %q", cf.Profiles["default"].ClientID, "flat-client")
	}
}

func TestSaveFile_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	aadsDir := filepath.Join(dir, ".aads")
	if err := os.MkdirAll(aadsDir, 0o700); err != nil {
		t.Fatalf("MkdirAll(%q): %v", aadsDir, err)
	}
	t.Setenv("HOME", dir)
	clearEnv(t)

	cf := &ConfigFile{
		DefaultProfile: "work",
		Profiles: map[string]Profile{
			"work": {
				ClientID:       "work-client",
				TeamID:         "work-team",
				KeyID:          "work-key",
				PrivateKeyPath: "/keys/work.pem",
				OrgID:          "123",
			},
			"personal": {
				ClientID:       "personal-client",
				TeamID:         "personal-team",
				KeyID:          "personal-key",
				PrivateKeyPath: "/keys/personal.pem",
			},
		},
	}

	if err := SaveFile(cf); err != nil {
		t.Fatalf("SaveFile error: %v", err)
	}

	// Reload and verify
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load after save error: %v", err)
	}
	if cfg.ClientID != "work-client" {
		t.Errorf("ClientID = %q, want %q (default_profile=work)", cfg.ClientID, "work-client")
	}

	cfg2, err := Load("personal")
	if err != nil {
		t.Fatalf("Load(personal) after save error: %v", err)
	}
	if cfg2.ClientID != "personal-client" {
		t.Errorf("personal ClientID = %q, want %q", cfg2.ClientID, "personal-client")
	}
}

func TestSaveFile_WritesQuotedDecimalStrings(t *testing.T) {
	dir := t.TempDir()
	aadsDir := filepath.Join(dir, ".aads")
	if err := os.MkdirAll(aadsDir, 0o700); err != nil {
		t.Fatalf("MkdirAll(%q): %v", aadsDir, err)
	}
	t.Setenv("HOME", dir)
	clearEnv(t)

	cf := &ConfigFile{
		DefaultProfile: "default",
		Profiles: map[string]Profile{
			"default": {
				ClientID:        "client",
				TeamID:          "team",
				KeyID:           "key",
				PrivateKeyPath:  "/keys/default.pem",
				DefaultCurrency: "USD",
				MaxDailyBudget:  DecimalText("100"),
				MaxBid:          DecimalText("5.5"),
				MaxCPAGoal:      DecimalText(""),
				MaxBudgetAmount: DecimalText("50000"),
			},
		},
	}

	if err := SaveFile(cf); err != nil {
		t.Fatalf("SaveFile error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(aadsDir, "config.yaml"))
	if err != nil {
		t.Fatalf("reading config file: %v", err)
	}
	text := string(data)
	for _, want := range []string{
		`max_daily_budget: "100"`,
		`max_bid: "5.5"`,
		`max_cpa_goal: ""`,
		`max_budget: "50000"`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("config file missing %q:\n%s", want, text)
		}
	}
}

// ============================================================
// Environment variable overrides
// ============================================================

func TestApplyEnv(t *testing.T) {
	t.Setenv("AADS_CLIENT_ID", "env-client")
	t.Setenv("AADS_TEAM_ID", "env-team")
	t.Setenv("AADS_KEY_ID", "env-key")
	t.Setenv("AADS_ORG_ID", "env-org")
	t.Setenv("AADS_PRIVATE_KEY_PATH", "/env/key.pem")

	cfg := &Profile{
		ClientID:       "file-client",
		TeamID:         "file-team",
		KeyID:          "file-key",
		OrgID:          "file-org",
		PrivateKeyPath: "/file/key.pem",
	}
	cfg.ApplyEnv()

	if cfg.ClientID != "env-client" {
		t.Errorf("ClientID = %q, want %q", cfg.ClientID, "env-client")
	}
	if cfg.TeamID != "env-team" {
		t.Errorf("TeamID = %q, want %q", cfg.TeamID, "env-team")
	}
	if cfg.KeyID != "env-key" {
		t.Errorf("KeyID = %q, want %q", cfg.KeyID, "env-key")
	}
	if cfg.OrgID != "env-org" {
		t.Errorf("OrgID = %q, want %q", cfg.OrgID, "env-org")
	}
	if cfg.PrivateKeyPath != "/env/key.pem" {
		t.Errorf("PrivateKeyPath = %q, want %q", cfg.PrivateKeyPath, "/env/key.pem")
	}
}

func TestApplyEnv_PartialOverride(t *testing.T) {
	t.Setenv("AADS_CLIENT_ID", "env-client")
	t.Setenv("AADS_TEAM_ID", "")
	t.Setenv("AADS_KEY_ID", "")
	t.Setenv("AADS_ORG_ID", "")
	t.Setenv("AADS_PRIVATE_KEY_PATH", "")

	cfg := &Profile{
		ClientID: "file-client",
		TeamID:   "file-team",
		KeyID:    "file-key",
	}
	cfg.ApplyEnv()

	if cfg.ClientID != "env-client" {
		t.Errorf("ClientID = %q, want %q", cfg.ClientID, "env-client")
	}
	if cfg.TeamID != "file-team" {
		t.Errorf("TeamID = %q, want %q (should not be overridden by empty env)", cfg.TeamID, "file-team")
	}
}

func TestLoad_EnvOverridesFile(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, `
client_id: file-client
team_id: file-team
key_id: file-key
private_key_path: /file/key.pem
`)
	t.Setenv("AADS_CLIENT_ID", "env-client")
	t.Setenv("AADS_TEAM_ID", "")
	t.Setenv("AADS_KEY_ID", "")
	t.Setenv("AADS_ORG_ID", "")
	t.Setenv("AADS_PRIVATE_KEY_PATH", "")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.ClientID != "env-client" {
		t.Errorf("ClientID = %q, want %q (env override)", cfg.ClientID, "env-client")
	}
	if cfg.TeamID != "file-team" {
		t.Errorf("TeamID = %q, want %q (from file)", cfg.TeamID, "file-team")
	}
}

// ============================================================
// Validation
// ============================================================

func TestValidate_MissingFields(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Profile
		wantErr []string
	}{
		{
			name:    "all missing",
			cfg:     Profile{},
			wantErr: []string{"client_id", "team_id", "key_id", "private_key_path"},
		},
		{
			name: "missing client_id only",
			cfg: Profile{
				TeamID:         "team",
				KeyID:          "key",
				PrivateKeyPath: "/path",
			},
			wantErr: []string{"client_id"},
		},
		{
			name: "missing team_id and key_id",
			cfg: Profile{
				ClientID:       "client",
				PrivateKeyPath: "/path",
			},
			wantErr: []string{"team_id", "key_id"},
		},
		{
			name: "missing private_key_path",
			cfg: Profile{
				ClientID: "client",
				TeamID:   "team",
				KeyID:    "key",
			},
			wantErr: []string{"private_key_path"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if err == nil {
				t.Fatal("Validate() should return error for missing fields")
			}
			for _, field := range tt.wantErr {
				if !strings.Contains(err.Error(), field) {
					t.Errorf("error should mention %q, got: %v", field, err)
				}
			}
		})
	}
}

func TestValidate_Valid(t *testing.T) {
	cfg := Profile{
		ClientID:       "client-id",
		TeamID:         "team-id",
		KeyID:          "key-id",
		PrivateKeyPath: "/path/to/key.pem",
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() should return nil for valid config, got: %v", err)
	}
}

// ============================================================
// Path helpers
// ============================================================

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot determine home dir: %v", err)
	}

	tests := []struct {
		name string
		path string
		want string
	}{
		{"tilde prefix", "~/some/path", filepath.Join(home, "some/path")},
		{"tilde only", "~", filepath.Join(home)},
		{"absolute path", "/absolute/path", "/absolute/path"},
		{"relative path", "relative/path", "relative/path"},
		{"empty path", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandPath(tt.path)
			if got != tt.want {
				t.Errorf("expandPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestDefaultPaths(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot determine home dir: %v", err)
	}

	expectedDir := filepath.Join(home, ".aads")
	if got := DefaultConfigDir(); got != expectedDir {
		t.Errorf("DefaultConfigDir() = %q, want %q", got, expectedDir)
	}

	expectedConfig := filepath.Join(home, ".aads", "config.yaml")
	if got := DefaultConfigPath(); got != expectedConfig {
		t.Errorf("DefaultConfigPath() = %q, want %q", got, expectedConfig)
	}

	expectedToken := filepath.Join(home, ".aads", "token.json")
	if got := DefaultTokenCachePath(); got != expectedToken {
		t.Errorf("DefaultTokenCachePath() = %q, want %q", got, expectedToken)
	}
}

func TestDefaultPaths_ConfigDirEnvOverride(t *testing.T) {
	clearEnv(t)
	dir := t.TempDir()
	t.Setenv("AADS_CONFIG_DIR", dir)

	if got := DefaultConfigDir(); got != dir {
		t.Errorf("DefaultConfigDir() = %q, want %q", got, dir)
	}
	if got := DefaultConfigPath(); got != filepath.Join(dir, "config.yaml") {
		t.Errorf("DefaultConfigPath() = %q, want %q", got, filepath.Join(dir, "config.yaml"))
	}
	if got := DefaultTokenCachePath(); got != filepath.Join(dir, "token.json") {
		t.Errorf("DefaultTokenCachePath() = %q, want %q", got, filepath.Join(dir, "token.json"))
	}
}

func TestDefaultPaths_ConfigDirSetterOverridesEnv(t *testing.T) {
	clearEnv(t)
	envDir := t.TempDir()
	overrideDir := t.TempDir()
	t.Setenv("AADS_CONFIG_DIR", envDir)
	SetConfigDir(overrideDir)

	if got := DefaultConfigDir(); got != overrideDir {
		t.Errorf("DefaultConfigDir() = %q, want %q", got, overrideDir)
	}
}

func TestLoad_UsesConfigDirOverride(t *testing.T) {
	clearEnv(t)
	home := t.TempDir()
	t.Setenv("HOME", home)

	overrideDir := filepath.Join(t.TempDir(), "custom-config")
	writeConfigToDir(t, overrideDir, `
default_profile: work
profiles:
  work:
    client_id: work-client
    team_id: work-team
    key_id: work-key
    private_key_path: /keys/work.pem
`)
	SetConfigDir(overrideDir)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load(\"\") error: %v", err)
	}
	if cfg.ClientID != "work-client" {
		t.Errorf("ClientID = %q, want %q", cfg.ClientID, "work-client")
	}
}
