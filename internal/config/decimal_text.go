package config

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"gopkg.in/yaml.v3"
)

// DecimalText stores a non-negative decimal as text in YAML.
// The empty string means "disabled".
type DecimalText string

// ParseDecimalText parses a config limit value.
// Empty string and zero both normalize to disabled.
func ParseDecimalText(value string) (DecimalText, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return DecimalText(""), nil
	}
	d, err := decimal.NewFromString(value)
	if err != nil {
		return "", fmt.Errorf("invalid decimal %q", value)
	}
	if d.IsNegative() {
		return "", fmt.Errorf("negative limits are not allowed")
	}
	if d.IsZero() {
		return DecimalText(""), nil
	}
	return DecimalText(d.String()), nil
}

// String returns the normalized decimal text.
func (d DecimalText) String() string {
	return string(d)
}

// Enabled reports whether the limit is set.
func (d DecimalText) Enabled() bool {
	return strings.TrimSpace(string(d)) != ""
}

// Decimal parses the value as a decimal. Disabled values return ok=false.
func (d DecimalText) Decimal() (decimal.Decimal, bool, error) {
	if !d.Enabled() {
		return decimal.Zero, false, nil
	}
	parsed, err := decimal.NewFromString(string(d))
	if err != nil {
		return decimal.Zero, false, err
	}
	return parsed, true, nil
}

func (d *DecimalText) UnmarshalYAML(node *yaml.Node) error {
	if node == nil {
		*d = DecimalText("")
		return nil
	}
	parsed, err := ParseDecimalText(node.Value)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

func (d DecimalText) MarshalYAML() (any, error) {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: string(d),
		Style: yaml.DoubleQuotedStyle,
	}, nil
}
