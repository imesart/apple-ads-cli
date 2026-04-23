package shared

import (
	"encoding/json"
	"testing"

	apiPkg "github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/config"
	"github.com/imesart/apple-ads-cli/internal/types"
)

func TestCheckBudgetLimit_NilAmount(t *testing.T) {
	err := CheckBudgetLimit(nil)
	if err != nil {
		t.Errorf("CheckBudgetLimit(nil) = %v, want nil", err)
	}
}

func TestCheckBidLimit_NilAmount(t *testing.T) {
	err := CheckBidLimit(nil)
	if err != nil {
		t.Errorf("CheckBidLimit(nil) = %v, want nil", err)
	}
}

func TestCheckBudgetAmountLimit_NilAmount(t *testing.T) {
	err := CheckBudgetAmountLimit(nil)
	if err != nil {
		t.Errorf("CheckBudgetAmountLimit(nil) = %v, want nil", err)
	}
}

func TestCheckCPAGoalLimit_NilAmount(t *testing.T) {
	err := CheckCPAGoalLimit(nil)
	if err != nil {
		t.Errorf("CheckCPAGoalLimit(nil) = %v, want nil", err)
	}
}

func TestCheckBudgetLimit_WithForce(t *testing.T) {
	prevForce := globalForce
	globalForce = true
	defer func() { globalForce = prevForce }()

	amount := &types.Money{Amount: "999999.00", Currency: "USD"}
	err := CheckBudgetLimit(amount)
	if err != nil {
		t.Errorf("CheckBudgetLimit with force = %v, want nil", err)
	}
}

func TestCheckBidLimit_WithoutConfig(t *testing.T) {
	restoreClient := SetClientForTesting(apiPkg.NewClient(nil, "", false), &config.Profile{})
	defer restoreClient()

	amount := &types.Money{Amount: "100.00", Currency: "USD"}
	err := CheckBidLimit(amount)
	if err != nil {
		t.Errorf("CheckBidLimit without config limit = %v, want nil", err)
	}
}

func TestCheckBudgetAmountLimit_WithoutConfig(t *testing.T) {
	restoreClient := SetClientForTesting(apiPkg.NewClient(nil, "", false), &config.Profile{})
	defer restoreClient()

	amount := &types.Money{Amount: "100000.00", Currency: "USD"}
	err := CheckBudgetAmountLimit(amount)
	if err != nil {
		t.Errorf("CheckBudgetAmountLimit without config limit = %v, want nil", err)
	}
}

func TestCheckCPAGoalLimit_WithoutConfig(t *testing.T) {
	restoreClient := SetClientForTesting(apiPkg.NewClient(nil, "", false), &config.Profile{})
	defer restoreClient()

	amount := &types.Money{Amount: "50.00", Currency: "USD"}
	err := CheckCPAGoalLimit(amount)
	if err != nil {
		t.Errorf("CheckCPAGoalLimit without config limit = %v, want nil", err)
	}
}

func TestCheckBudgetLimit_InvalidAmount(t *testing.T) {
	// Non-numeric amount should return nil (not parseable, so no limit check)
	amount := &types.Money{Amount: "not-a-number", Currency: "USD"}
	err := CheckBudgetLimit(amount)
	if err != nil {
		t.Errorf("CheckBudgetLimit(invalid amount) = %v, want nil", err)
	}
}

func TestCheckBidLimit_InvalidAmount(t *testing.T) {
	// Non-numeric amount should return nil
	amount := &types.Money{Amount: "abc", Currency: "USD"}
	err := CheckBidLimit(amount)
	if err != nil {
		t.Errorf("CheckBidLimit(invalid amount) = %v, want nil", err)
	}
}

func TestNewSafetyError_Message(t *testing.T) {
	msg := "bid amount 5.00 exceeds limit 2.00 (use --force to override)"
	err := NewSafetyError(msg)

	if err.Error() != msg {
		t.Errorf("safety error message = %q, want %q", err.Error(), msg)
	}

	if !IsSafetyError(err) {
		t.Error("IsSafetyError returned false for NewSafetyError result")
	}
}

func TestCheckMoneyLimit_ZeroLimit(t *testing.T) {
	// Limit of 0 means disabled — should always pass
	amount := &types.Money{Amount: "999999.00", Currency: "USD"}
	err := checkMoneyLimit(amount, config.DecimalText(""), "USD", "test")
	if err != nil {
		t.Errorf("checkMoneyLimit with zero limit = %v, want nil", err)
	}
}

func TestCheckMoneyLimit_UnderLimit(t *testing.T) {
	amount := &types.Money{Amount: "50.00", Currency: "USD"}
	err := checkMoneyLimit(amount, config.DecimalText("100"), "USD", "test")
	if err != nil {
		t.Errorf("checkMoneyLimit under limit = %v, want nil", err)
	}
}

func TestCheckMoneyLimit_OverLimit(t *testing.T) {
	amount := &types.Money{Amount: "150.00", Currency: "USD"}
	err := checkMoneyLimit(amount, config.DecimalText("100"), "USD", "test budget")
	if err == nil {
		t.Fatal("checkMoneyLimit over limit should return error")
	}
	if !IsSafetyError(err) {
		t.Error("should be a safety error")
	}
}

func TestCheckMoneyLimit_CurrencyMismatch(t *testing.T) {
	amount := &types.Money{Amount: "50.00", Currency: "EUR"}
	err := checkMoneyLimit(amount, config.DecimalText("100"), "USD", "test budget")
	if err == nil {
		t.Fatal("checkMoneyLimit with currency mismatch should return error")
	}
	if !IsSafetyError(err) {
		t.Error("should be a safety error")
	}
}

func TestCheckMoneyLimit_CurrencyMismatch_NoDefaultCurrency(t *testing.T) {
	// No default currency configured — skip mismatch check, compare amounts
	amount := &types.Money{Amount: "50.00", Currency: "EUR"}
	err := checkMoneyLimit(amount, config.DecimalText("100"), "", "test budget")
	if err != nil {
		t.Errorf("checkMoneyLimit with no default currency = %v, want nil", err)
	}
}

func TestCheckMoneyLimit_CurrencyMismatch_NoCurrencyOnValue(t *testing.T) {
	// No currency on the value — skip mismatch check, compare amounts
	amount := &types.Money{Amount: "50.00", Currency: ""}
	err := checkMoneyLimit(amount, config.DecimalText("100"), "USD", "test budget")
	if err != nil {
		t.Errorf("checkMoneyLimit with no value currency = %v, want nil", err)
	}
}

func TestCheckMoneyLimit_CaseInsensitiveCurrency(t *testing.T) {
	amount := &types.Money{Amount: "50.00", Currency: "usd"}
	err := checkMoneyLimit(amount, config.DecimalText("100"), "USD", "test budget")
	if err != nil {
		t.Errorf("checkMoneyLimit with case-different currency = %v, want nil", err)
	}
}

func TestCheckBudgetLimitJSON_BudgetAmount(t *testing.T) {
	// CheckBudgetLimitJSON should also check budgetAmount field
	body := json.RawMessage(`{"budgetAmount":{"amount":"200.00","currency":"USD"}}`)
	// Without config, limits are 0, so no error
	err := CheckBudgetLimitJSON(body)
	if err != nil {
		t.Errorf("CheckBudgetLimitJSON without config = %v, want nil", err)
	}
}

func TestCheckBidLimitJSON_CPAGoal(t *testing.T) {
	// CheckBidLimitJSON should also check cpaGoal field
	body := json.RawMessage(`{"cpaGoal":{"amount":"10.00","currency":"USD"}}`)
	// Without config, limits are 0, so no error
	err := CheckBidLimitJSON(body)
	if err != nil {
		t.Errorf("CheckBidLimitJSON with cpaGoal without config = %v, want nil", err)
	}
}

func TestCheckBidLimitJSON_TargetCPA(t *testing.T) {
	body := json.RawMessage(`{"targetCpa":{"amount":"10.00","currency":"USD"}}`)
	err := CheckBidLimitJSON(body)
	if err != nil {
		t.Errorf("CheckBidLimitJSON with targetCpa without config = %v, want nil", err)
	}
}
