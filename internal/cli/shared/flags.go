package shared

import (
	"flag"
	"os"

	"github.com/imesart/apple-ads-cli/internal/config"
	"golang.org/x/term"

	"github.com/imesart/apple-ads-cli/internal/output"
)

// Global flag values
var (
	globalProfile   string
	globalOrgID     string
	globalVerbose   bool
	globalNoColor   bool
	globalForce     bool
	globalCurrency  string
	globalConfigDir string
)

// BindRootFlags binds global flags to the root flag set.
func BindRootFlags(fs *flag.FlagSet) {
	globalProfile = ""
	globalOrgID = ""
	globalVerbose = false
	globalNoColor = false
	globalCurrency = ""
	globalConfigDir = ""

	fs.StringVar(&globalProfile, "profile", "", "Config profile name")
	fs.StringVar(&globalProfile, "p", "", "--profile (shorthand)")
	fs.StringVar(&globalConfigDir, "config-dir", "", "Configuration directory")
	fs.StringVar(&globalOrgID, "org-id", "", "Override org ID from config")
	fs.BoolVar(&globalVerbose, "verbose", false, "Verbose output, show HTTP requests/responses")
	fs.BoolVar(&globalVerbose, "v", false, "--verbose (shorthand)")
	fs.BoolVar(&globalNoColor, "no-color", false, "Disable color output")
	fs.StringVar(&globalCurrency, "currency", "", "Override currency for money fields")
}

// BindForceFlag binds --force to the shared global safety override.
// Use this on subcommands that perform safety-limit validation so the flag is
// accepted both before and after the subcommand name.
func BindForceFlag(fs *flag.FlagSet) {
	fs.BoolVar(&globalForce, "force", false, "Skip safety checks")
}

// Global flag accessors
func Profile() string   { return config.SelectedProfile(globalProfile) }
func ConfigDir() string { return globalConfigDir }
func OrgID() string     { return globalOrgID }
func Verbose() bool     { return globalVerbose }
func NoColor() bool     { return globalNoColor }
func Force() bool       { return globalForce }
func Currency() string  { return globalCurrency }

// OutputFlags holds output format flags for a command.
type OutputFlags struct {
	Output *string
	Fields *string
	Pretty *bool
}

// BindOutputFlags registers --format / -f and --fields on the given flag set.
func BindOutputFlags(fs *flag.FlagSet) OutputFlags {
	def := string(output.DefaultFormat())
	o := fs.String("format", def, "Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe")
	fs.StringVar(o, "f", def, "--format (shorthand)")
	fields := fs.String("fields", "", "Comma-separated output fields to include")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON even when stdout is not a TTY")
	return OutputFlags{Output: o, Fields: fields, Pretty: pretty}
}

// IsTTY returns true if stdout is a terminal.
func IsTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
