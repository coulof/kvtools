package engine

import (
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	virtv1 "kubevirt.io/api/core/v1"

	"github.com/coulof/kvtools/pkg/collector"
	"github.com/coulof/kvtools/pkg/utils"
)

// RunHealthAudit executes all built-in health, migration, and performance optimization rules.
func RunHealthAudit(raw *collector.RawData, nodeRecords []KVNodeRecord) []KVHealthRecord {
	var records []KVHealthRecord

	if raw == nil {
		return records
	}

	// Build lookups
	vmMap := make(map[string]virtv1.VirtualMachine)
	vmUIDMap := make(map[string]bool)
	for _, vm := range raw.VMs {
		key := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
		vmMap[key] = vm
		if string(vm.UID) != "" {
			vmUIDMap[string(vm.UID)] = true
		}
	}

	vmiMap := make(map[string]virtv1.VirtualMachineInstance)
	vmiUIDMap := make(map[string]bool)
	for _, vmi := range raw.VMIs {
		key := fmt.Sprintf("%s/%s", vmi.Namespace, vmi.Name)
		vmiMap[key] = vmi
		if string(vmi.UID) != "" {
			vmiUIDMap[string(vmi.UID)] = true
		}
	}

	pvcMap := make(map[string]corev1.PersistentVolumeClaim)
	for _, pvc := range raw.PVCs {
		key := fmt.Sprintf("%s/%s", pvc.Namespace, pvc.Name)
		pvcMap[key] = pvc
	}

	// 1. Audit VMs / VMIs
	for _, vm := range raw.VMs {
		vmiKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
		vmi, hasVMI := vmiMap[vmiKey]

		vSpec := vm.Spec.Template.Spec
		evictionStrategy := ""
		if vm.Spec.Template.Spec.EvictionStrategy != nil {
			evictionStrategy = string(*vm.Spec.Template.Spec.EvictionStrategy)
		}

		// HLTH-001: Migration Blocker (CRITICAL) - VM has RWO PVC and live migration is requested
		isLiveMigrateRequested := evictionStrategy == string(virtv1.EvictionStrategyLiveMigrate) ||
			evictionStrategy == string(virtv1.EvictionStrategyLiveMigrateIfPossible)
		if isLiveMigrateRequested {
			for _, vol := range vSpec.Volumes {
				pvcName := ""
				if vol.PersistentVolumeClaim != nil {
					pvcName = vol.PersistentVolumeClaim.ClaimName
				} else if vol.DataVolume != nil {
					pvcName = vol.DataVolume.Name
				}
				if pvcName != "" {
					pvcKey := fmt.Sprintf("%s/%s", vm.Namespace, pvcName)
					if pvc, ok := pvcMap[pvcKey]; ok {
						for _, accessMode := range pvc.Spec.AccessModes {
							if accessMode == corev1.ReadWriteOnce || accessMode == corev1.ReadWriteOncePod {
								records = append(records, KVHealthRecord{
									RuleID:       "HLTH-001",
									Category:     "Migration Blocker",
									Severity:     "CRITICAL",
									ResourceKind: "VirtualMachine",
									ResourceName: vm.Name,
									Cluster:      raw.ClusterName,
									Namespace:    vm.Namespace,
									IssueSummary: fmt.Sprintf("Live migration configured but volume '%s' is backed by ReadWriteOnce PVC '%s'", vol.Name, pvcName),
									Remediation:  "Change storage to RWX (shared) or configure CDI block volume migration before attempting live migration",
								})
								break
							}
						}
					}
				}
			}
		}

		// HLTH-002: Migration Blocker (WARNING) - VM uses host-passthrough CPU model
		if vSpec.Domain.CPU != nil && strings.EqualFold(vSpec.Domain.CPU.Model, "host-passthrough") {
			records = append(records, KVHealthRecord{
				RuleID:       "HLTH-002",
				Category:     "Migration Blocker",
				Severity:     "WARNING",
				ResourceKind: "VirtualMachine",
				ResourceName: vm.Name,
				Cluster:      raw.ClusterName,
				Namespace:    vm.Namespace,
				IssueSummary: "VM is configured with 'host-passthrough' CPU model",
				Remediation:  "Switch to 'host-model' or ensure all cluster nodes have identical CPU microarchitecture",
			})
		}

		// HLTH-003: Migration Blocker (CRITICAL) - VM mounts a local HostDisk
		for _, vol := range vSpec.Volumes {
			if vol.HostDisk != nil {
				records = append(records, KVHealthRecord{
					RuleID:       "HLTH-003",
					Category:     "Migration Blocker",
					Severity:     "CRITICAL",
					ResourceKind: "VirtualMachine",
					ResourceName: vm.Name,
					Cluster:      raw.ClusterName,
					Namespace:    vm.Namespace,
					IssueSummary: fmt.Sprintf("VM mounts local HostDisk '%s' at path '%s'", vol.Name, vol.HostDisk.Path),
					Remediation:  "Replace HostDisk with PVC or ContainerDisk before attempting live migration",
				})
			}
		}

		// HLTH-004: Migration Blocker (WARNING) - VM has direct SR-IOV NIC without bond failover
		hasSRIOV := false
		for _, iface := range vSpec.Domain.Devices.Interfaces {
			if iface.SRIOV != nil {
				hasSRIOV = true
				break
			}
		}
		if hasSRIOV && len(vSpec.Domain.Devices.Interfaces) == 1 {
			records = append(records, KVHealthRecord{
				RuleID:       "HLTH-004",
				Category:     "Migration Blocker",
				Severity:     "WARNING",
				ResourceKind: "VirtualMachine",
				ResourceName: vm.Name,
				Cluster:      raw.ClusterName,
				Namespace:    vm.Namespace,
				IssueSummary: "VM has direct SR-IOV NIC without a backup/failover interface",
				Remediation:  "Configure a secondary or bond failover network interface to ensure network connectivity during migration",
			})
		}

		// HLTH-007: Guest Visibility (WARNING) - Running VM has AgentConnected: false
		if hasVMI && vmi.Status.Phase == virtv1.Running {
			agentConnected := false
			for _, cond := range vmi.Status.Conditions {
				if cond.Type == virtv1.VirtualMachineInstanceAgentConnected && cond.Status == corev1.ConditionTrue {
					agentConnected = true
					break
				}
			}
			if !agentConnected {
				records = append(records, KVHealthRecord{
					RuleID:       "HLTH-007",
					Category:     "Guest Visibility",
					Severity:     "WARNING",
					ResourceKind: "VirtualMachine",
					ResourceName: vm.Name,
					Cluster:      raw.ClusterName,
					Namespace:    vm.Namespace,
					IssueSummary: "Running VM does not have QEMU Guest Agent connected",
					Remediation:  "Install and enable 'qemu-guest-agent' inside guest OS for IP reporting and filesystem quiescing",
				})
			}
		}

		// HLTH-009: QoS / Sizing (WARNING) - VM CPU limit > 4x CPU request
		if vSpec.Domain.Resources.Requests != nil && vSpec.Domain.Resources.Limits != nil {
			reqCPU := vSpec.Domain.Resources.Requests.Cpu()
			limCPU := vSpec.Domain.Resources.Limits.Cpu()
			if reqCPU != nil && limCPU != nil && !reqCPU.IsZero() && !limCPU.IsZero() {
				reqVal := reqCPU.MilliValue()
				limVal := limCPU.MilliValue()
				if limVal > 4*reqVal {
					records = append(records, KVHealthRecord{
						RuleID:       "HLTH-009",
						Category:     "QoS Risk",
						Severity:     "WARNING",
						ResourceKind: "VirtualMachine",
						ResourceName: vm.Name,
						Cluster:      raw.ClusterName,
						Namespace:    vm.Namespace,
						IssueSummary: fmt.Sprintf("CPU limit (%s) is greater than 4x CPU request (%s)", limCPU.String(), reqCPU.String()),
						Remediation:  "Set consistent CPU limits or remove limits to prevent aggressive CFS scheduler throttling",
					})
				}
			}
		}

		// HLTH-010: QoS / Sizing (CRITICAL) - Memory limits equal memory request with zero overhead margin
		if vSpec.Domain.Resources.Requests != nil && vSpec.Domain.Resources.Limits != nil {
			reqMem := vSpec.Domain.Resources.Requests.Memory()
			limMem := vSpec.Domain.Resources.Limits.Memory()
			guestMem := vSpec.Domain.Memory
			if reqMem != nil && limMem != nil && !reqMem.IsZero() && !limMem.IsZero() {
				if reqMem.Cmp(*limMem) == 0 && (guestMem == nil || guestMem.Guest == nil) {
					records = append(records, KVHealthRecord{
						RuleID:       "HLTH-010",
						Category:     "QoS Risk",
						Severity:     "CRITICAL",
						ResourceKind: "VirtualMachine",
						ResourceName: vm.Name,
						Cluster:      raw.ClusterName,
						Namespace:    vm.Namespace,
						IssueSummary: fmt.Sprintf("Memory limit (%s) is identical to memory request with no launcher overhead margin specified", limMem.String()),
						Remediation:  "Specify domain.memory.guest or configure launcher overhead to prevent pod OOM termination",
					})
				}
			}
		}

		// HLTH-013: Disk I/O Performance Tip (INFO) - VM has >3 disks and size > 500 GiB without dedicated IOThread or multi-queue
		totalDiskSizeGiB := 0.0
		diskCount := 0
		hasDedicatedIO := false
		for _, disk := range vSpec.Domain.Devices.Disks {
			if disk.DiskDevice.CDRom == nil {
				diskCount++
				if disk.DedicatedIOThread != nil && *disk.DedicatedIOThread {
					hasDedicatedIO = true
				}
			}
		}
		for _, vol := range vSpec.Volumes {
			pvcName := ""
			if vol.PersistentVolumeClaim != nil {
				pvcName = vol.PersistentVolumeClaim.ClaimName
			}
			if pvcName != "" {
				if pvc, ok := pvcMap[fmt.Sprintf("%s/%s", vm.Namespace, pvcName)]; ok {
					if q, ok := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; ok {
						totalDiskSizeGiB += utils.QuantityToGiB(&q)
					}
				}
			}
		}
		if diskCount > 3 && totalDiskSizeGiB > 500.0 && !hasDedicatedIO {
			records = append(records, KVHealthRecord{
				RuleID:       "HLTH-013",
				Category:     "Performance Tip",
				Severity:     "INFO",
				ResourceKind: "VirtualMachine",
				ResourceName: vm.Name,
				Cluster:      raw.ClusterName,
				Namespace:    vm.Namespace,
				IssueSummary: fmt.Sprintf("VM has %d disks totaling %.0f GiB without dedicated IOThreads enabled", diskCount, totalDiskSizeGiB),
				Remediation:  "Enable dedicatedIOThread on disk devices or configure blockMultiQueue for optimal I/O parallelism",
			})
		}

		// HLTH-014: In-Memory / NUMA Performance Tip (INFO) - VM has >= 4 cores without guest NUMA topology configured
		cores := uint32(1)
		sockets := uint32(1)
		threads := uint32(1)
		if vSpec.Domain.CPU != nil {
			if vSpec.Domain.CPU.Cores > 0 {
				cores = vSpec.Domain.CPU.Cores
			}
			if vSpec.Domain.CPU.Sockets > 0 {
				sockets = vSpec.Domain.CPU.Sockets
			}
			if vSpec.Domain.CPU.Threads > 0 {
				threads = vSpec.Domain.CPU.Threads
			}
		}
		totalVCPUs := cores * sockets * threads
		hasNUMA := vSpec.Domain.CPU != nil && vSpec.Domain.CPU.NUMA != nil
		if totalVCPUs >= 4 && !hasNUMA {
			records = append(records, KVHealthRecord{
				RuleID:       "HLTH-014",
				Category:     "Performance Tip",
				Severity:     "INFO",
				ResourceKind: "VirtualMachine",
				ResourceName: vm.Name,
				Cluster:      raw.ClusterName,
				Namespace:    vm.Namespace,
				IssueSummary: fmt.Sprintf("VM has %d vCPUs configured without guest NUMA topology mapping", totalVCPUs),
				Remediation:  "Configure domain.cpu.numa guestMappingPassthrough to align vCPU sockets with host NUMA nodes for reduced memory latency",
			})
		}

		// HLTH-015: Connected Temporary CD-ROM / ISO (WARNING) - Running VM in non-system namespace has CD-ROM connected
		if hasVMI && vmi.Status.Phase == virtv1.Running && !strings.HasPrefix(vm.Namespace, "kube-") {
			for _, disk := range vSpec.Domain.Devices.Disks {
				if disk.DiskDevice.CDRom != nil {
					records = append(records, KVHealthRecord{
						RuleID:       "HLTH-015",
						Category:     "Hygiene",
						Severity:     "WARNING",
						ResourceKind: "VirtualMachine",
						ResourceName: vm.Name,
						Cluster:      raw.ClusterName,
						Namespace:    vm.Namespace,
						IssueSummary: fmt.Sprintf("Running VM has CD-ROM / ISO device '%s' attached", disk.Name),
						Remediation:  "Disconnect temporary ISO/CD-ROM devices when installation is complete to avoid locking storage resources",
					})
				}
			}
		}
	}

	// HLTH-012: Low Guest Disk Space (WARNING) - Guest filesystem partition free space < 10% or < 5 GiB
	for key, fsList := range raw.FileSystemList {
		if fsList == nil {
			continue
		}
		parts := strings.SplitN(key, "/", 2)
		ns := parts[0]
		vmName := ""
		if len(parts) > 1 {
			vmName = parts[1]
		}

		for _, fs := range fsList.Items {
			if fs.TotalBytes <= 0 {
				continue
			}
			freeBytes := int64(fs.TotalBytes - fs.UsedBytes)
			if freeBytes < 0 {
				freeBytes = 0
			}
			freeGiB := utils.BytesToGiB(freeBytes)
			freePct := float64(freeBytes) / float64(fs.TotalBytes) * 100.0

			if freePct < 10.0 || freeGiB < 5.0 {
				records = append(records, KVHealthRecord{
					RuleID:       "HLTH-012",
					Category:     "Storage Space",
					Severity:     "WARNING",
					ResourceKind: "VirtualMachine",
					ResourceName: vmName,
					Cluster:      raw.ClusterName,
					Namespace:    ns,
					IssueSummary: fmt.Sprintf("Guest partition '%s' on %s has only %.1f GiB free (%.1f%%)", fs.MountPoint, fs.DiskName, freeGiB, freePct),
					Remediation:  "Expand underlying PVC or clean up guest filesystem to avoid application crashes due to full disk",
				})
			}
		}
	}

	// HLTH-005: Storage Zombie (WARNING) - Orphaned PVC matching KubeVirt pattern without an active VM/VMI
	for _, pvc := range raw.PVCs {
		isKubeVirtPVC := false
		for k := range pvc.Labels {
			if strings.Contains(k, "kubevirt.io") || strings.Contains(k, "cdi.kubevirt.io") {
				isKubeVirtPVC = true
				break
			}
		}
		if strings.Contains(pvc.Name, "-disk-") || strings.Contains(pvc.Name, "-dv-") {
			isKubeVirtPVC = true
		}

		if isKubeVirtPVC && len(pvc.OwnerReferences) > 0 {
			hasActiveOwner := false
			for _, owner := range pvc.OwnerReferences {
				if vmUIDMap[string(owner.UID)] || vmiUIDMap[string(owner.UID)] {
					hasActiveOwner = true
					break
				}
			}
			if !hasActiveOwner {
				records = append(records, KVHealthRecord{
					RuleID:       "HLTH-005",
					Category:     "Storage Zombie",
					Severity:     "WARNING",
					ResourceKind: "PersistentVolumeClaim",
					ResourceName: pvc.Name,
					Cluster:      raw.ClusterName,
					Namespace:    pvc.Namespace,
					IssueSummary: fmt.Sprintf("PVC '%s' has KubeVirt metadata but owning VM/VMI UID no longer exists", pvc.Name),
					Remediation:  "Delete orphaned PVC if no longer required to reclaim storage capacity",
				})
			}
		}
	}

	// HLTH-006: Storage Zombie (WARNING) - DataVolume in Failed or stuck ImportInProgress state > 2h
	for _, dv := range raw.DataVolumes {
		phase, _, _ := unstructured.NestedString(dv.Object, "status", "phase")
		dvName := dv.GetName()
		dvNS := dv.GetNamespace()
		creationTime := dv.GetCreationTimestamp().Time

		if phase == "Failed" {
			records = append(records, KVHealthRecord{
				RuleID:       "HLTH-006",
				Category:     "Storage Zombie",
				Severity:     "WARNING",
				ResourceKind: "DataVolume",
				ResourceName: dvName,
				Cluster:      raw.ClusterName,
				Namespace:    dvNS,
				IssueSummary: fmt.Sprintf("DataVolume '%s' is in Failed phase", dvName),
				Remediation:  "Inspect CDI importer pod logs and delete or re-import the failed DataVolume",
			})
		} else if (phase == "ImportInProgress" || phase == "CloneInProgress") && time.Since(creationTime) > 2*time.Hour {
			records = append(records, KVHealthRecord{
				RuleID:       "HLTH-006",
				Category:     "Storage Zombie",
				Severity:     "WARNING",
				ResourceKind: "DataVolume",
				ResourceName: dvName,
				Cluster:      raw.ClusterName,
				Namespace:    dvNS,
				IssueSummary: fmt.Sprintf("DataVolume '%s' has been in %s for %s", dvName, phase, utils.FormatDuration(time.Since(creationTime))),
				Remediation:  "Inspect CDI importer/cloner pod logs for storage throughput bottlenecks or stuck downloads",
			})
		}
	}

	// HLTH-008: Snapshot Sprawl (WARNING) - VirtualMachineSnapshot age > 14 days
	for _, snap := range raw.VMSnapshots {
		snapName := snap.GetName()
		snapNS := snap.GetNamespace()
		creationTime := snap.GetCreationTimestamp().Time
		ageDays := utils.FormatAgeDays(creationTime)

		if ageDays > 14 {
			records = append(records, KVHealthRecord{
				RuleID:       "HLTH-008",
				Category:     "Snapshot Sprawl",
				Severity:     "WARNING",
				ResourceKind: "VirtualMachineSnapshot",
				ResourceName: snapName,
				Cluster:      raw.ClusterName,
				Namespace:    snapNS,
				IssueSummary: fmt.Sprintf("VirtualMachineSnapshot '%s' is %d days old (> 14 days)", snapName, ageDays),
				Remediation:  "Merge or delete aged snapshots to avoid storage bloat and snapshot chain limits",
			})
		}
	}

	// HLTH-011: Node Imbalance (WARNING) - Node vCPU overcommit ratio exceeds 8:1
	for _, nr := range nodeRecords {
		if nr.VCPUOvercommitRatio > 8.0 {
			records = append(records, KVHealthRecord{
				RuleID:       "HLTH-011",
				Category:     "Node Imbalance",
				Severity:     "WARNING",
				ResourceKind: "Node",
				ResourceName: nr.Host,
				Cluster:      nr.Cluster,
				Namespace:    "-",
				IssueSummary: fmt.Sprintf("Node '%s' has vCPU overcommit ratio of %.2f:1 (> 8.0:1)", nr.Host, nr.VCPUOvercommitRatio),
				Remediation:  "Rebalance VMs across cluster nodes or add compute nodes to decrease overcommit ratio",
			})
		}
	}

	// HLTH-016: Node Under Pressure (WARNING) - Node reporting MemoryPressure, DiskPressure, or PIDPressure
	for _, node := range raw.Nodes {
		for _, cond := range node.Status.Conditions {
			if cond.Status == corev1.ConditionTrue {
				if cond.Type == corev1.NodeMemoryPressure || cond.Type == corev1.NodeDiskPressure || cond.Type == corev1.NodePIDPressure {
					records = append(records, KVHealthRecord{
						RuleID:       "HLTH-016",
						Category:     "Node Health",
						Severity:     "WARNING",
						ResourceKind: "Node",
						ResourceName: node.Name,
						Cluster:      raw.ClusterName,
						Namespace:    "-",
						IssueSummary: fmt.Sprintf("Node '%s' is under %s: %s", node.Name, cond.Type, cond.Message),
						Remediation:  "Investigate host node resource consumption, reclaim disk/memory, or drain node",
					})
				}
			}
		}
	}

	return records
}
