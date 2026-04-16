package registry

import "testing"

func TestSubcommands_AllGroupsRegistered(t *testing.T) {
	subs := Subcommands("test-version", "5.5")

	expected := []string{
		"campaigns", "adgroups", "keywords", "negatives", "ads", "creatives",
		"budgetorders", "product-pages", "ad-rejections",
		"reports", "impression-share",
		"apps", "geo", "orgs",
		"structure", "profiles", "version", "schema", "completion",
	}

	if len(subs) != len(expected) {
		t.Fatalf("got %d subcommands, want %d", len(subs), len(expected))
	}

	names := make(map[string]bool)
	for _, sub := range subs {
		names[sub.Name] = true
	}

	for _, name := range expected {
		if !names[name] {
			t.Errorf("missing command group: %s", name)
		}
	}
}

func TestSubcommands_Count(t *testing.T) {
	subs := Subcommands("test", "5.5")
	// Count all leaf subcommands
	total := 0
	for _, group := range subs {
		if len(group.Subcommands) == 0 {
			total++ // leaf command (version, schema)
		} else {
			total += len(group.Subcommands)
		}
	}
	// Should have at least 60 subcommands per spec
	if total < 60 {
		t.Errorf("total subcommands = %d, want at least 60", total)
	}
}

func TestSubcommands_NoDuplicateNames(t *testing.T) {
	subs := Subcommands("test-version", "5.5")
	seen := make(map[string]bool)
	for _, sub := range subs {
		if seen[sub.Name] {
			t.Errorf("duplicate command name: %s", sub.Name)
		}
		seen[sub.Name] = true
	}
}

func TestSubcommands_AllHaveNames(t *testing.T) {
	subs := Subcommands("test-version", "5.5")
	for i, sub := range subs {
		if sub.Name == "" {
			t.Errorf("subcommand at index %d has empty name", i)
		}
	}
}

func TestSubcommands_CoreWorkflowPresent(t *testing.T) {
	subs := Subcommands("test-version", "5.5")
	names := make(map[string]bool)
	for _, sub := range subs {
		names[sub.Name] = true
	}

	core := []string{"campaigns", "adgroups", "keywords", "ads"}
	for _, name := range core {
		if !names[name] {
			t.Errorf("missing core workflow command: %s", name)
		}
	}
}

func TestSubcommands_UtilityPresent(t *testing.T) {
	subs := Subcommands("test-version", "5.5")
	names := make(map[string]bool)
	for _, sub := range subs {
		names[sub.Name] = true
	}

	utility := []string{"structure", "profiles", "version", "schema"}
	for _, name := range utility {
		if !names[name] {
			t.Errorf("missing utility command: %s", name)
		}
	}
}

func TestSubcommands_GroupsHaveSubcommands(t *testing.T) {
	subs := Subcommands("test-version", "5.5")
	names := make(map[string]*struct {
		hasSubs bool
	})
	for _, sub := range subs {
		names[sub.Name] = &struct{ hasSubs bool }{hasSubs: len(sub.Subcommands) > 0}
	}

	// These groups should have subcommands (list, get, create, update, etc.)
	expectSubs := []string{"campaigns", "adgroups", "keywords", "ads"}
	for _, name := range expectSubs {
		info := names[name]
		if info == nil {
			t.Errorf("command %q not found", name)
			continue
		}
		if !info.hasSubs {
			t.Errorf("command %q expected to have subcommands but has none", name)
		}
	}
}
