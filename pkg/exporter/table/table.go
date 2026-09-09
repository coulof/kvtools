package table

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/coulof/kvtools/pkg/engine"
	"github.com/olekukonko/tablewriter"
)

// RenderSummary renders an executive summary banner to the writer.
func RenderSummary(report *engine.InventoryReport, w io.Writer) {
	fmt.Fprintf(w, "\n\033[1;36m=== KubeVirt Cluster Virtualization Summary ===\033[0m\n")
	fmt.Fprintf(w, "Cluster Name:       \033[1m%s\033[0m\n", report.ClusterName)
	fmt.Fprintf(w, "Generated At:       %s\n", report.GeneratedAt.Format("2006-01-02 15:04:05 UTC"))
	fmt.Fprintf(w, "Total VMs:          \033[1m%d\033[0m (\033[32m%d running\033[0m, \033[90m%d stopped\033[0m)\n",
		report.Summary.TotalVMs, report.Summary.RunningVMs, report.Summary.StoppedVMs)
	fmt.Fprintf(w, "Compute Nodes:      \033[1m%d\033[0m (\033[32m%d ready\033[0m)\n",
		report.Summary.TotalNodes, report.Summary.ReadyNodes)
	fmt.Fprintf(w, "Virtual Disks:      %d\n", report.Summary.TotalDisks)
	fmt.Fprintf(w, "VM Snapshots:       %d\n", report.Summary.TotalSnapshots)

	if report.Summary.HealthIssuesCount > 0 {
		fmt.Fprintf(w, "Health Findings:    \033[31m%d Critical\033[0m, \033[33m%d Warning\033[0m (Total: %d)\n",
			report.Summary.CriticalIssues, report.Summary.WarningIssues, report.Summary.HealthIssuesCount)
	} else {
		fmt.Fprintf(w, "Health Findings:    \033[32m0 issues detected\033[0m\n")
	}
	fmt.Fprintf(w, "\n")
}

// RenderHealthTable renders the kvHealth audit report in a formatted terminal table.
func RenderHealthTable(records []engine.KVHealthRecord, w io.Writer) {
	if len(records) == 0 {
		fmt.Fprintf(w, "\033[32m✔ No health, hygiene, or migration issues detected across the cluster.\033[0m\n")
		return
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Rule ID", "Severity", "Category", "Resource", "Cluster", "Namespace", "Issue Summary", "Remediation"})
	table.SetBorder(true)
	table.SetRowLine(true)
	table.SetAutoWrapText(true)
	table.SetColWidth(40)

	for _, r := range records {
		sevFormatted := r.Severity
		if r.Severity == "CRITICAL" {
			sevFormatted = fmt.Sprintf("\033[31;1m%s\033[0m", r.Severity)
		} else if r.Severity == "WARNING" {
			sevFormatted = fmt.Sprintf("\033[33;1m%s\033[0m", r.Severity)
		}

		table.Append([]string{
			r.RuleID,
			sevFormatted,
			r.Category,
			fmt.Sprintf("%s/%s", r.ResourceKind, r.ResourceName),
			r.Cluster,
			r.Namespace,
			r.IssueSummary,
			r.Remediation,
		})
	}

	table.Render()
}

// RenderInfoTable renders the kvInfo sheet in the terminal.
func RenderInfoTable(records []engine.KVInfoRecord, w io.Writer) {
	if len(records) == 0 {
		fmt.Fprintf(w, "No VirtualMachines found.\n")
		return
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"VM", "Powerstate", "Cluster", "Namespace", "Host", "IP Address", "Guest OS", "vCPUs", "RAM (GiB)", "Disks", "NICs", "Uptime"})
	table.SetBorder(true)
	table.SetAutoWrapText(false)

	for _, r := range records {
		state := r.Powerstate
		if r.Powerstate == "poweredOn" || r.Powerstate == "Running" {
			state = fmt.Sprintf("\033[32m%s\033[0m", r.Powerstate)
		} else if r.Powerstate == "poweredOff" || r.Powerstate == "Stopped" {
			state = fmt.Sprintf("\033[90m%s\033[0m", r.Powerstate)
		}

		table.Append([]string{
			r.VM,
			state,
			r.Cluster,
			r.Namespace,
			r.Host,
			r.PrimaryIPAddress,
			r.GuestOS,
			r.CPUsSummary,
			fmt.Sprintf("%.1f", r.MemoryGiB),
			fmt.Sprintf("%d", r.Disks),
			fmt.Sprintf("%d", r.NICs),
			r.Uptime,
		})
	}

	table.Render()
}

// RenderCPUTable renders the kvCPU sheet in the terminal.
func RenderCPUTable(records []engine.KVCPURecord, w io.Writer) {
	if len(records) == 0 {
		fmt.Fprintf(w, "No CPU records found.\n")
		return
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"VM", "Powerstate", "Cluster", "Namespace", "Host", "CPUs", "Sockets", "Cores p/s", "Threads", "CPU Model", "Dedicated", "Req Cores", "Lim Cores", "Host Arch"})
	table.SetBorder(true)
	table.SetAutoWrapText(false)

	for _, r := range records {
		table.Append([]string{
			r.VM,
			r.Powerstate,
			r.Cluster,
			r.Namespace,
			r.Host,
			fmt.Sprintf("%d", r.CPUs),
			fmt.Sprintf("%d", r.Sockets),
			fmt.Sprintf("%d", r.CoresPerSocket),
			fmt.Sprintf("%d", r.Threads),
			r.CPUModel,
			fmt.Sprintf("%t", r.DedicatedCPUPlacement),
			fmt.Sprintf("%.2f", r.CPURequests),
			fmt.Sprintf("%.2f", r.CPULimits),
			r.HostNodeCPUModel,
		})
	}

	table.Render()
}

// RenderMemoryTable renders the kvMemory sheet in the terminal.
func RenderMemoryTable(records []engine.KVMemoryRecord, w io.Writer) {
	if len(records) == 0 {
		fmt.Fprintf(w, "No Memory records found.\n")
		return
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"VM", "Powerstate", "Cluster", "Namespace", "Host", "Size (GiB)", "Req (GiB)", "Lim (GiB)", "Overhead (MiB)", "Hugepages", "Balloon"})
	table.SetBorder(true)
	table.SetAutoWrapText(false)

	for _, r := range records {
		table.Append([]string{
			r.VM,
			r.Powerstate,
			r.Cluster,
			r.Namespace,
			r.Host,
			fmt.Sprintf("%.2f", r.SizeGiB),
			fmt.Sprintf("%.2f", r.MemoryRequestsGiB),
			fmt.Sprintf("%.2f", r.MemoryLimitsGiB),
			fmt.Sprintf("%.1f", r.LauncherOverheadMiB),
			r.Hugepages,
			fmt.Sprintf("%t", r.AutoattachMemBalloon),
		})
	}

	table.Render()
}

// RenderDiskTable renders the kvDisk sheet in the terminal.
func RenderDiskTable(records []engine.KVDiskRecord, w io.Writer) {
	if len(records) == 0 {
		fmt.Fprintf(w, "No Virtual Disks found.\n")
		return
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"VM", "Powerstate", "Cluster", "Namespace", "Host", "Disk", "Type", "Claim Name", "StorageClass", "Size (GiB)", "Access Mode", "Bus", "CSI Driver"})
	table.SetBorder(true)
	table.SetAutoWrapText(false)

	for _, r := range records {
		table.Append([]string{
			r.VM,
			r.Powerstate,
			r.Cluster,
			r.Namespace,
			r.Host,
			r.Disk,
			r.VolumeType,
			r.ClaimName,
			r.StorageClass,
			fmt.Sprintf("%.1f", r.CapacityGiB),
			r.AccessMode,
			r.BusType,
			r.CSIDriver,
		})
	}

	table.Render()
}

// RenderPartitionTable renders the kvPartition sheet in the terminal.
func RenderPartitionTable(records []engine.KVPartitionRecord, w io.Writer) {
	if len(records) == 0 {
		fmt.Fprintf(w, "No Guest OS Partition records found.\n")
		return
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"VM", "Powerstate", "Cluster", "Namespace", "Host", "Mount Point", "FS Type", "Disk Device", "Total (GiB)", "Consumed (GiB)", "Free (GiB)", "Free %"})
	table.SetBorder(true)
	table.SetAutoWrapText(false)

	for _, r := range records {
		table.Append([]string{
			r.VM,
			r.Powerstate,
			r.Cluster,
			r.Namespace,
			r.Host,
			r.MountPoint,
			r.FSType,
			r.Disk,
			fmt.Sprintf("%.2f", r.CapacityGiB),
			fmt.Sprintf("%.2f", r.ConsumedGiB),
			fmt.Sprintf("%.2f", r.FreeGiB),
			fmt.Sprintf("%.1f%%", r.FreePercent),
		})
	}

	table.Render()
}

// RenderNetworkTable renders the kvNetwork sheet in the terminal.
func RenderNetworkTable(records []engine.KVNetworkRecord, w io.Writer) {
	if len(records) == 0 {
		fmt.Fprintf(w, "No Network Interface records found.\n")
		return
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"VM", "Powerstate", "Cluster", "Namespace", "Host", "NIC Label", "Network", "Binding", "MAC Address", "IPv4 Address", "Guest IPs"})
	table.SetBorder(true)
	table.SetAutoWrapText(false)

	for _, r := range records {
		table.Append([]string{
			r.VM,
			r.Powerstate,
			r.Cluster,
			r.Namespace,
			r.Host,
			r.NICLabel,
			r.Network,
			r.BindingType,
			r.MacAddress,
			r.IPv4Address,
			r.GuestReportedIPs,
		})
	}

	table.Render()
}

// RenderCDTable renders the kvCD sheet in the terminal.
func RenderCDTable(records []engine.KVCDRecord, w io.Writer) {
	if len(records) == 0 {
		fmt.Fprintf(w, "No CD-ROM / ISO records found.\n")
		return
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"VM", "Powerstate", "Cluster", "Namespace", "Host", "Device", "Type", "Source Image", "Boot Order", "State"})
	table.SetBorder(true)
	table.SetAutoWrapText(false)

	for _, r := range records {
		table.Append([]string{
			r.VM,
			r.Powerstate,
			r.Cluster,
			r.Namespace,
			r.Host,
			r.DeviceNode,
			r.SourceType,
			r.SourceImage,
			r.BootOrder,
			r.Connected,
		})
	}

	table.Render()
}

// RenderSnapshotTable renders the kvSnapshot sheet in the terminal.
func RenderSnapshotTable(records []engine.KVSnapshotRecord, w io.Writer) {
	if len(records) == 0 {
		fmt.Fprintf(w, "No Snapshots found.\n")
		return
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Snapshot Name", "Cluster", "Namespace", "Source VM", "Ready", "Age (Days)", "Date / time", "Error"})
	table.SetBorder(true)
	table.SetAutoWrapText(false)

	for _, r := range records {
		readyStr := "No"
		if r.ReadyToUse {
			readyStr = "\033[32mYes\033[0m"
		}
		table.Append([]string{
			r.SnapshotName,
			r.Cluster,
			r.Namespace,
			r.SourceVM,
			readyStr,
			fmt.Sprintf("%d", r.AgeDays),
			r.CreationDate,
			r.ErrorReason,
		})
	}

	table.Render()
}

// RenderGuestAgentTable renders the kvGuestAgent sheet in the terminal.
func RenderGuestAgentTable(records []engine.KVGuestAgentRecord, w io.Writer) {
	if len(records) == 0 {
		fmt.Fprintf(w, "No Guest Agent records found.\n")
		return
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"VM", "Powerstate", "Cluster", "Namespace", "Host", "Agent Connected", "Agent Ver", "Guest Hostname", "Guest OS", "Kernel Release", "Timezone"})
	table.SetBorder(true)
	table.SetAutoWrapText(false)

	for _, r := range records {
		connStr := "\033[31mNo\033[0m"
		if r.AgentConnected {
			connStr = "\033[32mYes\033[0m"
		}
		table.Append([]string{
			r.VM,
			r.Powerstate,
			r.Cluster,
			r.Namespace,
			r.Host,
			connStr,
			r.AgentVersion,
			r.GuestHostname,
			r.GuestOS,
			r.KernelRelease,
			r.Timezone,
		})
	}

	table.Render()
}

// RenderStoragePoolTable renders the kvStoragePool sheet in the terminal.
func RenderStoragePoolTable(records []engine.KVStoragePoolRecord, w io.Writer) {
	if len(records) == 0 {
		fmt.Fprintf(w, "No Storage Pools / Classes found.\n")
		return
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"StorageClass Name", "Cluster", "Provisioner / CSI Driver", "Reclaim Policy", "Binding Mode", "Expansion", "Default", "Bound PVCs", "Allocated (GiB)"})
	table.SetBorder(true)
	table.SetAutoWrapText(false)

	for _, r := range records {
		defStr := "No"
		if r.IsDefaultClass {
			defStr = "\033[32mYes\033[0m"
		}
		table.Append([]string{
			r.StorageClassName,
			r.Cluster,
			r.ProvisionerCSIDriver,
			r.ReclaimPolicy,
			r.VolumeBindingMode,
			fmt.Sprintf("%t", r.AllowVolumeExpansion),
			defStr,
			fmt.Sprintf("%d", r.TotalBoundPVCCount),
			fmt.Sprintf("%.1f", r.TotalAllocatedSizeGiB),
		})
	}

	table.Render()
}

// RenderHardwareTable renders the kvHardware sheet in the terminal.
func RenderHardwareTable(records []engine.KVHardwareRecord, w io.Writer) {
	if len(records) == 0 {
		fmt.Fprintf(w, "No specialized hardware / GPU devices found.\n")
		return
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"VM", "Cluster", "Namespace", "Host", "Device Type", "Device Name", "Resource Name"})
	table.SetBorder(true)
	table.SetAutoWrapText(false)

	for _, r := range records {
		table.Append([]string{
			r.VM,
			r.Cluster,
			r.Namespace,
			r.Host,
			r.DeviceType,
			r.DeviceName,
			r.ResourceName,
		})
	}

	table.Render()
}

// RenderNodeTable renders the kvNode sheet in the terminal.
func RenderNodeTable(records []engine.KVNodeRecord, w io.Writer) {
	if len(records) == 0 {
		fmt.Fprintf(w, "No Nodes found.\n")
		return
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Host", "Cluster", "Status", "# Cores", "RAM (GiB)", "Alloc CPU", "Alloc RAM", "VM vCPUs", "VM RAM", "Overcommit", "# VMs", "KVM Accel"})
	table.SetBorder(true)
	table.SetAutoWrapText(false)

	for _, r := range records {
		status := r.Status
		if r.Status == "Ready" {
			status = fmt.Sprintf("\033[32m%s\033[0m", r.Status)
		} else {
			status = fmt.Sprintf("\033[31m%s\033[0m", r.Status)
		}

		kvm := "No"
		if r.KVMHardwareAccel {
			kvm = "\033[32mYes\033[0m"
		}

		table.Append([]string{
			r.Host,
			r.Cluster,
			status,
			fmt.Sprintf("%d", r.PhysicalCores),
			fmt.Sprintf("%.1f", r.TotalRAMGiB),
			fmt.Sprintf("%.1f", r.AllocatableCPU),
			fmt.Sprintf("%.1f", r.AllocatableRAMGiB),
			fmt.Sprintf("%.1f", r.AllocatedVMvCPUs),
			fmt.Sprintf("%.1f", r.AllocatedVMRAMGiB),
			fmt.Sprintf("%.2f:1", r.VCPUOvercommitRatio),
			fmt.Sprintf("%d", r.ActiveVMCount),
			kvm,
		})
	}

	table.Render()
}

// RenderTableSheet renders a specific sheet or defaults to Info + Health.
func RenderTableSheet(report *engine.InventoryReport, sheetName string, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}

	RenderSummary(report, w)

	s := strings.ToLower(sheetName)
	switch s {
	case "kvhealth", "health":
		fmt.Fprintf(w, "\033[1;31m--- kvHealth Audit Findings ---\033[0m\n")
		RenderHealthTable(report.Health, w)
	case "kvnode", "node", "nodes":
		fmt.Fprintf(w, "\033[1;34m--- kvNode Compute Infrastructure ---\033[0m\n")
		RenderNodeTable(report.Node, w)
	case "kvcpu", "cpu":
		fmt.Fprintf(w, "\033[1;34m--- kvCPU Compute & CPU Topology ---\033[0m\n")
		RenderCPUTable(report.CPU, w)
	case "kvmemory", "memory", "mem":
		fmt.Fprintf(w, "\033[1;34m--- kvMemory RAM Allocation & Overhead ---\033[0m\n")
		RenderMemoryTable(report.Memory, w)
	case "kvdisk", "disk", "disks":
		fmt.Fprintf(w, "\033[1;33m--- kvDisk Virtual Disks & Storage ---\033[0m\n")
		RenderDiskTable(report.Disk, w)
	case "kvpartition", "partition", "partitions", "fs":
		fmt.Fprintf(w, "\033[1;33m--- kvPartition Guest Filesystems ---\033[0m\n")
		RenderPartitionTable(report.Partition, w)
	case "kvnetwork", "network", "net", "nics":
		fmt.Fprintf(w, "\033[1;36m--- kvNetwork Interfaces & IP Allocation ---\033[0m\n")
		RenderNetworkTable(report.Network, w)
	case "kvcd", "cd", "cdrom", "iso":
		fmt.Fprintf(w, "\033[1;35m--- kvCD CD-ROM & ISO Attachments ---\033[0m\n")
		RenderCDTable(report.CD, w)
	case "kvsnapshot", "snapshot", "snapshots", "snap":
		fmt.Fprintf(w, "\033[1;31m--- kvSnapshot VM Snapshots ---\033[0m\n")
		RenderSnapshotTable(report.Snapshot, w)
	case "kvguestagent", "guestagent", "agent", "ga":
		fmt.Fprintf(w, "\033[1;34m--- kvGuestAgent QEMU Guest Agent Status ---\033[0m\n")
		RenderGuestAgentTable(report.GuestAgent, w)
	case "kvstoragepool", "storagepool", "sc", "storageclasses":
		fmt.Fprintf(w, "\033[1;33m--- kvStoragePool Storage Classes & CSI Pools ---\033[0m\n")
		RenderStoragePoolTable(report.StoragePool, w)
	case "kvhardware", "hardware", "gpu", "gpus", "pci":
		fmt.Fprintf(w, "\033[1;35m--- kvHardware Host Devices & GPUs ---\033[0m\n")
		RenderHardwareTable(report.Hardware, w)
	case "kvinfo", "info", "vms", "":
		fmt.Fprintf(w, "\033[1;32m--- kvInfo VM Inventory ---\033[0m\n")
		RenderInfoTable(report.Info, w)
		if len(report.Health) > 0 {
			fmt.Fprintf(w, "\n\033[1;31m--- kvHealth Audit Findings ---\033[0m\n")
			RenderHealthTable(report.Health, w)
		}
	default:
		RenderInfoTable(report.Info, w)
	}
}
