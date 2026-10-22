package main

import (
	"fmt"
	"os"

	"github.com/matthiasharzer/ilias-sync-webdav/cmd/sync"
	"github.com/matthiasharzer/ilias-sync-webdav/cmd/version"

	"github.com/spf13/cobra"
)

var rootCommand = &cobra.Command{
	Use: "ilias-sync-webdav",
	RunE: func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCommand.AddCommand(version.Command)
	rootCommand.AddCommand(sync.Command)
}

func main() {
	err := rootCommand.Execute()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
