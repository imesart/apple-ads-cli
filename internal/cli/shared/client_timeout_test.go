package shared

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/imesart/apple-ads-cli/internal/config"
)

func resetClientStateForTest() func() {
	prevOnce := clientOnce
	prevState := cachedState
	prevErr := clientErr

	clientOnce = &sync.Once{}
	cachedState = nil
	clientErr = nil

	return func() {
		clientOnce = prevOnce
		cachedState = prevState
		clientErr = prevErr
	}
}

func writeConfigForTimeoutTest(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("MkdirAll(%q): %v", dir, err)
	}
	content := `
default_profile: default
profiles:
  default:
    client_id: client
    team_id: team
    key_id: key
    org_id: "123"
    private_key_path: /tmp/key.pem
`
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(config.yaml): %v", err)
	}
}

func TestContextWithTimeout_UsesEnvOverride(t *testing.T) {
	t.Setenv("AADS_TIMEOUT", "120")

	before := time.Now()
	ctx, cancel := ContextWithTimeout(context.Background())
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("ContextWithTimeout() did not set a deadline")
	}
	got := deadline.Sub(before)
	if got < 119*time.Second || got > 121*time.Second {
		t.Fatalf("deadline delta = %v, want about 120s", got)
	}
}

func TestGetClient_InvalidTimeout(t *testing.T) {
	restore := resetClientStateForTest()
	defer restore()

	t.Setenv("AADS_TIMEOUT", "abc")
	configDir := t.TempDir()
	writeConfigForTimeoutTest(t, configDir)
	config.SetConfigDir(configDir)
	defer config.SetConfigDir("")

	_, err := GetClient()
	if err == nil {
		t.Fatal("GetClient() should fail for invalid AADS_TIMEOUT")
	}
	if got := err.Error(); got == "" || !containsAll(got, "config validation", "AADS_TIMEOUT") {
		t.Fatalf("GetClient() error = %q, want mention of config validation and AADS_TIMEOUT", got)
	}
}

func containsAll(s string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(s, part) {
			return false
		}
	}
	return true
}
