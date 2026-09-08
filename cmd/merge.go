package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/coulof/kvtools/pkg/exporter/excel"
	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
)

var mergeOutputFile string

var mergeCmd = &cobra.Command{
	Use:   "merge <file1.xlsx> <file2.xlsx> [files...]",
	Short: "Merge multiple kvtools Excel export workbooks into a single consolidated workbook",
	Long: `Combines multiple kvtools .xlsx export files from different clusters into a single multi-tab spreadsheet.
Preserves sheet formatting, auto-filters, frozen panes, and styling across all inventory tabs.`,
	Args: cobra.MinimumNArgs(2),
	RunE: runMerge,
}

func init() {
	mergeCmd.Flags().StringVarP(&mergeOutputFile, "output-file", "o", "kvtools_merged.xlsx", "Path for the merged output Excel file")
	rootCmd.AddCommand(mergeCmd)
}

func runMerge(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("at least two Excel files are required to merge")
	}

	outPath := mergeOutputFile
	if !strings.HasSuffix(outPath, ".xlsx") {
		outPath += ".xlsx"
	}

	mergedFile := excelize.NewFile()
	defer func() {
		_ = mergedFile.Close()
	}()

	evenStyle, _ := excel.CreateDataRowStyle(mergedFile, true)
	oddStyle, _ := excel.CreateDataRowStyle(mergedFile, false)
	critStyle, _ := excel.CreateHealthSeverityStyle(mergedFile, "CRITICAL")
	warnStyle, _ := excel.CreateHealthSeverityStyle(mergedFile, "WARNING")
	infoStyle, _ := excel.CreateHealthSeverityStyle(mergedFile, "INFO")

	sheetNames := []string{
		"kvInfo", "kvCPU", "kvMemory", "kvDisk", "kvPartition",
		"kvNetwork", "kvCD", "kvSnapshot", "kvGuestAgent",
		"kvNode", "kvStoragePool", "kvHardware", "kvHealth",
	}

	for _, sheet := range sheetNames {
		mergedFile.NewSheet(sheet)
		styleCfg := excel.SheetStyles[sheet]
		headerStyle, _ := excel.CreateHeaderStyle(mergedFile, styleCfg.HeaderBg)

		var headers []string
		var allDataRows [][]string

		for _, srcPath := range args {
			cleanPath := filepath.Clean(srcPath)
			f, err := excelize.OpenFile(cleanPath)
			if err != nil {
				return fmt.Errorf("failed to open input file %s: %w", cleanPath, err)
			}

			rows, err := f.GetRows(sheet)
			_ = f.Close()
			if err != nil || len(rows) == 0 {
				continue
			}

			if len(headers) == 0 {
				headers = rows[0]
			}

			// Append data rows
			if len(rows) > 1 {
				allDataRows = append(allDataRows, rows[1:]...)
			}
		}

		if len(headers) == 0 {
			continue
		}

		// Write headers
		maxLens := make([]int, len(headers))
		for cIdx, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, 1)
			_ = mergedFile.SetCellValue(sheet, cell, h)
			_ = mergedFile.SetCellStyle(sheet, cell, cell, headerStyle)
			maxLens[cIdx] = len(h)
		}

		// Write all combined data rows
		for rIdx, row := range allDataRows {
			rowNum := rIdx + 2
			rowStyle := oddStyle
			if rIdx%2 == 0 {
				rowStyle = evenStyle
			}

			if sheet == "kvHealth" && len(row) > 2 {
				sev := row[2] // Severity column
				if sev == "CRITICAL" {
					rowStyle = critStyle
				} else if sev == "WARNING" {
					rowStyle = warnStyle
				} else {
					rowStyle = infoStyle
				}
			}

			for cIdx, val := range row {
				cell, _ := excelize.CoordinatesToCellName(cIdx+1, rowNum)
				_ = mergedFile.SetCellValue(sheet, cell, val)
				_ = mergedFile.SetCellStyle(sheet, cell, cell, rowStyle)

				if cIdx < len(maxLens) && len(val) > maxLens[cIdx] {
					maxLens[cIdx] = len(val)
				}
			}
		}

		_ = excel.SetupSheetView(mergedFile, sheet, len(headers), len(allDataRows)+1)
		excel.AutoFitColumnWidths(mergedFile, sheet, len(headers), maxLens)
	}

	_ = mergedFile.DeleteSheet("Sheet1")
	infoIdx, err := mergedFile.GetSheetIndex("kvInfo")
	if err == nil && infoIdx >= 0 {
		mergedFile.SetActiveSheet(infoIdx)
	}

	if err := mergedFile.SaveAs(outPath); err != nil {
		return fmt.Errorf("failed to save merged Excel file to %s: %w", outPath, err)
	}

	fmt.Printf("\n\033[32;1m✔ Successfully merged %d Excel files into:\033[0m \033[1m%s\033[0m\n", len(args), outPath)
	return nil
}
