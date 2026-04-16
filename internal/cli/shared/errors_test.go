package shared

import (
	"errors"
	"flag"
	"fmt"
	"testing"

	"github.com/imesart/apple-ads-cli/internal/api"
)

func TestReportedError(t *testing.T) {
	inner := fmt.Errorf("something failed")
	err := ReportError(inner)

	if !IsReportedError(err) {
		t.Error("IsReportedError returned false for ReportError result")
	}

	if err.Error() != "something failed" {
		t.Errorf("error message = %q, want %q", err.Error(), "something failed")
	}
}

func TestReportedError_Unwrap(t *testing.T) {
	inner := fmt.Errorf("inner cause")
	err := ReportError(inner)

	if !errors.Is(err, inner) {
		t.Error("reported error does not unwrap to inner error")
	}
}

func TestIsReportedError_False(t *testing.T) {
	err := fmt.Errorf("normal error")
	if IsReportedError(err) {
		t.Error("IsReportedError returned true for a normal error")
	}
}

func TestIsReportedError_Nil(t *testing.T) {
	if IsReportedError(nil) {
		t.Error("IsReportedError returned true for nil")
	}
}

func TestSafetyError(t *testing.T) {
	err := NewSafetyError("budget too high")

	if !IsSafetyError(err) {
		t.Error("IsSafetyError returned false for NewSafetyError result")
	}

	if err.Error() != "budget too high" {
		t.Errorf("error message = %q, want %q", err.Error(), "budget too high")
	}
}

func TestIsSafetyError_False(t *testing.T) {
	err := fmt.Errorf("generic error")
	if IsSafetyError(err) {
		t.Error("IsSafetyError returned true for a generic error")
	}
}

func TestIsSafetyError_Nil(t *testing.T) {
	if IsSafetyError(nil) {
		t.Error("IsSafetyError returned true for nil")
	}
}

func TestUsageError(t *testing.T) {
	err := UsageError("missing required flag")

	if err.Error() != "missing required flag" {
		t.Errorf("error message = %q, want %q", err.Error(), "missing required flag")
	}

	if !IsUsageError(err) {
		t.Error("IsUsageError returned false for UsageError result")
	}

	// UsageError wraps flag.ErrHelp so ffcli prints usage
	if !errors.Is(err, flag.ErrHelp) {
		t.Error("UsageError should unwrap to flag.ErrHelp")
	}
}

func TestUsageErrorf(t *testing.T) {
	err := UsageErrorf("flag %q is required", "--org-id")

	expected := `flag "--org-id" is required`
	if err.Error() != expected {
		t.Errorf("error message = %q, want %q", err.Error(), expected)
	}

	if !IsUsageError(err) {
		t.Error("IsUsageError returned false for UsageErrorf result")
	}

	if !errors.Is(err, flag.ErrHelp) {
		t.Error("UsageErrorf should unwrap to flag.ErrHelp")
	}
}

func TestValidationError(t *testing.T) {
	err := ValidationError("invalid status value")

	if err.Error() != "invalid status value" {
		t.Errorf("error message = %q, want %q", err.Error(), "invalid status value")
	}

	if !IsValidationError(err) {
		t.Error("IsValidationError returned false for ValidationError result")
	}

	if errors.Is(err, flag.ErrHelp) {
		t.Error("ValidationError should not unwrap to flag.ErrHelp")
	}
}

func TestValidationErrorf(t *testing.T) {
	err := ValidationErrorf("flag %q is invalid", "--status")

	expected := `flag "--status" is invalid`
	if err.Error() != expected {
		t.Errorf("error message = %q, want %q", err.Error(), expected)
	}

	if !IsValidationError(err) {
		t.Error("IsValidationError returned false for ValidationErrorf result")
	}

	if errors.Is(err, flag.ErrHelp) {
		t.Error("ValidationErrorf should not unwrap to flag.ErrHelp")
	}
}

func TestIsAuthError_True(t *testing.T) {
	err := &api.APIError{StatusCode: 401}
	if !IsAuthError(err) {
		t.Error("IsAuthError returned false for 401 APIError")
	}
}

func TestIsAuthError_False(t *testing.T) {
	err := &api.APIError{StatusCode: 400}
	if IsAuthError(err) {
		t.Error("IsAuthError returned true for 400 APIError")
	}
}

func TestIsAuthError_Nil(t *testing.T) {
	if IsAuthError(nil) {
		t.Error("IsAuthError returned true for nil")
	}
}

func TestIsAuthError_GenericError(t *testing.T) {
	err := fmt.Errorf("some error")
	if IsAuthError(err) {
		t.Error("IsAuthError returned true for a generic error")
	}
}

func TestIsAPIError_True(t *testing.T) {
	err := &api.APIError{StatusCode: 400}
	if !IsAPIError(err) {
		t.Error("IsAPIError returned false for 400 APIError")
	}
}

func TestIsAPIError_422(t *testing.T) {
	err := &api.APIError{StatusCode: 422}
	if !IsAPIError(err) {
		t.Error("IsAPIError returned false for 422 APIError")
	}
}

func TestIsAPIError_499(t *testing.T) {
	err := &api.APIError{StatusCode: 499}
	if !IsAPIError(err) {
		t.Error("IsAPIError returned false for 499 APIError")
	}
}

func TestIsAPIError_500(t *testing.T) {
	err := &api.APIError{StatusCode: 500}
	if IsAPIError(err) {
		t.Error("IsAPIError returned true for 500 APIError (should be network)")
	}
}

func TestIsAPIError_Nil(t *testing.T) {
	if IsAPIError(nil) {
		t.Error("IsAPIError returned true for nil")
	}
}

func TestIsAPIError_GenericError(t *testing.T) {
	err := fmt.Errorf("generic")
	if IsAPIError(err) {
		t.Error("IsAPIError returned true for a generic error")
	}
}

func TestIsNetworkError_True(t *testing.T) {
	err := &api.APIError{StatusCode: 500}
	if !IsNetworkError(err) {
		t.Error("IsNetworkError returned false for 500 APIError")
	}
}

func TestIsNetworkError_502(t *testing.T) {
	err := &api.APIError{StatusCode: 502}
	if !IsNetworkError(err) {
		t.Error("IsNetworkError returned false for 502 APIError")
	}
}

func TestIsNetworkError_503(t *testing.T) {
	err := &api.APIError{StatusCode: 503}
	if !IsNetworkError(err) {
		t.Error("IsNetworkError returned false for 503 APIError")
	}
}

func TestIsNetworkError_400(t *testing.T) {
	err := &api.APIError{StatusCode: 400}
	if IsNetworkError(err) {
		t.Error("IsNetworkError returned true for 400 APIError")
	}
}

func TestIsNetworkError_Nil(t *testing.T) {
	if IsNetworkError(nil) {
		t.Error("IsNetworkError returned true for nil")
	}
}

func TestIsNetworkError_GenericError(t *testing.T) {
	err := fmt.Errorf("generic")
	if IsNetworkError(err) {
		t.Error("IsNetworkError returned true for a generic error")
	}
}
