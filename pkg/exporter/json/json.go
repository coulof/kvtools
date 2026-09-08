package json

import (
	"encoding/json"
	"io"

	"github.com/coulof/kvtools/pkg/engine"
)

// ExportJSON marshals an InventoryReport into formatted JSON and writes it to w.
func ExportJSON(report *engine.InventoryReport, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}
