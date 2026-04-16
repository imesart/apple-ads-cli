package shared

import (
	"fmt"

	"github.com/imesart/apple-ads-cli/internal/output"
)

// PrintOutput prints data in the specified format.
// Optional entityIDName sets the column name for the "id" field in ids format.
func PrintOutput(data any, format string, fields string, pretty bool, entityIDName ...string) error {
	outFormat := output.Format(format)
	name := ""
	if len(entityIDName) > 0 {
		name = entityIDName[0]
	}
	if summary, ok := data.(MutationCheckSummary); ok {
		data = summary.OutputData()
	}
	if fields != "" {
		if outFormat == output.FormatIDs || outFormat == output.FormatPipe {
			return fmt.Errorf("--fields cannot be used with -f %s", outFormat)
		}
		filtered, err := output.SelectFields(data, fields, name)
		if err != nil {
			return err
		}
		data = filtered
	}
	return output.PrintWithPretty(outFormat, data, pretty, entityIDName...)
}
