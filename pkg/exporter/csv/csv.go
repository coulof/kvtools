package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/coulof/kvtools/pkg/engine"
)

// ExportSheetToCSV writes a specific sheet from the report to the given writer.
func ExportSheetToCSV(report *engine.InventoryReport, sheetName string, w io.Writer) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	switch sheetName {
	case "kvInfo", "info":
		_ = cw.Write([]string{
			"VM Name", "Namespace", "Power State", "Run Strategy", "Node",
			"IP Address", "Guest OS", "Firmware / Boot", "TPM Enabled",
			"CPUs (C/S/T)", "Memory Configured (GiB)", "Disks Count", "NICs Count",
			"Created Time", "Uptime", "Labels", "Annotations", "UID",
		})
		for _, r := range report.Info {
			_ = cw.Write([]string{
				r.VMName, r.Namespace, r.PowerState, r.RunStrategy, r.Node,
				r.IPAddress, r.GuestOS, r.FirmwareBoot, fmt.Sprintf("%t", r.TPMEnabled),
				r.CPUsSummary, fmt.Sprintf("%.2f", r.MemoryConfigGiB), fmt.Sprintf("%d", r.DisksCount), fmt.Sprintf("%d", r.NICsCount),
				r.CreatedTime, r.Uptime, r.Labels, r.Annotations, r.UID,
			})
		}
	case "kvCPU", "cpu":
		_ = cw.Write([]string{
			"VM Name", "Namespace", "Cores", "Sockets", "Threads",
			"Total vCPUs", "CPU Model", "Dedicated CPU Placement", "NUMA Nodes",
			"CPU Requests", "CPU Limits", "Host Node CPU Model",
		})
		for _, r := range report.CPU {
			_ = cw.Write([]string{
				r.VMName, r.Namespace, fmt.Sprintf("%d", r.Cores), fmt.Sprintf("%d", r.Sockets), fmt.Sprintf("%d", r.Threads),
				fmt.Sprintf("%d", r.TotalVCPUs), r.CPUModel, fmt.Sprintf("%t", r.DedicatedCPUPlacement), fmt.Sprintf("%d", r.NUMANodes),
				fmt.Sprintf("%.2f", r.CPURequests), fmt.Sprintf("%.2f", r.CPULimits), r.HostNodeCPUModel,
			})
		}
	case "kvMemory", "memory":
		_ = cw.Write([]string{
			"VM Name", "Namespace", "Guest RAM (GiB)", "Memory Requests (GiB)",
			"Memory Limits (GiB)", "Launcher Overhead (MiB)", "Hugepages",
			"Autoattach Memory Balloon", "Memory Dump Enabled",
		})
		for _, r := range report.Memory {
			_ = cw.Write([]string{
				r.VMName, r.Namespace, fmt.Sprintf("%.2f", r.GuestRAMGiB), fmt.Sprintf("%.2f", r.MemoryRequestsGiB),
				fmt.Sprintf("%.2f", r.MemoryLimitsGiB), fmt.Sprintf("%.2f", r.LauncherOverheadMiB), r.Hugepages,
				fmt.Sprintf("%t", r.AutoattachMemBalloon), fmt.Sprintf("%t", r.MemoryDumpEnabled),
			})
		}
	case "kvDisk", "disk":
		_ = cw.Write([]string{
			"VM Name", "Namespace", "Disk Target Name", "Volume Type",
			"Claim Name", "StorageClass", "Provisioned Size (GiB)", "Volume Mode",
			"Access Mode", "Bus Type", "Cache Mode", "IO Mode",
			"Dedicated IO Thread", "CSI Driver",
		})
		for _, r := range report.Disk {
			_ = cw.Write([]string{
				r.VMName, r.Namespace, r.DiskTargetName, r.VolumeType,
				r.ClaimName, r.StorageClass, fmt.Sprintf("%.2f", r.ProvisionedSizeGiB), r.VolumeMode,
				r.AccessMode, r.BusType, r.CacheMode, r.IOMode,
				fmt.Sprintf("%t", r.DedicatedIOThread), r.CSIDriver,
			})
		}
	case "kvNode", "node":
		_ = cw.Write([]string{
			"Node Name", "Status", "Total Physical Cores", "Total RAM (GiB)",
			"Allocatable CPU", "Allocatable RAM (GiB)", "Allocated VM vCPUs",
			"Allocated VM RAM (GiB)", "vCPU Overcommit Ratio", "Active VM Count",
			"KVM Hardware Acceleration Enabled", "Kubernetes Version", "OS Image", "Kernel Version",
		})
		for _, r := range report.Node {
			_ = cw.Write([]string{
				r.NodeName, r.Status, fmt.Sprintf("%d", r.TotalPhysicalCores), fmt.Sprintf("%.2f", r.TotalRAMGiB),
				fmt.Sprintf("%.2f", r.AllocatableCPU), fmt.Sprintf("%.2f", r.AllocatableRAMGiB), fmt.Sprintf("%.2f", r.AllocatedVMvCPUs),
				fmt.Sprintf("%.2f", r.AllocatedVMRAMGiB), fmt.Sprintf("%.2f", r.VCPUOvercommitRatio), fmt.Sprintf("%d", r.ActiveVMCount),
				fmt.Sprintf("%t", r.KVMHardwareAccel), r.KubernetesVersion, r.OSImage, r.KernelVersion,
			})
		}
	case "kvHealth", "health":
		_ = cw.Write([]string{
			"Rule ID", "Category", "Severity", "Resource Kind",
			"Resource Name", "Namespace", "Issue Summary", "Remediation Recommendation",
		})
		for _, r := range report.Health {
			_ = cw.Write([]string{
				r.RuleID, r.Category, r.Severity, r.ResourceKind,
				r.ResourceName, r.Namespace, r.IssueSummary, r.Remediation,
			})
		}
	default:
		// Default to kvInfo
		return ExportSheetToCSV(report, "kvInfo", w)
	}

	return nil
}

// ExportAllSheetsToCSV writes each sheet to a separate CSV file in outDir.
func ExportAllSheetsToCSV(report *engine.InventoryReport, outDir string) ([]string, error) {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", outDir, err)
	}

	sheetNames := []string{
		"kvInfo", "kvCPU", "kvMemory", "kvDisk", "kvPartition",
		"kvNetwork", "kvCD", "kvSnapshot", "kvGuestAgent",
		"kvNode", "kvStoragePool", "kvHardware", "kvHealth",
	}

	var generatedFiles []string
	for _, s := range sheetNames {
		filePath := filepath.Join(outDir, fmt.Sprintf("%s.csv", s))
		f, err := os.Create(filePath)
		if err != nil {
			return generatedFiles, fmt.Errorf("failed to create file %s: %w", filePath, err)
		}
		err = ExportSheetToCSV(report, s, f)
		_ = f.Close()
		if err != nil {
			return generatedFiles, err
		}
		generatedFiles = append(generatedFiles, filePath)
	}

	return generatedFiles, nil
}
