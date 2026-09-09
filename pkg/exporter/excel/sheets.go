package excel

import (
	"fmt"

	"github.com/coulof/kvtools/pkg/engine"
	"github.com/xuri/excelize/v2"
)

// maxLenTracker updates max column string lengths for auto-fitting widths.
func updateMaxLens(maxLens []int, values []interface{}) []int {
	for i, v := range values {
		s := fmt.Sprintf("%v", v)
		l := len(s)
		if i >= len(maxLens) {
			maxLens = append(maxLens, l)
		} else if l > maxLens[i] {
			maxLens[i] = l
		}
	}
	return maxLens
}

func writeHeaders(f *excelize.File, sheet string, headers []string, headerStyle int) []int {
	maxLens := make([]int, len(headers))
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
		_ = f.SetCellStyle(sheet, cell, cell, headerStyle)
		maxLens[i] = len(h)
	}
	return maxLens
}

func writeKVInfoSheet(f *excelize.File, report *engine.InventoryReport, headerStyle, evenStyle, oddStyle int) {
	sheet := "kvInfo"
	f.NewSheet(sheet)

	headers := []string{
		"VM", "Powerstate", "Cluster", "Namespace", "Host",
		"Primary IP Address", "DNS Name", "Guest OS", "Firmware", "EFI Secure boot", "TPM Enabled",
		"CPUs", "CPUs Summary", "CPU Hot Add Max", "Memory (GiB)", "Memory Hot Add Max (GiB)",
		"Disks", "Total Disk capacity (GiB)", "NICs", "Run Strategy", "Cluster rules",
		"Creation date", "Uptime", "Annotation", "Labels", "VM UUID",
	}

	maxLens := writeHeaders(f, sheet, headers, headerStyle)

	for rIdx, r := range report.Info {
		row := rIdx + 2
		rowStyle := oddStyle
		if rIdx%2 == 0 {
			rowStyle = evenStyle
		}

		vals := []interface{}{
			r.VM, r.Powerstate, r.Cluster, r.Namespace, r.Host,
			r.PrimaryIPAddress, r.DNSName, r.GuestOS, r.Firmware, r.EFISecureBoot, r.TPMEnabled,
			r.CPUs, r.CPUsSummary, r.CPUHotAddMax, r.MemoryGiB, r.MemoryHotAddMaxGiB,
			r.Disks, r.TotalDiskCapacityGB, r.NICs, r.RunStrategy, r.ClusterRules,
			r.CreationDate, r.Uptime, r.Annotation, r.Labels, r.VMUUID,
		}
		maxLens = updateMaxLens(maxLens, vals)

		for cIdx, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, row)
			_ = f.SetCellValue(sheet, cell, v)
			_ = f.SetCellStyle(sheet, cell, cell, rowStyle)
		}
	}

	_ = SetupSheetView(f, sheet, len(headers), len(report.Info)+1)
	AutoFitColumnWidths(f, sheet, len(headers), maxLens)
}

func writeKVCPUSheet(f *excelize.File, report *engine.InventoryReport, headerStyle, evenStyle, oddStyle int) {
	sheet := "kvCPU"
	f.NewSheet(sheet)

	headers := []string{
		"VM", "Powerstate", "Cluster", "Namespace", "Host",
		"CPUs", "Sockets", "Cores p/s", "Threads",
		"CPU Model", "Dedicated CPU Placement", "NUMA Nodes",
		"CPU Requests", "CPU Limits", "Host Node CPU Model",
	}

	maxLens := writeHeaders(f, sheet, headers, headerStyle)

	for rIdx, r := range report.CPU {
		row := rIdx + 2
		rowStyle := oddStyle
		if rIdx%2 == 0 {
			rowStyle = evenStyle
		}

		vals := []interface{}{
			r.VM, r.Powerstate, r.Cluster, r.Namespace, r.Host,
			r.CPUs, r.Sockets, r.CoresPerSocket, r.Threads,
			r.CPUModel, r.DedicatedCPUPlacement, r.NUMANodes,
			r.CPURequests, r.CPULimits, r.HostNodeCPUModel,
		}
		maxLens = updateMaxLens(maxLens, vals)

		for cIdx, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, row)
			_ = f.SetCellValue(sheet, cell, v)
			_ = f.SetCellStyle(sheet, cell, cell, rowStyle)
		}
	}

	_ = SetupSheetView(f, sheet, len(headers), len(report.CPU)+1)
	AutoFitColumnWidths(f, sheet, len(headers), maxLens)
}

func writeKVMemorySheet(f *excelize.File, report *engine.InventoryReport, headerStyle, evenStyle, oddStyle int) {
	sheet := "kvMemory"
	f.NewSheet(sheet)

	headers := []string{
		"VM", "Powerstate", "Cluster", "Namespace", "Host",
		"Size (GiB)", "Memory Requests (GiB)", "Memory Limits (GiB)",
		"Launcher Overhead (MiB)", "Hugepages", "Ballooned", "Memory Dump Enabled",
	}

	maxLens := writeHeaders(f, sheet, headers, headerStyle)

	for rIdx, r := range report.Memory {
		row := rIdx + 2
		rowStyle := oddStyle
		if rIdx%2 == 0 {
			rowStyle = evenStyle
		}

		vals := []interface{}{
			r.VM, r.Powerstate, r.Cluster, r.Namespace, r.Host,
			r.SizeGiB, r.MemoryRequestsGiB, r.MemoryLimitsGiB,
			r.LauncherOverheadMiB, r.Hugepages, r.AutoattachMemBalloon, r.MemoryDumpEnabled,
		}
		maxLens = updateMaxLens(maxLens, vals)

		for cIdx, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, row)
			_ = f.SetCellValue(sheet, cell, v)
			_ = f.SetCellStyle(sheet, cell, cell, rowStyle)
		}
	}

	_ = SetupSheetView(f, sheet, len(headers), len(report.Memory)+1)
	AutoFitColumnWidths(f, sheet, len(headers), maxLens)
}

func writeKVDiskSheet(f *excelize.File, report *engine.InventoryReport, headerStyle, evenStyle, oddStyle int) {
	sheet := "kvDisk"
	f.NewSheet(sheet)

	headers := []string{
		"VM", "Powerstate", "Cluster", "Namespace", "Host",
		"Disk", "Volume Type", "Claim Name", "StorageClass",
		"Capacity (GiB)", "Volume Mode", "Access Mode", "Bus Type",
		"Cache Mode", "IO Mode", "Dedicated IO Thread", "CSI Driver",
	}

	maxLens := writeHeaders(f, sheet, headers, headerStyle)

	for rIdx, r := range report.Disk {
		row := rIdx + 2
		rowStyle := oddStyle
		if rIdx%2 == 0 {
			rowStyle = evenStyle
		}

		vals := []interface{}{
			r.VM, r.Powerstate, r.Cluster, r.Namespace, r.Host,
			r.Disk, r.VolumeType, r.ClaimName, r.StorageClass,
			r.CapacityGiB, r.VolumeMode, r.AccessMode, r.BusType,
			r.CacheMode, r.IOMode, r.DedicatedIOThread, r.CSIDriver,
		}
		maxLens = updateMaxLens(maxLens, vals)

		for cIdx, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, row)
			_ = f.SetCellValue(sheet, cell, v)
			_ = f.SetCellStyle(sheet, cell, cell, rowStyle)
		}
	}

	_ = SetupSheetView(f, sheet, len(headers), len(report.Disk)+1)
	AutoFitColumnWidths(f, sheet, len(headers), maxLens)
}

func writeKVPartitionSheet(f *excelize.File, report *engine.InventoryReport, headerStyle, evenStyle, oddStyle int) {
	sheet := "kvPartition"
	f.NewSheet(sheet)

	headers := []string{
		"VM", "Powerstate", "Cluster", "Namespace", "Host",
		"Mount Point", "FS Type", "Disk", "Capacity (GiB)",
		"Consumed (GiB)", "Free (GiB)", "Free %",
	}

	maxLens := writeHeaders(f, sheet, headers, headerStyle)

	for rIdx, r := range report.Partition {
		row := rIdx + 2
		rowStyle := oddStyle
		if rIdx%2 == 0 {
			rowStyle = evenStyle
		}

		vals := []interface{}{
			r.VM, r.Powerstate, r.Cluster, r.Namespace, r.Host,
			r.MountPoint, r.FSType, r.Disk, r.CapacityGiB,
			r.ConsumedGiB, r.FreeGiB, r.FreePercent,
		}
		maxLens = updateMaxLens(maxLens, vals)

		for cIdx, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, row)
			_ = f.SetCellValue(sheet, cell, v)
			_ = f.SetCellStyle(sheet, cell, cell, rowStyle)
		}
	}

	_ = SetupSheetView(f, sheet, len(headers), len(report.Partition)+1)
	AutoFitColumnWidths(f, sheet, len(headers), maxLens)
}

func writeKVNetworkSheet(f *excelize.File, report *engine.InventoryReport, headerStyle, evenStyle, oddStyle int) {
	sheet := "kvNetwork"
	f.NewSheet(sheet)

	headers := []string{
		"VM", "Powerstate", "Cluster", "Namespace", "Host",
		"NIC label", "Network", "Binding Type", "Mac Address", "Mac Type",
		"IPv4 Address", "Guest Reported IPs", "Adapter Model", "PCI Address",
	}

	maxLens := writeHeaders(f, sheet, headers, headerStyle)

	for rIdx, r := range report.Network {
		row := rIdx + 2
		rowStyle := oddStyle
		if rIdx%2 == 0 {
			rowStyle = evenStyle
		}

		vals := []interface{}{
			r.VM, r.Powerstate, r.Cluster, r.Namespace, r.Host,
			r.NICLabel, r.Network, r.BindingType, r.MacAddress, r.MacType,
			r.IPv4Address, r.GuestReportedIPs, r.AdapterModel, r.PCIAddress,
		}
		maxLens = updateMaxLens(maxLens, vals)

		for cIdx, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, row)
			_ = f.SetCellValue(sheet, cell, v)
			_ = f.SetCellStyle(sheet, cell, cell, rowStyle)
		}
	}

	_ = SetupSheetView(f, sheet, len(headers), len(report.Network)+1)
	AutoFitColumnWidths(f, sheet, len(headers), maxLens)
}

func writeKVCDSheet(f *excelize.File, report *engine.InventoryReport, headerStyle, evenStyle, oddStyle int) {
	sheet := "kvCD"
	f.NewSheet(sheet)

	headers := []string{
		"VM", "Powerstate", "Cluster", "Namespace", "Host",
		"Device Node", "Source Type", "Source Image", "Boot Order", "Connected",
	}

	maxLens := writeHeaders(f, sheet, headers, headerStyle)

	for rIdx, r := range report.CD {
		row := rIdx + 2
		rowStyle := oddStyle
		if rIdx%2 == 0 {
			rowStyle = evenStyle
		}

		vals := []interface{}{
			r.VM, r.Powerstate, r.Cluster, r.Namespace, r.Host,
			r.DeviceNode, r.SourceType, r.SourceImage, r.BootOrder, r.Connected,
		}
		maxLens = updateMaxLens(maxLens, vals)

		for cIdx, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, row)
			_ = f.SetCellValue(sheet, cell, v)
			_ = f.SetCellStyle(sheet, cell, cell, rowStyle)
		}
	}

	_ = SetupSheetView(f, sheet, len(headers), len(report.CD)+1)
	AutoFitColumnWidths(f, sheet, len(headers), maxLens)
}

func writeKVSnapshotSheet(f *excelize.File, report *engine.InventoryReport, headerStyle, evenStyle, oddStyle int) {
	sheet := "kvSnapshot"
	f.NewSheet(sheet)

	headers := []string{
		"Snapshot Name", "Cluster", "Namespace", "Source VM", "Ready to Use",
		"Date / time", "Age (Days)", "Volume Snapshot Count",
		"Total Restorable Size (GiB)", "Error Reason",
	}

	maxLens := writeHeaders(f, sheet, headers, headerStyle)

	for rIdx, r := range report.Snapshot {
		row := rIdx + 2
		rowStyle := oddStyle
		if rIdx%2 == 0 {
			rowStyle = evenStyle
		}

		vals := []interface{}{
			r.SnapshotName, r.Cluster, r.Namespace, r.SourceVM, r.ReadyToUse,
			r.CreationDate, r.AgeDays, r.VolumeSnapshotCount,
			r.TotalRestorableSizeGiB, r.ErrorReason,
		}
		maxLens = updateMaxLens(maxLens, vals)

		for cIdx, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, row)
			_ = f.SetCellValue(sheet, cell, v)
			_ = f.SetCellStyle(sheet, cell, cell, rowStyle)
		}
	}

	_ = SetupSheetView(f, sheet, len(headers), len(report.Snapshot)+1)
	AutoFitColumnWidths(f, sheet, len(headers), maxLens)
}

func writeKVGuestAgentSheet(f *excelize.File, report *engine.InventoryReport, headerStyle, evenStyle, oddStyle int) {
	sheet := "kvGuestAgent"
	f.NewSheet(sheet)

	headers := []string{
		"VM", "Powerstate", "Cluster", "Namespace", "Host",
		"Agent Connected", "Agent Version", "Guest Hostname", "Guest OS",
		"Kernel Release", "Timezone", "FS Freeze Supported", "Logged-in Users Count",
	}

	maxLens := writeHeaders(f, sheet, headers, headerStyle)

	for rIdx, r := range report.GuestAgent {
		row := rIdx + 2
		rowStyle := oddStyle
		if rIdx%2 == 0 {
			rowStyle = evenStyle
		}

		vals := []interface{}{
			r.VM, r.Powerstate, r.Cluster, r.Namespace, r.Host,
			r.AgentConnected, r.AgentVersion, r.GuestHostname, r.GuestOS,
			r.KernelRelease, r.Timezone, r.FSFreezeSupported, r.LoggedInUsersCount,
		}
		maxLens = updateMaxLens(maxLens, vals)

		for cIdx, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, row)
			_ = f.SetCellValue(sheet, cell, v)
			_ = f.SetCellStyle(sheet, cell, cell, rowStyle)
		}
	}

	_ = SetupSheetView(f, sheet, len(headers), len(report.GuestAgent)+1)
	AutoFitColumnWidths(f, sheet, len(headers), maxLens)
}

func writeKVNodeSheet(f *excelize.File, report *engine.InventoryReport, headerStyle, evenStyle, oddStyle int) {
	sheet := "kvNode"
	f.NewSheet(sheet)

	headers := []string{
		"Host", "Cluster", "Status", "# Cores", "Total RAM (GiB)",
		"Allocatable CPU", "Allocatable RAM (GiB)", "VM vCPUs",
		"VM RAM (GiB)", "vCPU Overcommit Ratio", "# VMs",
		"KVM Hardware Accel", "Kubernetes Version", "OS Image", "Kernel Version",
		"Node Taints", "Node Conditions",
	}

	maxLens := writeHeaders(f, sheet, headers, headerStyle)

	for rIdx, r := range report.Node {
		row := rIdx + 2
		rowStyle := oddStyle
		if rIdx%2 == 0 {
			rowStyle = evenStyle
		}

		vals := []interface{}{
			r.Host, r.Cluster, r.Status, r.PhysicalCores, r.TotalRAMGiB,
			r.AllocatableCPU, r.AllocatableRAMGiB, r.AllocatedVMvCPUs,
			r.AllocatedVMRAMGiB, r.VCPUOvercommitRatio, r.ActiveVMCount,
			r.KVMHardwareAccel, r.KubernetesVersion, r.OSImage, r.KernelVersion,
			r.NodeTaints, r.NodeConditions,
		}
		maxLens = updateMaxLens(maxLens, vals)

		for cIdx, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, row)
			_ = f.SetCellValue(sheet, cell, v)
			_ = f.SetCellStyle(sheet, cell, cell, rowStyle)
		}
	}

	_ = SetupSheetView(f, sheet, len(headers), len(report.Node)+1)
	AutoFitColumnWidths(f, sheet, len(headers), maxLens)
}

func writeKVStoragePoolSheet(f *excelize.File, report *engine.InventoryReport, headerStyle, evenStyle, oddStyle int) {
	sheet := "kvStoragePool"
	f.NewSheet(sheet)

	headers := []string{
		"StorageClass Name", "Cluster", "Provisioner / CSI Driver", "Reclaim Policy",
		"VolumeBindingMode", "AllowVolumeExpansion", "Is Default Class",
		"Total Bound PVC Count", "Total Allocated Capacity (GiB)",
	}

	maxLens := writeHeaders(f, sheet, headers, headerStyle)

	for rIdx, r := range report.StoragePool {
		row := rIdx + 2
		rowStyle := oddStyle
		if rIdx%2 == 0 {
			rowStyle = evenStyle
		}

		vals := []interface{}{
			r.StorageClassName, r.Cluster, r.ProvisionerCSIDriver, r.ReclaimPolicy,
			r.VolumeBindingMode, r.AllowVolumeExpansion, r.IsDefaultClass,
			r.TotalBoundPVCCount, r.TotalAllocatedSizeGiB,
		}
		maxLens = updateMaxLens(maxLens, vals)

		for cIdx, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, row)
			_ = f.SetCellValue(sheet, cell, v)
			_ = f.SetCellStyle(sheet, cell, cell, rowStyle)
		}
	}

	_ = SetupSheetView(f, sheet, len(headers), len(report.StoragePool)+1)
	AutoFitColumnWidths(f, sheet, len(headers), maxLens)
}

func writeKVHardwareSheet(f *excelize.File, report *engine.InventoryReport, headerStyle, evenStyle, oddStyle int) {
	sheet := "kvHardware"
	f.NewSheet(sheet)

	headers := []string{
		"VM", "Cluster", "Namespace", "Host", "Device Type", "Device Name", "Resource Name",
	}

	maxLens := writeHeaders(f, sheet, headers, headerStyle)

	for rIdx, r := range report.Hardware {
		row := rIdx + 2
		rowStyle := oddStyle
		if rIdx%2 == 0 {
			rowStyle = evenStyle
		}

		vals := []interface{}{
			r.VM, r.Cluster, r.Namespace, r.Host, r.DeviceType, r.DeviceName, r.ResourceName,
		}
		maxLens = updateMaxLens(maxLens, vals)

		for cIdx, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, row)
			_ = f.SetCellValue(sheet, cell, v)
			_ = f.SetCellStyle(sheet, cell, cell, rowStyle)
		}
	}

	_ = SetupSheetView(f, sheet, len(headers), len(report.Hardware)+1)
	AutoFitColumnWidths(f, sheet, len(headers), maxLens)
}

func writeKVHealthSheet(f *excelize.File, report *engine.InventoryReport, headerStyle, critStyle, warnStyle, infoStyle int) {
	sheet := "kvHealth"
	f.NewSheet(sheet)

	headers := []string{
		"Rule ID", "Severity", "Category", "Resource Kind",
		"Resource Name", "Cluster", "Namespace", "Issue Summary", "Remediation Recommendation",
	}

	maxLens := writeHeaders(f, sheet, headers, headerStyle)

	for rIdx, r := range report.Health {
		row := rIdx + 2
		rowStyle := infoStyle
		if r.Severity == "CRITICAL" {
			rowStyle = critStyle
		} else if r.Severity == "WARNING" {
			rowStyle = warnStyle
		}

		vals := []interface{}{
			r.RuleID, r.Severity, r.Category, r.ResourceKind,
			r.ResourceName, r.Cluster, r.Namespace, r.IssueSummary, r.Remediation,
		}
		maxLens = updateMaxLens(maxLens, vals)

		for cIdx, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, row)
			_ = f.SetCellValue(sheet, cell, v)
			_ = f.SetCellStyle(sheet, cell, cell, rowStyle)
		}
	}

	_ = SetupSheetView(f, sheet, len(headers), len(report.Health)+1)
	AutoFitColumnWidths(f, sheet, len(headers), maxLens)
}
