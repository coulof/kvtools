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
				VM:          "vm-alpha",
				Cluster:     "test-cluster",
				Namespace:   "prod",
				Powerstate:  "poweredOn",
				RunStrategy: "Always",
				Host:        "node-1",
				PrimaryIPAddress: "10.0.0.1",
				GuestOS:     "Linux",
				MemoryGiB:   4.0,
			},
		},
		Health: []engine.KVHealthRecord{
			{
				RuleID:       "HLTH-001",
				Category:     "Migration Blocker",
				Severity:     "CRITICAL",
				ResourceKind: "VirtualMachine",
				ResourceName: "vm-alpha",
				Cluster:      "test-cluster",
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
	if !strings.Contains(content, "VM,Powerstate,Cluster,Namespace") || !strings.Contains(content, "vm-alpha,poweredOn,test-cluster,prod") {
		t.Errorf("unexpected CSV content:\n%s", content)
	}

	var healthBuf bytes.Buffer
	if err := ExportSheetToCSV(report, "kvHealth", &healthBuf); err != nil {
		t.Fatalf("failed to export kvHealth CSV: %v", err)
	}

	healthContent := healthBuf.String()
	if !strings.Contains(healthContent, "HLTH-001,CRITICAL,Migration Blocker,VirtualMachine,vm-alpha,test-cluster,prod,RWO PVC,Fix storage") {
		t.Errorf("unexpected health CSV content:\n%s", healthContent)
	}
}
