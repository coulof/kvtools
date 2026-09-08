package json

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/coulof/kvtools/pkg/engine"
)

func TestExportJSON(t *testing.T) {
	report := &engine.InventoryReport{
		ClusterName: "test-cluster",
		GeneratedAt: time.Now().UTC(),
		Info: []engine.KVInfoRecord{
			{
				VMName:     "vm-1",
				Namespace:  "default",
				PowerState: "Running",
			},
		},
		Summary: engine.ReportSummary{
			TotalVMs:   1,
			RunningVMs: 1,
		},
	}

	var buf bytes.Buffer
	if err := ExportJSON(report, &buf); err != nil {
		t.Fatalf("failed to export JSON: %v", err)
	}

	var parsed engine.InventoryReport
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal exported JSON: %v", err)
	}

	if parsed.ClusterName != "test-cluster" || len(parsed.Info) != 1 || parsed.Info[0].VMName != "vm-1" {
		t.Errorf("unexpected unmarshaled JSON: %+v", parsed)
	}
}
