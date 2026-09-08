package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/coulof/kvtools/pkg/client"
	"github.com/coulof/kvtools/pkg/collector"
	"github.com/coulof/kvtools/pkg/engine"
	"github.com/coulof/kvtools/pkg/exporter/csv"
	"github.com/coulof/kvtools/pkg/exporter/excel"
	"github.com/coulof/kvtools/pkg/exporter/json"
	"github.com/coulof/kvtools/pkg/exporter/table"
	"github.com/coulof/kvtools/pkg/utils"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export [sheet]",
	Short: "Export KubeVirt inventory and health audit to Excel, JSON, CSV, or Table",
	RunE:  runExport,
}

var validSheets = []string{
	"kvInfo", "kvCPU", "kvMemory", "kvDisk", "kvPartition",
	"kvNetwork", "kvCD", "kvSnapshot", "kvGuestAgent",
	"kvNode", "kvStoragePool", "kvHardware",
}

func init() {
	rootCmd.AddCommand(exportCmd)

	// Register top-level shorthand commands for each sheet (e.g. `kvtools kvNode` or `kvtools node`)
	for _, sheet := range validSheets {
		sName := sheet
		sheetCmd := &cobra.Command{
			Use:     sName,
			Aliases: []string{strings.ToLower(sName), strings.TrimPrefix(strings.ToLower(sName), "kv")},
			Short:   fmt.Sprintf("Inspect %s inventory sheet in the terminal", sName),
			RunE: func(cmd *cobra.Command, args []string) error {
				if !cmd.Flags().Changed("output") {
					flags.Output = "table"
				}
				return runExport(cmd, []string{sName})
			},
		}
		rootCmd.AddCommand(sheetCmd)
	}
}

func runExport(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// If namespace flag was explicitly provided, disable all-namespaces
	allNamespaces := flags.AllNamespaces
	if cmd.Flags().Changed("namespace") && flags.Namespace != "" {
		allNamespaces = false
	}

	prog := utils.NewProgress(flags.Quiet)

	// Build client configuration
	clientOpts := client.DefaultConfigOptions()
	if flags.Kubeconfig != "" {
		clientOpts.KubeconfigPath = flags.Kubeconfig
	}
	if flags.Context != "" {
		clientOpts.Context = flags.Context
	}

	k8sClient, err := client.NewClient(clientOpts)
	if err != nil {
		prog.Error("Failed to initialize Kubernetes client")
		return fmt.Errorf("client initialization error: %w", err)
	}

	// Setup collection options
	colOpts := collector.CollectorOptions{
		Namespace:              flags.Namespace,
		AllNamespaces:          allNamespaces,
		FetchGuestSubresources: flags.GuestSubresources,
		Concurrency:            flags.Concurrency,
		Timeout:                60 * time.Second,
	}

	col := collector.NewCollector(k8sClient)
	rawData, err := col.Collect(ctx, colOpts, prog)
	if err != nil {
		prog.Error("Data collection failed")
		return fmt.Errorf("data collection failed: %w", err)
	}

	// Transform data
	if prog != nil {
		prog.Start("Transforming inventory and running kvHealth audit engine")
	}
	report := engine.Transform(rawData)
	if prog != nil {
		prog.Success(fmt.Sprintf("Report ready (%d VMs, %d Nodes, %d Health Findings)",
			report.Summary.TotalVMs, report.Summary.TotalNodes, report.Summary.HealthIssuesCount))
	}

	// Handle health-only override
	if flags.HealthOnly {
		flags.Output = "table"
		if len(args) == 0 {
			args = []string{"kvHealth"}
		}
	}

	// Export formatting
	targetSheet := ""
	if len(args) > 0 {
		targetSheet = args[0]
	}

	switch strings.ToLower(flags.Output) {
	case "json":
		if flags.OutputFile != "" {
			f, err := os.Create(flags.OutputFile)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer f.Close()
			if err := json.ExportJSON(report, f); err != nil {
				return err
			}
			if !flags.Quiet {
				fmt.Printf("✔ JSON inventory exported to %s\n", flags.OutputFile)
			}
		} else {
			if err := json.ExportJSON(report, os.Stdout); err != nil {
				return err
			}
		}

	case "csv":
		if flags.OutputFile != "" {
			if strings.HasSuffix(flags.OutputFile, ".csv") {
				f, err := os.Create(flags.OutputFile)
				if err != nil {
					return fmt.Errorf("failed to create output file: %w", err)
				}
				defer f.Close()
				sheetName := targetSheet
				if sheetName == "" {
					sheetName = "kvInfo"
				}
				if err := csv.ExportSheetToCSV(report, sheetName, f); err != nil {
					return err
				}
				if !flags.Quiet {
					fmt.Printf("✔ CSV sheet '%s' exported to %s\n", sheetName, flags.OutputFile)
				}
			} else {
				files, err := csv.ExportAllSheetsToCSV(report, flags.OutputFile)
				if err != nil {
					return err
				}
				if !flags.Quiet {
					fmt.Printf("✔ 13 CSV sheets exported to directory %s/ (%d files)\n", flags.OutputFile, len(files))
				}
			}
		} else {
			sheetName := targetSheet
			if sheetName == "" {
				sheetName = "kvInfo"
			}
			if err := csv.ExportSheetToCSV(report, sheetName, os.Stdout); err != nil {
				return err
			}
		}

	case "table":
		table.RenderTableSheet(report, targetSheet, os.Stdout)

	case "excel", "xlsx", "":
		fileName := flags.OutputFile
		if fileName == "" {
			safeClusterName := strings.ReplaceAll(report.ClusterName, "/", "_")
			safeClusterName = strings.ReplaceAll(safeClusterName, ":", "_")
			fileName = fmt.Sprintf("kvtools_%s_%s.xlsx", safeClusterName, time.Now().Format("20060102_150405"))
		}
		if !strings.HasSuffix(fileName, ".xlsx") {
			fileName += ".xlsx"
		}

		if err := excel.ExportExcel(report, fileName); err != nil {
			return fmt.Errorf("failed to generate Excel export: %w", err)
		}

		if !flags.Quiet {
			fmt.Printf("\n\033[32;1m✔ Excel export successfully saved to:\033[0m \033[1m%s\033[0m\n", fileName)
			table.RenderSummary(report, os.Stdout)
		}

	default:
		return fmt.Errorf("unsupported output format '%s' (valid: excel, json, csv, table)", flags.Output)
	}

	return nil
}
