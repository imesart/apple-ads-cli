package negatives

import (
	"encoding/json"

	"github.com/imesart/apple-ads-cli/internal/cli/shared"
)

func readBodyFile(path string) (json.RawMessage, error) {
	return shared.ReadJSONInputArg(path)
}
