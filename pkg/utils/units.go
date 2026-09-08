package utils

import (
	"fmt"
	"math"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	Byte = 1
	KiB  = 1024 * Byte
	MiB  = 1024 * KiB
	GiB  = 1024 * MiB
	TiB  = 1024 * GiB
)

// BytesToGiB converts byte count to GiB float rounded to 2 decimal places.
func BytesToGiB(bytes int64) float64 {
	if bytes <= 0 {
		return 0
	}
	gib := float64(bytes) / float64(GiB)
	return math.Round(gib*100) / 100
}

// BytesToMiB converts byte count to MiB float rounded to 2 decimal places.
func BytesToMiB(bytes int64) float64 {
	if bytes <= 0 {
		return 0
	}
	mib := float64(bytes) / float64(MiB)
	return math.Round(mib*100) / 100
}

// QuantityToGiB converts a k8s resource.Quantity to GiB float.
func QuantityToGiB(q *resource.Quantity) float64 {
	if q == nil || q.IsZero() {
		return 0
	}
	return BytesToGiB(q.Value())
}

// QuantityToMiB converts a k8s resource.Quantity to MiB float.
func QuantityToMiB(q *resource.Quantity) float64 {
	if q == nil || q.IsZero() {
		return 0
	}
	return BytesToMiB(q.Value())
}

// QuantityToCores converts a k8s CPU resource.Quantity to CPU cores float.
func QuantityToCores(q *resource.Quantity) float64 {
	if q == nil || q.IsZero() {
		return 0
	}
	cores := float64(q.MilliValue()) / 1000.0
	return math.Round(cores*100) / 100
}

// FormatDuration formats duration in a friendly format like "14d 6h 32m" or "45m".
func FormatDuration(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// FormatUptime calculates uptime from a start time to now.
func FormatUptime(startTime *time.Time) string {
	if startTime == nil || startTime.IsZero() {
		return "N/A"
	}
	return FormatDuration(time.Since(*startTime))
}

// FormatAgeDays returns the age in days as an integer.
func FormatAgeDays(t time.Time) int {
	if t.IsZero() {
		return 0
	}
	return int(time.Since(t).Hours() / 24)
}

// FormatTimestamp returns an RFC3339 formatted timestamp string or "N/A".
func FormatTimestamp(t *time.Time) string {
	if t == nil || t.IsZero() {
		return "N/A"
	}
	return t.UTC().Format(time.RFC3339)
}
