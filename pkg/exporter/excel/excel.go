package excel

import (
	"fmt"

	"github.com/coulof/kvtools/pkg/engine"
	"github.com/xuri/excelize/v2"
)

// ExportExcel generates a styled 13-tab Excel spreadsheet from an InventoryReport and writes to filename.
func ExportExcel(report *engine.InventoryReport, filename string) error {
	f := excelize.NewFile()
	defer func() {
		_ = f.Close()
	}()

	evenStyle, err := CreateDataRowStyle(f, true)
	if err != nil {
		return fmt.Errorf("failed to create even row style: %w", err)
	}

	oddStyle, err := CreateDataRowStyle(f, false)
	if err != nil {
		return fmt.Errorf("failed to create odd row style: %w", err)
	}

	critStyle, err := CreateHealthSeverityStyle(f, "CRITICAL")
	if err != nil {
		return fmt.Errorf("failed to create critical style: %w", err)
	}

	warnStyle, err := CreateHealthSeverityStyle(f, "WARNING")
	if err != nil {
		return fmt.Errorf("failed to create warning style: %w", err)
	}

	infoStyle, err := CreateHealthSeverityStyle(f, "INFO")
	if err != nil {
		return fmt.Errorf("failed to create info style: %w", err)
	}

	// 1. kvInfo
	infoHStyle, _ := CreateHeaderStyle(f, SheetStyles["kvInfo"].HeaderBg)
	writeKVInfoSheet(f, report, infoHStyle, evenStyle, oddStyle)

	// 2. kvCPU
	cpuHStyle, _ := CreateHeaderStyle(f, SheetStyles["kvCPU"].HeaderBg)
	writeKVCPUSheet(f, report, cpuHStyle, evenStyle, oddStyle)

	// 3. kvMemory
	memHStyle, _ := CreateHeaderStyle(f, SheetStyles["kvMemory"].HeaderBg)
	writeKVMemorySheet(f, report, memHStyle, evenStyle, oddStyle)

	// 4. kvDisk
	diskHStyle, _ := CreateHeaderStyle(f, SheetStyles["kvDisk"].HeaderBg)
	writeKVDiskSheet(f, report, diskHStyle, evenStyle, oddStyle)

	// 5. kvPartition
	partHStyle, _ := CreateHeaderStyle(f, SheetStyles["kvPartition"].HeaderBg)
	writeKVPartitionSheet(f, report, partHStyle, evenStyle, oddStyle)

	// 6. kvNetwork
	netHStyle, _ := CreateHeaderStyle(f, SheetStyles["kvNetwork"].HeaderBg)
	writeKVNetworkSheet(f, report, netHStyle, evenStyle, oddStyle)

	// 7. kvCD
	cdHStyle, _ := CreateHeaderStyle(f, SheetStyles["kvCD"].HeaderBg)
	writeKVCDSheet(f, report, cdHStyle, evenStyle, oddStyle)

	// 8. kvSnapshot
	snapHStyle, _ := CreateHeaderStyle(f, SheetStyles["kvSnapshot"].HeaderBg)
	writeKVSnapshotSheet(f, report, snapHStyle, evenStyle, oddStyle)

	// 9. kvGuestAgent
	gaHStyle, _ := CreateHeaderStyle(f, SheetStyles["kvGuestAgent"].HeaderBg)
	writeKVGuestAgentSheet(f, report, gaHStyle, evenStyle, oddStyle)

	// 10. kvNode
	nodeHStyle, _ := CreateHeaderStyle(f, SheetStyles["kvNode"].HeaderBg)
	writeKVNodeSheet(f, report, nodeHStyle, evenStyle, oddStyle)

	// 11. kvStoragePool
	spHStyle, _ := CreateHeaderStyle(f, SheetStyles["kvStoragePool"].HeaderBg)
	writeKVStoragePoolSheet(f, report, spHStyle, evenStyle, oddStyle)

	// 12. kvHardware
	hwHStyle, _ := CreateHeaderStyle(f, SheetStyles["kvHardware"].HeaderBg)
	writeKVHardwareSheet(f, report, hwHStyle, evenStyle, oddStyle)

	// 13. kvHealth
	hlthHStyle, _ := CreateHeaderStyle(f, SheetStyles["kvHealth"].HeaderBg)
	writeKVHealthSheet(f, report, hlthHStyle, critStyle, warnStyle, infoStyle)

	// Delete default sheet created by excelize
	_ = f.DeleteSheet("Sheet1")

	// Set active sheet to kvInfo
	infoIdx, err := f.GetSheetIndex("kvInfo")
	if err == nil && infoIdx >= 0 {
		f.SetActiveSheet(infoIdx)
	}

	if err := f.SaveAs(filename); err != nil {
		return fmt.Errorf("failed to save Excel file to %s: %w", filename, err)
	}

	return nil
}
