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
			"VM", "Powerstate", "Cluster", "Namespace", "Host",
			"Primary IP Address", "DNS Name", "Guest OS", "Firmware", "EFI Secure boot", "TPM Enabled",
			"CPUs", "CPUs Summary", "CPU Hot Add Max", "Memory (GiB)", "Memory Hot Add Max (GiB)",
			"Disks", "Total Disk capacity (GiB)", "NICs", "Run Strategy", "Cluster rules",
			"Creation date", "Uptime", "Annotation", "Labels", "VM UUID",
		})
		for _, r := range report.Info {
			_ = cw.Write([]string{
				r.VM, r.Powerstate, r.Cluster, r.Namespace, r.Host,
				r.PrimaryIPAddress, r.DNSName, r.GuestOS, r.Firmware, fmt.Sprintf("%t", r.EFISecureBoot), fmt.Sprintf("%t", r.TPMEnabled),
				fmt.Sprintf("%d", r.CPUs), r.CPUsSummary, fmt.Sprintf("%d", r.CPUHotAddMax), fmt.Sprintf("%.2f", r.MemoryGiB), fmt.Sprintf("%.2f", r.MemoryHotAddMaxGiB),
				fmt.Sprintf("%d", r.Disks), fmt.Sprintf("%.2f", r.TotalDiskCapacityGB), fmt.Sprintf("%d", r.NICs), r.RunStrategy, r.ClusterRules,
				r.CreationDate, r.Uptime, r.Annotation, r.Labels, r.VMUUID,
			})
		}
	case "kvCPU", "cpu":
		_ = cw.Write([]string{
			"VM", "Powerstate", "Cluster", "Namespace", "Host",
			"CPUs", "Sockets", "Cores p/s", "Threads",
			"CPU Model", "Dedicated CPU Placement", "NUMA Nodes",
			"CPU Requests", "CPU Limits", "Host Node CPU Model",
		})
		for _, r := range report.CPU {
			_ = cw.Write([]string{
				r.VM, r.Powerstate, r.Cluster, r.Namespace, r.Host,
				fmt.Sprintf("%d", r.CPUs), fmt.Sprintf("%d", r.Sockets), fmt.Sprintf("%d", r.CoresPerSocket), fmt.Sprintf("%d", r.Threads),
				r.CPUModel, fmt.Sprintf("%t", r.DedicatedCPUPlacement), fmt.Sprintf("%d", r.NUMANodes),
				fmt.Sprintf("%.2f", r.CPURequests), fmt.Sprintf("%.2f", r.CPULimits), r.HostNodeCPUModel,
			})
		}
	case "kvMemory", "memory":
		_ = cw.Write([]string{
			"VM", "Powerstate", "Cluster", "Namespace", "Host",
			"Size (GiB)", "Memory Requests (GiB)", "Memory Limits (GiB)",
			"Launcher Overhead (MiB)", "Hugepages", "Ballooned", "Memory Dump Enabled",
		})
		for _, r := range report.Memory {
			_ = cw.Write([]string{
				r.VM, r.Powerstate, r.Cluster, r.Namespace, r.Host,
				fmt.Sprintf("%.2f", r.SizeGiB), fmt.Sprintf("%.2f", r.MemoryRequestsGiB),
				fmt.Sprintf("%.2f", r.MemoryLimitsGiB), fmt.Sprintf("%.2f", r.LauncherOverheadMiB), r.Hugepages,
				fmt.Sprintf("%t", r.AutoattachMemBalloon), fmt.Sprintf("%t", r.MemoryDumpEnabled),
			})
		}
	case "kvDisk", "disk":
		_ = cw.Write([]string{
			"VM", "Powerstate", "Cluster", "Namespace", "Host",
			"Disk", "Volume Type", "Claim Name", "StorageClass",
			"Capacity (GiB)", "Volume Mode", "Access Mode", "Bus Type",
			"Cache Mode", "IO Mode", "Dedicated IO Thread", "CSI Driver",
		})
		for _, r := range report.Disk {
			_ = cw.Write([]string{
				r.VM, r.Powerstate, r.Cluster, r.Namespace, r.Host,
				r.Disk, r.VolumeType, r.ClaimName, r.StorageClass,
				fmt.Sprintf("%.2f", r.CapacityGiB), r.VolumeMode,
				r.AccessMode, r.BusType, r.CacheMode, r.IOMode,
				fmt.Sprintf("%t", r.DedicatedIOThread), r.CSIDriver,
			})
		}
	case "kvPartition", "partition":
		_ = cw.Write([]string{
			"VM", "Powerstate", "Cluster", "Namespace", "Host",
			"Mount Point", "FS Type", "Disk", "Capacity (GiB)",
			"Consumed (GiB)", "Free (GiB)", "Free %",
		})
		for _, r := range report.Partition {
			_ = cw.Write([]string{
				r.VM, r.Powerstate, r.Cluster, r.Namespace, r.Host,
				r.MountPoint, r.FSType, r.Disk,
				fmt.Sprintf("%.2f", r.CapacityGiB), fmt.Sprintf("%.2f", r.ConsumedGiB),
				fmt.Sprintf("%.2f", r.FreeGiB), fmt.Sprintf("%.2f", r.FreePercent),
			})
		}
	case "kvNetwork", "network":
		_ = cw.Write([]string{
			"VM", "Powerstate", "Cluster", "Namespace", "Host",
			"NIC label", "Network", "Binding Type", "Mac Address", "Mac Type",
			"IPv4 Address", "Guest Reported IPs", "Adapter Model", "PCI Address",
		})
		for _, r := range report.Network {
			_ = cw.Write([]string{
				r.VM, r.Powerstate, r.Cluster, r.Namespace, r.Host,
				r.NICLabel, r.Network, r.BindingType, r.MacAddress, r.MacType,
				r.IPv4Address, r.GuestReportedIPs, r.AdapterModel, r.PCIAddress,
			})
		}
	case "kvNode", "node":
		_ = cw.Write([]string{
			"Host", "Cluster", "Status", "Physical Cores", "Total RAM (GiB)",
			"Allocatable CPU", "Allocatable RAM (GiB)", "Allocated VM vCPUs",
			"Allocated VM RAM (GiB)", "vCPU Overcommit Ratio", "Active VM Count",
			"KVM Hardware Acceleration Enabled", "Kubernetes Version", "OS Image", "Kernel Version",
			"Node Taints", "Node Conditions",
		})
		for _, r := range report.Node {
			_ = cw.Write([]string{
				r.Host, r.Cluster, r.Status, fmt.Sprintf("%d", r.PhysicalCores), fmt.Sprintf("%.2f", r.TotalRAMGiB),
				fmt.Sprintf("%.2f", r.AllocatableCPU), fmt.Sprintf("%.2f", r.AllocatableRAMGiB), fmt.Sprintf("%.2f", r.AllocatedVMvCPUs),
				fmt.Sprintf("%.2f", r.AllocatedVMRAMGiB), fmt.Sprintf("%.2f", r.VCPUOvercommitRatio), fmt.Sprintf("%d", r.ActiveVMCount),
				fmt.Sprintf("%t", r.KVMHardwareAccel), r.KubernetesVersion, r.OSImage, r.KernelVersion,
				r.NodeTaints, r.NodeConditions,
			})
		}
	case "kvHealth", "health":
		_ = cw.Write([]string{
			"Rule ID", "Severity", "Category", "Resource Kind",
			"Resource Name", "Cluster", "Namespace", "Issue Summary", "Remediation Recommendation",
		})
		for _, r := range report.Health {
			_ = cw.Write([]string{
				r.RuleID, r.Severity, r.Category, r.ResourceKind,
				r.ResourceName, r.Cluster, r.Namespace, r.IssueSummary, r.Remediation,
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
