package shared

import "os"

// ReadJSONInputArg reads a JSON-bearing argument.
// - "@file.json" reads from file
// - "@-" or "-" reads from stdin
// - anything else is treated as inline JSON text
func ReadJSONInputArg(arg string) ([]byte, error) {
	if arg == "-" {
		return os.ReadFile("/dev/stdin")
	}
	if len(arg) > 0 && arg[0] == '@' {
		path := arg[1:]
		if path == "-" {
			return os.ReadFile("/dev/stdin")
		}
		return os.ReadFile(path)
	}
	return []byte(arg), nil
}

// IsStdinJSONInputArg reports whether the argument reads JSON from stdin.
// Both the documented "@-" form and the accepted "-" alias are supported.
func IsStdinJSONInputArg(arg string) bool {
	return arg == "@-" || arg == "-"
}
