package output

import (
	"bytes"
	"encoding/json"
	"os"

	"golang.org/x/term"
)

// UnmarshalUseNumber decodes JSON using json.Number instead of float64
// for numeric values. This preserves the exact string representation of
// numbers, avoiding precision loss for large IDs and maintaining decimal
// formatting for monetary amounts.
func UnmarshalUseNumber(data []byte, out any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	return dec.Decode(out)
}

// PrintJSON writes data as JSON to stdout.
// When pretty is true or stdout is a TTY, output is pretty-printed with indentation.
// Otherwise output is compact (single line).
func PrintJSON(data any, pretty bool) error {
	enc := json.NewEncoder(os.Stdout)
	if pretty || term.IsTerminal(int(os.Stdout.Fd())) {
		enc.SetIndent("", "  ")
	}
	enc.SetEscapeHTML(false)
	return enc.Encode(data)
}
