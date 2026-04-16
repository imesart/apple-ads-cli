package campaigns

import (
	"encoding/json"
	"fmt"

	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

func readBodyFile(path string) (json.RawMessage, error) {
	return shared.ReadJSONInputArg(path)
}

func readJSONObjectArg(arg string) (map[string]any, error) {
	if arg == "" {
		return nil, nil
	}
	raw, err := shared.ReadJSONInputArg(arg)
	if err != nil {
		return nil, err
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, fmt.Errorf("expected JSON object: %w", err)
	}
	return obj, nil
}
