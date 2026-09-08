package excel

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// SheetStyleConfig defines the visual color scheme for an Excel sheet.
type SheetStyleConfig struct {
	SheetName string
	TabColor  string
	HeaderBg  string
}

var SheetStyles = map[string]SheetStyleConfig{
	"kvInfo":        {SheetName: "kvInfo", TabColor: "2ECC71", HeaderBg: "27AE60"},
	"kvCPU":         {SheetName: "kvCPU", TabColor: "3498DB", HeaderBg: "2980B9"},
	"kvMemory":      {SheetName: "kvMemory", TabColor: "2980B9", HeaderBg: "1F618D"},
	"kvDisk":        {SheetName: "kvDisk", TabColor: "F39C12", HeaderBg: "D68910"},
	"kvPartition":   {SheetName: "kvPartition", TabColor: "E67E22", HeaderBg: "CA6F1E"},
	"kvNetwork":     {SheetName: "kvNetwork", TabColor: "1ABC9C", HeaderBg: "16A085"},
	"kvCD":          {SheetName: "kvCD", TabColor: "9B59B6", HeaderBg: "884EA0"},
	"kvSnapshot":    {SheetName: "kvSnapshot", TabColor: "E74C3C", HeaderBg: "C0392B"},
	"kvGuestAgent":  {SheetName: "kvGuestAgent", TabColor: "34495E", HeaderBg: "2C3E50"},
	"kvNode":        {SheetName: "kvNode", TabColor: "2C3E50", HeaderBg: "1A252F"},
	"kvStoragePool": {SheetName: "kvStoragePool", TabColor: "D35400", HeaderBg: "BA4A00"},
	"kvHardware":    {SheetName: "kvHardware", TabColor: "795548", HeaderBg: "5D4037"},
	"kvHealth":      {SheetName: "kvHealth", TabColor: "C0392B", HeaderBg: "922B21"},
}

// CreateHeaderStyle creates a styled header for the given background color.
func CreateHeaderStyle(f *excelize.File, bgHex string) (int, error) {
	return f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Color:  "FFFFFF",
			Size:   11,
			Family: "Segoe UI",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{bgHex},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
			WrapText:   false,
		},
		Border: []excelize.Border{
			{Type: "bottom", Color: "FFFFFF", Style: 1},
		},
	})
}

// CreateDataRowStyle creates alternating row styles (white / light gray).
func CreateDataRowStyle(f *excelize.File, isEven bool) (int, error) {
	bgHex := "FFFFFF"
	if isEven {
		bgHex = "F8F9FA"
	}
	return f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   10,
			Family: "Segoe UI",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{bgHex},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
	})
}

// CreateHealthSeverityStyle creates specific cell styles for CRITICAL/WARNING health rows.
func CreateHealthSeverityStyle(f *excelize.File, severity string) (int, error) {
	bgHex := "FFFFFF"
	textHex := "000000"
	bold := false

	switch severity {
	case "CRITICAL":
		bgHex = "FDEDEC"
		textHex = "C0392B"
		bold = true
	case "WARNING":
		bgHex = "FEF9E7"
		textHex = "B7950B"
		bold = true
	case "INFO":
		bgHex = "EBF5FB"
		textHex = "2980B9"
	}

	return f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   bold,
			Color:  textHex,
			Size:   10,
			Family: "Segoe UI",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{bgHex},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
	})
}

// SetupSheetView configures tab color, freeze panes, and auto-filters.
func SetupSheetView(f *excelize.File, sheetName string, colCount int, rowCount int) error {
	styleCfg, exists := SheetStyles[sheetName]
	if exists && styleCfg.TabColor != "" {
		tabColor := styleCfg.TabColor
		_ = f.SetSheetProps(sheetName, &excelize.SheetPropsOptions{
			TabColorRGB: &tabColor,
		})
	}

	// Freeze top row
	if err := f.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	}); err != nil {
		return err
	}

	// Set row height for header
	_ = f.SetRowHeight(sheetName, 1, 26)

	// Enable auto-filter across all columns
	if colCount > 0 && rowCount > 0 {
		endCol, err := excelize.ColumnNumberToName(colCount)
		if err == nil {
			filterRange := fmt.Sprintf("A1:%s%d", endCol, rowCount)
			_ = f.AutoFilter(sheetName, filterRange, []excelize.AutoFilterOptions{})
		}
	}

	return nil
}

// AutoFitColumnWidths calculates optimal column widths with padding.
func AutoFitColumnWidths(f *excelize.File, sheetName string, colCount int, maxLens []int) {
	for i := 1; i <= colCount; i++ {
		colName, err := excelize.ColumnNumberToName(i)
		if err != nil {
			continue
		}
		width := 12.0
		if i-1 < len(maxLens) && maxLens[i-1] > 0 {
			width = float64(maxLens[i-1]) + 4.0
			if width < 12.0 {
				width = 12.0
			}
			if width > 60.0 {
				width = 60.0
			}
		}
		_ = f.SetColWidth(sheetName, colName, colName, width)
	}
}
