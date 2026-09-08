package cmd

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/coulof/kvtools/pkg/engine"
	"github.com/coulof/kvtools/pkg/exporter/excel"
	"github.com/xuri/excelize/v2"
)

func TestMergeCommand(t *testing.T) {
	tempDir := t.TempDir()

	file1 := filepath.Join(tempDir, "cluster1.xlsx")
	file2 := filepath.Join(tempDir, "cluster2.xlsx")
	mergedFile := filepath.Join(tempDir, "merged.xlsx")

	rep1 := &engine.InventoryReport{
		ClusterName: "cluster-1",
		GeneratedAt: time.Now(),
		Info: []engine.KVInfoRecord{
			{VMName: "vm-c1-01", Namespace: "default", PowerState: "Running"},
		},
	}

	rep2 := &engine.InventoryReport{
		ClusterName: "cluster-2",
		GeneratedAt: time.Now(),
		Info: []engine.KVInfoRecord{
			{VMName: "vm-c2-01", Namespace: "default", PowerState: "Stopped"},
		},
	}

	if err := excel.ExportExcel(rep1, file1); err != nil {
		t.Fatalf("failed to export file1: %v", err)
	}
	if err := excel.ExportExcel(rep2, file2); err != nil {
		t.Fatalf("failed to export file2: %v", err)
	}

	mergeOutputFile = mergedFile
	err := runMerge(mergeCmd, []string{file1, file2})
	if err != nil {
		t.Fatalf("failed to merge Excel files: %v", err)
	}

	// Verify merged file
	f, err := excelize.OpenFile(mergedFile)
	if err != nil {
		t.Fatalf("failed to open merged Excel file: %v", err)
	}
	defer f.Close()

	rows, err := f.GetRows("kvInfo")
	if err != nil {
		t.Fatalf("failed to get rows from merged kvInfo: %v", err)
	}

	// Expect 1 header row + 2 data rows = 3 rows
	if len(rows) != 3 {
		t.Errorf("expected 3 rows in merged kvInfo, got %d", len(rows))
	}
}
