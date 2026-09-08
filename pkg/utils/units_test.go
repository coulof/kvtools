package utils

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
)

func TestUnitsConversion(t *testing.T) {
	// 10 GiB in bytes
	bytes := int64(10 * GiB)
	if gib := BytesToGiB(bytes); gib != 10.0 {
		t.Errorf("expected 10.0 GiB, got %f", gib)
	}

	// 512 MiB in bytes
	bytesMib := int64(512 * MiB)
	if mib := BytesToMiB(bytesMib); mib != 512.0 {
		t.Errorf("expected 512.0 MiB, got %f", mib)
	}

	// Quantity conversion
	qMem := resource.MustParse("16Gi")
	if gib := QuantityToGiB(&qMem); gib != 16.0 {
		t.Errorf("expected 16.0 GiB, got %f", gib)
	}

	qCPU := resource.MustParse("4000m")
	if cores := QuantityToCores(&qCPU); cores != 4.0 {
		t.Errorf("expected 4.0 cores, got %f", cores)
	}

	qCPUFraction := resource.MustParse("2500m")
	if cores := QuantityToCores(&qCPUFraction); cores != 2.5 {
		t.Errorf("expected 2.5 cores, got %f", cores)
	}
}

func TestFormatDuration(t *testing.T) {
	d1 := 50 * time.Second
	if s := FormatDuration(d1); s != "50s" {
		t.Errorf("expected 50s, got %s", s)
	}

	d2 := 45*time.Minute + 12*time.Second
	if s := FormatDuration(d2); s != "45m 12s" {
		t.Errorf("expected 45m 12s, got %s", s)
	}

	d3 := 3*time.Hour + 25*time.Minute
	if s := FormatDuration(d3); s != "3h 25m" {
		t.Errorf("expected 3h 25m, got %s", s)
	}

	d4 := 10*24*time.Hour + 4*time.Hour + 15*time.Minute
	if s := FormatDuration(d4); s != "10d 4h 15m" {
		t.Errorf("expected 10d 4h 15m, got %s", s)
	}
}

func TestFormatAgeDays(t *testing.T) {
	fifteenDaysAgo := time.Now().Add(-15 * 24 * time.Hour)
	age := FormatAgeDays(fifteenDaysAgo)
	if age != 15 {
		t.Errorf("expected 15 days, got %d", age)
	}
}
