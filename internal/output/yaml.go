package output

import (
	"os"

	"gopkg.in/yaml.v3"
)

// PrintYAML writes data as YAML to stdout.
func PrintYAML(data any) error {
	enc := yaml.NewEncoder(os.Stdout)
	enc.SetIndent(2)
	if err := enc.Encode(data); err != nil {
		return err
	}
	return enc.Close()
}
