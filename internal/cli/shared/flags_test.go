package shared

import (
	"flag"
	"testing"
)

func TestProfile_UsesEnvWhenFlagUnset(t *testing.T) {
	t.Setenv("AADS_PROFILE", "work")
	globalProfile = ""

	if got := Profile(); got != "work" {
		t.Fatalf("Profile() = %q, want %q", got, "work")
	}
}

func TestProfile_FlagOverridesEnv(t *testing.T) {
	t.Setenv("AADS_PROFILE", "work")
	globalProfile = "personal"

	if got := Profile(); got != "personal" {
		t.Fatalf("Profile() = %q, want %q", got, "personal")
	}
}

func TestBindOutputFlags_RegistersPretty(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	output := BindOutputFlags(fs)

	if err := fs.Parse([]string{"--pretty"}); err != nil {
		t.Fatalf("Parse(--pretty) returned error: %v", err)
	}
	if output.Pretty == nil || !*output.Pretty {
		t.Fatalf("expected --pretty to set OutputFlags.Pretty")
	}
}
