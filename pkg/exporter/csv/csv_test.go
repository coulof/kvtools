package csv

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/coulof/kvtools/pkg/engine"
)

func TestExportCSV(t *testing.T) {
	report := &engine.InventoryReport{
		ClusterName: "test-cluster",
		GeneratedAt: time.Now().UTC(),
		Info: []engine.KVInfoRecord{
			{
				VMName:          "vm-alpha",
				Namespace:       "prod",
				PowerState:      "Running",
				RunStrategy:     "Always",
				Node:            "node-1",
				IPAddress:       "10.0.0.1",
				GuestOS:         "Linux",
				MemoryConfigGiB: 4.0,
			},
		},
		Health: []engine.KVHealthRecord{
			{
				RuleID:       "HLTH-001",
				Category:     "Migration Blocker",
				Severity:     "CRITICAL",
				ResourceKind: "VirtualMachine",
				ResourceName: "vm-alpha",
				Namespace:    "prod",
				IssueSummary: "RWO PVC",
				Remediation:  "Fix storage",
			},
		},
	}

	var buf bytes.Buffer
	if err := ExportSheetToCSV(report, "kvInfo", &buf); err != nil {
		t.Fatalf("failed to export kvInfo CSV: %v", err)
	}

	content := buf.String()
	if !strings.Contains(content, "VM Name,Namespace,Power State") || !strings.Contains(content, "vm-alpha,prod,Running") {
		t.Errorf("unexpected CSV content:\n%s", content)
	}

	var healthBuf bytes.Buffer
	if err := ExportSheetToCSV(report, "kvHealth", &healthBuf); err != nil {
		t.Fatalf("failed to export kvHealth CSV: %v", err)
	}

	healthContent := healthBuf.String()
	if !strings.Contains(healthContent, "HLTH-001,Migration Blocker,CRITICAL,VirtualMachine,vm-alpha,prod,RWO PVC,Fix storage") {
		t.Errorf("unexpected health CSV content:\n%s", healthContent)
	}
}
