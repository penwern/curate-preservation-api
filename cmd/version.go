package cmd

import (
	"fmt"
	"runtime"

	"github.com/penwern/curate-preservation-api/pkg/version"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Print the version information for the preservation API server.`,
	Run: func(_ *cobra.Command, _ []string) {
		//nolint:forbidigo // Version command needs to output directly to stdout
		fmt.Printf("Preservation API Server\n")
		//nolint:forbidigo // Version command needs to output directly to stdout
		fmt.Printf("Version:    %s\n", version.Version())
		//nolint:forbidigo // Version command needs to output directly to stdout
		fmt.Printf("Git Commit: %s\n", version.Commit())
		//nolint:forbidigo // Version command needs to output directly to stdout
		fmt.Printf("Build Date: %s\n", version.BuildTime())
		//nolint:forbidigo // Version command needs to output directly to stdout
		fmt.Printf("Go Version: %s\n", runtime.Version())
		//nolint:forbidigo // Version command needs to output directly to stdout
		fmt.Printf("OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
