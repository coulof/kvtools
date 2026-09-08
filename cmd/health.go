package cmd

import (
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Run the kvHealth best-practice, hygiene, and migration-readiness audit",
	Long:  "Inspects all virtual machines, nodes, and storage to diagnose migration blockers, orphaned resources, and sizing risks.",
	RunE: func(cmd *cobra.Command, args []string) error {
		flags.HealthOnly = true
		flags.Output = "table"
		return runExport(cmd, []string{"kvHealth"})
	},
}

func init() {
	rootCmd.AddCommand(healthCmd)
}
