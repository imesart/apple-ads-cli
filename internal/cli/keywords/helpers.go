package keywords

import (
	"encoding/json"

	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

func readBodyFile(path string) (json.RawMessage, error) {
	return shared.ReadJSONInputArg(path)
}

type Fields struct {
	Bid          map[string]string
	Status       string
	MatchType    string
	MatchTypeSet bool
}

// Validate normalizes Status and MatchType to surface invalid flag values
// without writing to a payload map. Use this before building per-item maps
// so a bad --match-type fails fast even when no items will be produced.
func (f Fields) Validate() error {
	if f.Status != "" {
		if _, err := shared.NormalizeStatus(f.Status, "ACTIVE"); err != nil {
			return err
		}
	}
	if f.MatchTypeSet || f.MatchType != "" {
		if _, err := shared.NormalizeMatchType(f.MatchType); err != nil {
			return err
		}
	}
	return nil
}

func ApplyFields(m map[string]any, fields Fields) error {
	if fields.Bid != nil {
		m["bidAmount"] = fields.Bid
	}
	if fields.Status != "" {
		s, err := shared.NormalizeStatus(fields.Status, "ACTIVE")
		if err != nil {
			return err
		}
		m["status"] = s
	}
	if fields.MatchTypeSet || fields.MatchType != "" {
		mt, err := shared.NormalizeMatchType(fields.MatchType)
		if err != nil {
			return err
		}
		m["matchType"] = mt
	}
	return nil
}

func ValidatePayload(body json.RawMessage) error {
	return shared.CheckBidLimitJSON(body)
}
