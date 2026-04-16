package cmd

import "testing"

func TestExitCodeConstants(t *testing.T) {
	tests := []struct {
		name  string
		code  int
		value int
	}{
		{"ExitSuccess", ExitSuccess, 0},
		{"ExitError", ExitError, 1},
		{"ExitUsage", ExitUsage, 2},
		{"ExitAuth", ExitAuth, 3},
		{"ExitAPIError", ExitAPIError, 4},
		{"ExitNetworkError", ExitNetworkError, 5},
		{"ExitSafetyLimit", ExitSafetyLimit, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.value {
				t.Errorf("%s = %d, want %d", tt.name, tt.code, tt.value)
			}
		})
	}
}

func TestExitCodesAreUnique(t *testing.T) {
	codes := []int{
		ExitSuccess, ExitError, ExitUsage,
		ExitAuth, ExitAPIError, ExitNetworkError, ExitSafetyLimit,
	}

	seen := make(map[int]bool)
	for _, code := range codes {
		if seen[code] {
			t.Errorf("duplicate exit code: %d", code)
		}
		seen[code] = true
	}
}

func TestExitSuccessIsZero(t *testing.T) {
	if ExitSuccess != 0 {
		t.Errorf("ExitSuccess = %d, want 0", ExitSuccess)
	}
}
