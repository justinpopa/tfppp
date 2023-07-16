package main

import (
	"fmt"

	"github.com/justinpopa/tfppp/cmd"
	"github.com/spf13/cobra"
)

var (
	// These get overwritten by goreleaser during build using ldflags.
	version = "dev"
	commit  = "none"
	date    = "unknown"

	// versionCmd represents the version command.
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Version will output the current build information",
		Long:  ``,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("Version: %v\n", version)
			fmt.Printf("Commit: %v\n", commit)
			fmt.Printf("Date: %v\n", date)
		},
	}
)

func main() {
	cmd.AddCommand(versionCmd)
	cmd.Execute()
}
