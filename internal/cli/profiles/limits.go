package profiles

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/imesart/apple-ads-cli/internal/config"
)

func parseProfileLimitFlag(value string, defaultCurrency string) (config.DecimalText, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return config.DecimalText(""), nil
	}

	parts := strings.Fields(value)
	if len(parts) > 2 {
		return "", fmt.Errorf("invalid limit format %q: use \"AMOUNT\" or \"AMOUNT CURRENCY\"", value)
	}

	amountText := parts[0]
	parsed, err := decimal.NewFromString(amountText)
	if err != nil {
		return "", fmt.Errorf("invalid limit amount %q", amountText)
	}
	if parsed.IsNegative() {
		return "", fmt.Errorf("negative limits are not allowed")
	}
	if len(parts) == 2 {
		if strings.TrimSpace(defaultCurrency) == "" {
			return "", fmt.Errorf("limit currency requires default currency to be set")
		}
		if !strings.EqualFold(parts[1], defaultCurrency) {
			return "", fmt.Errorf("limit currency %s does not match default currency %s", strings.ToUpper(parts[1]), strings.ToUpper(defaultCurrency))
		}
	}

	return config.ParseDecimalText(parsed.String())
}
