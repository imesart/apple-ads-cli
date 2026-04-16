package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/cli/completion"
	"github.com/imesart/apple-ads-cli/internal/cli/registry"
	"github.com/imesart/apple-ads-cli/internal/cli/shared"
	versioncmd "github.com/imesart/apple-ads-cli/internal/cli/version"
	"github.com/imesart/apple-ads-cli/internal/config"
)

var versionRequested bool

// RootCommand returns the root aads command.
func RootCommand(version string, targetAPIVersion string) *ffcli.Command {
	versionRequested = false
	subcommands := registry.Subcommands(version, targetAPIVersion)

	// Build the help subcommand that delegates to the matching command's usage.
	helpCmd := &ffcli.Command{
		Name:       "help",
		ShortUsage: "aads help <command> [subcommand]",
		ShortHelp:  "Show help for a command.",
		FlagSet:    flag.NewFlagSet("help", flag.ContinueOnError),
	}

	root := &ffcli.Command{
		Name:        "aads",
		ShortUsage:  "aads <subcommand> [flags]",
		ShortHelp:   "Fast, lightweight CLI for the Apple Ads Campaign Management API.",
		FlagSet:     flag.NewFlagSet("aads", flag.ContinueOnError),
		Subcommands: append(subcommands, helpCmd),
	}

	// help Exec needs to walk root's subcommands, so set it after root is created.
	helpCmd.Exec = func(ctx context.Context, args []string) error {
		if len(args) == 0 {
			// Show root help
			printCommandUsage(root)
			return nil
		}
		cmd := findSubcommand(root, args)
		if cmd == nil {
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", strings.Join(args, " "))
			printCommandUsage(root)
			return nil
		}
		printCommandUsage(cmd)
		return nil
	}

	root.FlagSet.BoolVar(&versionRequested, "version", false, "Print version and exit")
	shared.BindRootFlags(root.FlagSet)

	// Provide root to the completion command so it can walk the tree.
	completion.SetRoot(root)

	root.Exec = func(ctx context.Context, args []string) error {
		if versionRequested {
			fmt.Fprintln(os.Stdout, versioncmd.Render(version, targetAPIVersion))
			return nil
		}
		if len(args) > 0 {
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", args[0])
		}
		return flag.ErrHelp
	}

	return root
}

// printCommandUsage prints the usage text for a command, handling nil FlagSets safely.
func printCommandUsage(cmd *ffcli.Command) {
	// Ensure FlagSet is non-nil so DefaultUsageFunc doesn't panic
	if cmd.FlagSet == nil {
		cmd.FlagSet = flag.NewFlagSet(cmd.Name, flag.ContinueOnError)
	}
	usageFunc := cmd.UsageFunc
	if usageFunc == nil {
		usageFunc = ffcli.DefaultUsageFunc
	}
	fmt.Fprintln(os.Stdout, usageFunc(cmd))
}

// findSubcommand walks the command tree to find the subcommand matching the given path.
func findSubcommand(root *ffcli.Command, path []string) *ffcli.Command {
	current := root
	for _, name := range path {
		var found *ffcli.Command
		for _, sub := range current.Subcommands {
			if strings.EqualFold(sub.Name, name) {
				found = sub
				break
			}
		}
		if found == nil {
			return nil
		}
		current = found
	}
	return current
}

// helpFlagIndex returns the index of the first --help / -h token in args,
// or -1 if none is present.
func helpFlagIndex(args []string) int {
	for i, a := range args {
		switch a {
		case "--help", "-help", "-h", "--h":
			return i
		}
	}
	return -1
}

// resolveCommand walks args and descends into the deepest matching subcommand.
// Flag tokens and flag values are ignored; positional tokens that match a
// subcommand at the current level advance the walk.
func resolveCommand(root *ffcli.Command, args []string) *ffcli.Command {
	current := root
	for _, a := range args {
		if a == "--help" || a == "-help" || a == "-h" || a == "--h" {
			break
		}
		if strings.HasPrefix(a, "-") {
			continue
		}
		var found *ffcli.Command
		for _, sub := range current.Subcommands {
			if strings.EqualFold(sub.Name, a) {
				found = sub
				break
			}
		}
		if found == nil {
			continue
		}
		current = found
	}
	return current
}

// Run executes the CLI using the provided args and version string.
// Returns the intended process exit code.
func Run(args []string, versionInfo string, targetAPIVersion string) int {
	// Fast path for --version
	if len(args) == 1 && strings.TrimSpace(args[0]) == "--version" {
		fmt.Fprintln(os.Stdout, versioncmd.Render(versionInfo, targetAPIVersion))
		return ExitSuccess
	}

	root := RootCommand(versionInfo, targetAPIVersion)

	// Global --help / -h: resolve the command path and print its usage,
	// matching the behavior of `aads help <command>`.
	if helpFlagIndex(args) >= 0 {
		printCommandUsage(resolveCommand(root, args))
		return ExitSuccess
	}

	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := root.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitSuccess
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return ExitUsage
	}
	config.SetConfigDir(shared.ConfigDir())

	if err := root.Run(runCtx); err != nil {
		// Usage and validation errors were already printed to stderr.
		if shared.IsUsageError(err) || shared.IsValidationError(err) {
			return ExitUsage
		}
		// Normal help request (e.g., -h flag)
		if errors.Is(err, flag.ErrHelp) {
			return ExitSuccess
		}
		if !shared.IsReportedError(err) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		return exitCodeFromError(err)
	}

	return ExitSuccess
}

func exitCodeFromError(err error) int {
	if err == nil {
		return ExitSuccess
	}
	if shared.IsUsageError(err) || shared.IsValidationError(err) || err == flag.ErrHelp {
		return ExitUsage
	}
	if shared.IsAuthError(err) {
		return ExitAuth
	}
	if shared.IsSafetyError(err) {
		return ExitSafetyLimit
	}
	if shared.IsAPIError(err) {
		return ExitAPIError
	}
	if shared.IsNetworkError(err) {
		return ExitNetworkError
	}
	return ExitError
}
