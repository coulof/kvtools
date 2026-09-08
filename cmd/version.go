package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	// Version is the current semantic release version of kvtools.
	Version = "v0.1.0-dev"
	// GitCommit is the commit SHA injected during build.
	GitCommit = "unknown"
	// BuildDate is the RFC3339 build timestamp injected during build.
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version and build metadata of kvtools",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("kvtools %s\n", Version)
		fmt.Printf("  Git Commit: %s\n", GitCommit)
		fmt.Printf("  Build Date: %s\n", BuildDate)
		fmt.Printf("  Go Version: %s\n", runtime.Version())
		fmt.Printf("  OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
