package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// RootFlags holds CLI flag values for kvtools.
type RootFlags struct {
	AllNamespaces      bool
	Namespace          string
	Output             string
	OutputFile         string
	Kubeconfig         string
	Context            string
	GuestSubresources  bool
	HealthOnly         bool
	Concurrency        int
	Quiet              bool
}

var flags RootFlags

var rootCmd = &cobra.Command{
	Use:   "kvtools [sheet]",
	Short: "kvtools - RVTools for KubeVirt",
	Long: `kvtools is a comprehensive inventory extraction and health auditing tool for KubeVirt clusters.
It inspects virtual machines, compute nodes, storage, networks, snapshots, and guest agent metrics,
running automated health audits and exporting to styled multi-tab Excel (.xlsx), JSON, CSV, or terminal tables.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default action is to execute export
		return runExport(cmd, args)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&flags.AllNamespaces, "all-namespaces", "A", true, "Query resources across all namespaces")
	rootCmd.PersistentFlags().StringVarP(&flags.Namespace, "namespace", "n", "", "Filter resources by a specific namespace")
	rootCmd.PersistentFlags().StringVarP(&flags.Output, "output", "o", "excel", "Output format: excel, json, csv, table")
	rootCmd.PersistentFlags().StringVarP(&flags.OutputFile, "output-file", "f", "", "Output file path (default: kvtools_<cluster>_<timestamp>.xlsx)")
	rootCmd.PersistentFlags().StringVar(&flags.Kubeconfig, "kubeconfig", "", "Path to kubeconfig file")
	rootCmd.PersistentFlags().StringVar(&flags.Context, "context", "", "Kubernetes context to use")
	rootCmd.PersistentFlags().BoolVar(&flags.GuestSubresources, "guest-subresources", true, "Fetch guest agent subresources (guestosinfo, filesystemlist)")
	rootCmd.PersistentFlags().BoolVar(&flags.HealthOnly, "health-only", false, "Only run and display kvHealth audit findings")
	rootCmd.PersistentFlags().IntVarP(&flags.Concurrency, "concurrency", "c", 10, "Worker pool concurrency for guest subresource queries")
	rootCmd.PersistentFlags().BoolVarP(&flags.Quiet, "quiet", "q", false, "Suppress progress spinner and informational messages")
}
