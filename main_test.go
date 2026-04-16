package main

import (
	"strings"
	"testing"
)

func TestVersionInfoString(t *testing.T) {
	s := versionInfoString()
	if s == "" {
		t.Error("versionInfoString() returned empty string")
	}
}

func TestVersionInfoString_ContainsVersion(t *testing.T) {
	s := versionInfoString()
	// Default version is "dev" when not set by ldflags
	if !strings.Contains(s, version) {
		t.Errorf("versionInfoString() = %q, should contain version %q", s, version)
	}
}

func TestVersionInfoString_ContainsCommit(t *testing.T) {
	s := versionInfoString()
	if !strings.Contains(s, commit) {
		t.Errorf("versionInfoString() = %q, should contain commit %q", s, commit)
	}
}

func TestVersionInfoString_ContainsDate(t *testing.T) {
	s := versionInfoString()
	if !strings.Contains(s, date) {
		t.Errorf("versionInfoString() = %q, should contain date %q", s, date)
	}
}

func TestVersionInfoString_Format(t *testing.T) {
	s := versionInfoString()
	// Should match format: "version (commit: xxx, date: xxx)"
	if !strings.Contains(s, "commit:") {
		t.Errorf("versionInfoString() = %q, should contain 'commit:'", s)
	}
	if !strings.Contains(s, "date:") {
		t.Errorf("versionInfoString() = %q, should contain 'date:'", s)
	}
}
