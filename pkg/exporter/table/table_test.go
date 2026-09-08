package table

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/coulof/kvtools/pkg/engine"
)

func TestRenderTables(t *testing.T) {
	report := &engine.InventoryReport{
		ClusterName: "prod-cluster",
		GeneratedAt: time.Now().UTC(),
		Summary: engine.ReportSummary{
			TotalVMs:          2,
			RunningVMs:        1,
			StoppedVMs:        1,
			TotalNodes:        1,
			ReadyNodes:        1,
			TotalDisks:        2,
			TotalSnapshots:    0,
			HealthIssuesCount: 1,
			CriticalIssues:    1,
		},
		Info: []engine.KVInfoRecord{
			{
				VMName:          "web-vm",
				Namespace:       "default",
				PowerState:      "Running",
				Node:            "node-1",
				IPAddress:       "10.0.0.5",
				GuestOS:         "Linux",
				CPUsSummary:     "2 Cores, 1 Socket (2 vCPUs)",
				MemoryConfigGiB: 4.0,
				DisksCount:      1,
				NICsCount:       1,
				Uptime:          "1d 2h",
			},
		},
		Node: []engine.KVNodeRecord{
			{
				NodeName:            "node-1",
				Status:              "Ready",
				TotalPhysicalCores:  16,
				TotalRAMGiB:         64.0,
				AllocatableCPU:      15.0,
				AllocatableRAMGiB:   60.0,
				AllocatedVMvCPUs:    2.0,
				AllocatedVMRAMGiB:   4.0,
				VCPUOvercommitRatio: 0.13,
				ActiveVMCount:       1,
				KVMHardwareAccel:    true,
			},
		},
		Health: []engine.KVHealthRecord{
			{
				RuleID:       "HLTH-001",
				Category:     "Migration Blocker",
				Severity:     "CRITICAL",
				ResourceKind: "VirtualMachine",
				ResourceName: "web-vm",
				Namespace:    "default",
				IssueSummary: "RWO PVC attached with live migration requested",
				Remediation:  "Switch to RWX or CDI block volume migration",
			},
		},
	}

	var buf bytes.Buffer
	RenderTableSheet(report, "kvInfo", &buf)
	output := buf.String()

	if !strings.Contains(output, "KubeVirt Cluster Virtualization Summary") {
		t.Errorf("expected summary banner in output, got:\n%s", output)
	}
	if !strings.Contains(output, "web-vm") {
		t.Errorf("expected web-vm in output, got:\n%s", output)
	}
	if !strings.Contains(output, "HLTH-001") {
		t.Errorf("expected HLTH-001 in output, got:\n%s", output)
	}

	// Test Health table render
	var healthBuf bytes.Buffer
	RenderTableSheet(report, "kvHealth", &healthBuf)
	healthOut := healthBuf.String()
	if !strings.Contains(healthOut, "HLTH-001") || !strings.Contains(healthOut, "Migration Blocker") {
		t.Errorf("expected health table details, got:\n%s", healthOut)
	}

	// Test Node table render
	var nodeBuf bytes.Buffer
	RenderTableSheet(report, "kvNode", &nodeBuf)
	nodeOut := nodeBuf.String()
	if !strings.Contains(nodeOut, "node-1") {
		t.Errorf("expected node-1 in node table, got:\n%s", nodeOut)
	}
}
