package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestHarvesterFixtureReport(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "test", "fixtures", "harvester_report.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read harvester fixture: %v", err)
	}

	var report InventoryReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("failed to unmarshal harvester fixture: %v", err)
	}

	// Verify Cluster & Summary
	if report.ClusterName != "local" {
		t.Errorf("expected clusterName 'local', got '%s'", report.ClusterName)
	}
	if report.Summary.TotalVMs != 1 || report.Summary.TotalNodes != 4 {
		t.Errorf("expected 1 VM and 4 Nodes, got VMs=%d, Nodes=%d", report.Summary.TotalVMs, report.Summary.TotalNodes)
	}

	// Verify VM details
	if len(report.Info) != 1 {
		t.Fatalf("expected 1 VM Info record, got %d", len(report.Info))
	}
	vm := report.Info[0]
	if vm.VMName != "test" || vm.GuestOS != "SUSE Linux Enterprise Server 16.0" || vm.Node != "hv-01" {
		t.Errorf("unexpected VM values: %+v", vm)
	}

	// Verify Partitions
	if len(report.Partition) != 2 {
		t.Fatalf("expected 2 Partition records, got %d", len(report.Partition))
	}
	rootFS := report.Partition[1]
	if rootFS.MountPoint != "/" || rootFS.FSType != "xfs" {
		t.Errorf("unexpected root partition: %+v", rootFS)
	}

	// Verify Health Checks
	if len(report.Health) == 0 {
		t.Fatalf("expected health findings, got 0")
	}
	if report.Health[0].RuleID != "HLTH-009" {
		t.Errorf("expected HLTH-009 health finding, got '%s'", report.Health[0].RuleID)
	}
}
