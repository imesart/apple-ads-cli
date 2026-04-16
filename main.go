package main

import (
	"fmt"
	"os"

	"github.com/imesart/apple-ads-cli/cmd"
)

var (
	version          = "dev"
	commit           = "unknown"
	date             = "unknown"
	targetAPIVersion = "5.5"
)

func versionInfoString() string {
	return fmt.Sprintf("%s (commit: %s, date: %s)", version, commit, date)
}

func main() {
	os.Exit(cmd.Run(os.Args[1:], versionInfoString(), targetAPIVersion))
}
