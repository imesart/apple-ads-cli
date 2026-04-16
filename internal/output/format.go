package output

import (
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/term"
)

// Format represents an output format for CLI results.
type Format string

const (
	FormatJSON     Format = "json"
	FormatTable    Format = "table"
	FormatYAML     Format = "yaml"
	FormatMarkdown Format = "markdown"
	FormatIDs      Format = "ids"
	FormatPipe     Format = "pipe"
)

// DefaultFormat returns the appropriate default format based on whether
// stdout is a terminal. Returns table for TTY, json for pipes.
func DefaultFormat() Format {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		return FormatTable
	}
	return FormatJSON
}

// Print outputs data to stdout in the specified format.
// Optional entityIDName sets the column name for the "id" field in ids format.
func Print(format Format, data any, entityIDName ...string) error {
	return PrintWithPretty(format, data, false, entityIDName...)
}

// PrintWithPretty outputs data to stdout in the specified format, with an
// optional pretty-print override for JSON output.
func PrintWithPretty(format Format, data any, pretty bool, entityIDName ...string) error {
	data = unwrapOutputData(data)

	switch format {
	case FormatJSON:
		return PrintJSON(data, pretty)
	case FormatTable:
		return PrintTable(data)
	case FormatYAML:
		return PrintYAML(data)
	case FormatMarkdown:
		return PrintMarkdown(data)
	case FormatIDs:
		name := ""
		if len(entityIDName) > 0 {
			name = entityIDName[0]
		}
		return PrintIDs(data, name)
	case FormatPipe:
		return PrintPipe(data)
	default:
		return fmt.Errorf("unsupported output format: %q", format)
	}
}

func unwrapOutputData(data any) any {
	switch v := data.(type) {
	case OrderedData:
		return v
	case json.RawMessage:
		return unwrapRawData(v)
	case []byte:
		return unwrapRawData(json.RawMessage(v))
	default:
		return unwrapGenericData(v)
	}
}

func unwrapRawData(raw json.RawMessage) any {
	var decoded any
	if err := UnmarshalUseNumber(raw, &decoded); err != nil {
		return raw
	}
	return unwrapGenericData(decoded)
}

func unwrapGenericData(data any) any {
	obj, ok := data.(map[string]any)
	if !ok {
		return data
	}
	unwrapped, ok := obj["data"]
	if !ok {
		return data
	}
	if unwrapped == nil {
		return []any{}
	}
	switch unwrapped.(type) {
	case map[string]any, []any:
		return unwrapped
	default:
		return data
	}
}
