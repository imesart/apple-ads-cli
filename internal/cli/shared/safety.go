package shared

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/imesart/apple-ads-cli/internal/config"
	"github.com/imesart/apple-ads-cli/internal/types"
)

// checkMoneyLimit validates a money amount against a configured limit.
// label is used in error messages (e.g. "daily budget", "bid amount").
// defaultCurrency is the profile's default currency for mismatch detection.
func checkMoneyLimit(amount *types.Money, limit config.DecimalText, defaultCurrency, label string) error {
	if Force() || amount == nil || !limit.Enabled() {
		return nil
	}

	// Currency mismatch: if the value specifies a different currency than
	// the one limits are denominated in, we cannot safely compare.
	if amount.Currency != "" && defaultCurrency != "" &&
		!strings.EqualFold(amount.Currency, defaultCurrency) {
		return NewSafetyError(fmt.Sprintf(
			"%s uses currency %s but limits are in %s; cannot compare (use --force to override)",
			label, amount.Currency, defaultCurrency,
		))
	}

	val, err := decimal.NewFromString(amount.Amount)
	if err != nil {
		return nil
	}
	limitValue, _, err := limit.Decimal()
	if err != nil {
		return nil
	}
	if val.GreaterThan(limitValue) {
		return NewSafetyError(fmt.Sprintf(
			"%s %s exceeds limit %s (use --force to override)",
			label, val.StringFixed(2), limitValue.StringFixed(2),
		))
	}
	return nil
}

// CheckBudgetLimit validates a daily budget against the configured max.
func CheckBudgetLimit(amount *types.Money) error {
	cfg, err := GetConfig()
	if err != nil {
		return nil
	}
	return checkMoneyLimit(amount, cfg.MaxDailyBudget, cfg.DefaultCurrency, "daily budget")
}

// CheckBudgetAmountLimit validates a total budget against the configured max.
func CheckBudgetAmountLimit(amount *types.Money) error {
	cfg, err := GetConfig()
	if err != nil {
		return nil
	}
	return checkMoneyLimit(amount, cfg.MaxBudgetAmount, cfg.DefaultCurrency, "budget amount")
}

// CheckBidLimit validates a bid amount against the configured max.
func CheckBidLimit(amount *types.Money) error {
	cfg, err := GetConfig()
	if err != nil {
		return nil
	}
	return checkMoneyLimit(amount, cfg.MaxBid, cfg.DefaultCurrency, "bid amount")
}

// CheckCPAGoalLimit validates a CPA goal against the configured max.
func CheckCPAGoalLimit(amount *types.Money) error {
	cfg, err := GetConfig()
	if err != nil {
		return nil
	}
	return checkMoneyLimit(amount, cfg.MaxCPAGoal, cfg.DefaultCurrency, "CPA goal")
}

// CheckBudgetLimitJSON checks a raw JSON body for budget safety limits.
// Validates dailyBudgetAmount against max_daily_budget and
// budgetAmount against max_budget.
func CheckBudgetLimitJSON(body json.RawMessage) error {
	var payload struct {
		DailyBudgetAmount *types.Money `json:"dailyBudgetAmount"`
		BudgetAmount      *types.Money `json:"budgetAmount"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil // can't parse, skip check
	}
	if err := CheckBudgetLimit(payload.DailyBudgetAmount); err != nil {
		return err
	}
	return CheckBudgetAmountLimit(payload.BudgetAmount)
}

// CheckBidLimitJSON checks a raw JSON body for bid and CPA safety limits.
// Validates bidAmount and defaultBidAmount against max_bid, and
// cpaGoal/targetCpa against max_cpa_goal.
func CheckBidLimitJSON(body json.RawMessage) error {
	var payload struct {
		BidAmount        *types.Money `json:"bidAmount"`
		DefaultBidAmount *types.Money `json:"defaultBidAmount"`
		CPAGoal          *types.Money `json:"cpaGoal"`
		TargetCpa        *types.Money `json:"targetCpa"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}
	if err := CheckBidLimit(payload.BidAmount); err != nil {
		return err
	}
	if err := CheckBidLimit(payload.DefaultBidAmount); err != nil {
		return err
	}
	if err := CheckCPAGoalLimit(payload.CPAGoal); err != nil {
		return err
	}
	return CheckCPAGoalLimit(payload.TargetCpa)
}
