package version

import (
	"context"
	"fmt"

	"github.com/peterbourgon/ff/v3/ffcli"
)

const targetAPIName = "Apple Ads Campaign Management API"

func TargetAPI(targetAPIVersion string) string {
	if targetAPIVersion == "" {
		return targetAPIName
	}
	return fmt.Sprintf("%s v%s", targetAPIName, targetAPIVersion)
}

// Render returns the version output shown by both `aads version` and `aads --version`.
func Render(versionStr string, targetAPIVersion string) string {
	return fmt.Sprintf("%s\nTarget API: %s", versionStr, TargetAPI(targetAPIVersion))
}

// Command returns the version command.
func Command(versionStr string, targetAPIVersion string) *ffcli.Command {
	return &ffcli.Command{
		Name:       "version",
		ShortUsage: "aads version",
		ShortHelp:  "Print the CLI version and target Apple Ads API version.",
		Exec: func(ctx context.Context, args []string) error {
			fmt.Println(Render(versionStr, targetAPIVersion))
			return nil
		},
	}
}
