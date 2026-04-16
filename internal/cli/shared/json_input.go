package shared

import "os"

// ReadJSONInputArg reads a JSON-bearing argument.
// - "@file.json" reads from file
// - "@-" reads from stdin
// - anything else is treated as inline JSON text
func ReadJSONInputArg(arg string) ([]byte, error) {
	if len(arg) > 0 && arg[0] == '@' {
		path := arg[1:]
		if path == "-" {
			return os.ReadFile("/dev/stdin")
		}
		return os.ReadFile(path)
	}
	return []byte(arg), nil
}

// IsStdinJSONInputArg reports whether the argument uses @- stdin notation.
func IsStdinJSONInputArg(arg string) bool {
	return arg == "@-"
}
