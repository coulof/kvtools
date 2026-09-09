package engine

import (
	"fmt"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	virtv1 "kubevirt.io/api/core/v1"

	"github.com/coulof/kvtools/pkg/collector"
	"github.com/coulof/kvtools/pkg/utils"
)

// Transform converts collector.RawData into a fully populated InventoryReport.
func Transform(raw *collector.RawData) *InventoryReport {
	report := &InventoryReport{
		ClusterName: raw.ClusterName,
		GeneratedAt: raw.CollectedAt,
	}

	// Build lookups
	vmiMap := make(map[string]virtv1.VirtualMachineInstance)
	for _, vmi := range raw.VMIs {
		vmiMap[fmt.Sprintf("%s/%s", vmi.Namespace, vmi.Name)] = vmi
	}

	pvcMap := make(map[string]corev1.PersistentVolumeClaim)
	for _, pvc := range raw.PVCs {
		pvcMap[fmt.Sprintf("%s/%s", pvc.Namespace, pvc.Name)] = pvc
	}

	pvMap := make(map[string]corev1.PersistentVolume)
	for _, pv := range raw.PVs {
		pvMap[pv.Name] = pv
	}

	scMap := make(map[string]storagev1.StorageClass)
	for _, sc := range raw.StorageClasses {
		scMap[sc.Name] = sc
	}

	nodeMap := make(map[string]corev1.Node)
	for _, node := range raw.Nodes {
		nodeMap[node.Name] = node
	}

	// Calculate per-node VM allocations
	nodeVMCount := make(map[string]int)
	nodeAllocCPU := make(map[string]float64)
	nodeAllocRAM := make(map[string]float64)

	for _, vmi := range raw.VMIs {
		if vmi.Status.Phase == virtv1.Running && vmi.Status.NodeName != "" {
			nodeName := vmi.Status.NodeName
			nodeVMCount[nodeName]++

			// Calculate vCPUs
			cores := uint32(1)
			sockets := uint32(1)
			threads := uint32(1)
			if vmi.Spec.Domain.CPU != nil {
				if vmi.Spec.Domain.CPU.Cores > 0 {
					cores = vmi.Spec.Domain.CPU.Cores
				}
				if vmi.Spec.Domain.CPU.Sockets > 0 {
					sockets = vmi.Spec.Domain.CPU.Sockets
				}
				if vmi.Spec.Domain.CPU.Threads > 0 {
					threads = vmi.Spec.Domain.CPU.Threads
				}
			}
			nodeAllocCPU[nodeName] += float64(cores * sockets * threads)

			// Calculate Memory
			if vmi.Spec.Domain.Memory != nil && vmi.Spec.Domain.Memory.Guest != nil {
				nodeAllocRAM[nodeName] += utils.QuantityToGiB(vmi.Spec.Domain.Memory.Guest)
			} else if vmi.Spec.Domain.Resources.Requests != nil {
				mem := vmi.Spec.Domain.Resources.Requests.Memory()
				nodeAllocRAM[nodeName] += utils.QuantityToGiB(mem)
			}
		}
	}

	// 1. Transform kvNode (vHost in RVTools)
	for _, node := range raw.Nodes {
		status := "NotReady"
		for _, cond := range node.Status.Conditions {
			if cond.Type == corev1.NodeReady && cond.Status == corev1.ConditionTrue {
				status = "Ready"
				break
			}
		}
		if node.Spec.Unschedulable {
			status = "SchedulingDisabled"
		}

		totalCores := node.Status.Capacity.Cpu().Value()
		totalRAM := utils.QuantityToGiB(node.Status.Capacity.Memory())
		allocCPU := utils.QuantityToCores(node.Status.Allocatable.Cpu())
		allocRAM := utils.QuantityToGiB(node.Status.Allocatable.Memory())
		allocatedVMvCPUs := nodeAllocCPU[node.Name]
		allocatedVMRAMGiB := nodeAllocRAM[node.Name]

		overcommitRatio := 0.0
		if allocCPU > 0 {
			overcommitRatio = utils.BytesToGiB(int64(allocatedVMvCPUs / allocCPU * float64(utils.GiB)))
		}

		// KVM hardware acceleration check
		hasKVM := false
		if _, exists := node.Status.Allocatable["devices.kubevirt.io/kvm"]; exists {
			hasKVM = true
		}
		if val, exists := node.Labels["kubevirt.io/schedulable"]; exists && val == "true" {
			hasKVM = true
		}

		var taintList []string
		for _, t := range node.Spec.Taints {
			taintList = append(taintList, fmt.Sprintf("%s=%s:%s", t.Key, t.Value, t.Effect))
		}
		taintsStr := strings.Join(taintList, ", ")

		var condList []string
		for _, c := range node.Status.Conditions {
			if c.Status == corev1.ConditionTrue && c.Type != corev1.NodeReady {
				condList = append(condList, string(c.Type))
			}
		}
		condsStr := strings.Join(condList, ", ")

		report.Node = append(report.Node, KVNodeRecord{
			Host:                node.Name,
			Cluster:             raw.ClusterName,
			Status:              status,
			PhysicalCores:       totalCores,
			TotalRAMGiB:         totalRAM,
			AllocatableCPU:      allocCPU,
			AllocatableRAMGiB:   allocRAM,
			AllocatedVMvCPUs:    allocatedVMvCPUs,
			AllocatedVMRAMGiB:   allocatedVMRAMGiB,
			VCPUOvercommitRatio: overcommitRatio,
			ActiveVMCount:       nodeVMCount[node.Name],
			KVMHardwareAccel:    hasKVM,
			KubernetesVersion:   node.Status.NodeInfo.KubeletVersion,
			OSImage:             node.Status.NodeInfo.OSImage,
			KernelVersion:       node.Status.NodeInfo.KernelVersion,
			NodeTaints:          taintsStr,
			NodeConditions:      condsStr,
		})
	}
	sort.Slice(report.Node, func(i, j int) bool {
		return report.Node[i].Host < report.Node[j].Host
	})

	// 2. Transform VMs into kvInfo, kvCPU, kvMemory, kvDisk, kvNetwork, kvCD, kvHardware, kvPartition, kvGuestAgent
	for _, vm := range raw.VMs {
		key := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
		vmi, hasVMI := vmiMap[key]

		// Powerstate resolution (RVTools: poweredOn, poweredOff, etc.)
		powerState := "poweredOff"
		if hasVMI {
			if vmi.Status.Phase == virtv1.Running {
				powerState = "poweredOn"
				for _, cond := range vmi.Status.Conditions {
					if cond.Type == virtv1.VirtualMachineInstancePaused && cond.Status == corev1.ConditionTrue {
						powerState = "suspended"
						break
					}
				}
			} else {
				powerState = string(vmi.Status.Phase)
			}
		} else if vm.Spec.Running != nil && *vm.Spec.Running {
			powerState = "starting"
		} else if vm.Spec.RunStrategy != nil && *vm.Spec.RunStrategy != virtv1.RunStrategyHalted {
			powerState = "starting"
		}

		runStrategy := "N/A"
		if vm.Spec.RunStrategy != nil {
			runStrategy = string(*vm.Spec.RunStrategy)
		} else if vm.Spec.Running != nil {
			if *vm.Spec.Running {
				runStrategy = "Running: true"
			} else {
				runStrategy = "Running: false"
			}
		}

		hostNode := ""
		if hasVMI {
			hostNode = vmi.Status.NodeName
		}

		// Primary IP resolution
		primaryIP := ""
		if hasVMI && len(vmi.Status.Interfaces) > 0 {
			primaryIP = vmi.Status.Interfaces[0].IP
		}

		dnsName := ""
		guestOS := "Unknown"
		if osInfo, ok := raw.GuestOSInfo[key]; ok && osInfo != nil {
			dnsName = osInfo.Hostname
			if osInfo.OS.PrettyName != "" {
				guestOS = osInfo.OS.PrettyName
			} else if osInfo.OS.Name != "" {
				guestOS = osInfo.OS.Name
			}
		} else if hasVMI && vmi.Status.GuestOSInfo.PrettyName != "" {
			guestOS = vmi.Status.GuestOSInfo.PrettyName
		}

		// Firmware & Boot
		vSpec := vm.Spec.Template.Spec
		firmware := "BIOS"
		efiSecureBoot := false
		if vSpec.Domain.Firmware != nil && vSpec.Domain.Firmware.Bootloader != nil {
			if vSpec.Domain.Firmware.Bootloader.EFI != nil {
				firmware = "EFI"
				if vSpec.Domain.Firmware.Bootloader.EFI.SecureBoot != nil && *vSpec.Domain.Firmware.Bootloader.EFI.SecureBoot {
					efiSecureBoot = true
				}
			}
		}

		// TPM
		tpmEnabled := vSpec.Domain.Devices.TPM != nil

		// CPU Topology
		cores := uint32(1)
		sockets := uint32(1)
		threads := uint32(1)
		cpuModel := "default"
		dedicatedPlacement := false
		numaNodes := 0
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
			if vSpec.Domain.CPU.Model != "" {
				cpuModel = vSpec.Domain.CPU.Model
			}
			dedicatedPlacement = vSpec.Domain.CPU.DedicatedCPUPlacement
			if vSpec.Domain.CPU.NUMA != nil && vSpec.Domain.CPU.NUMA.GuestMappingPassthrough != nil {
				numaNodes = 1
			}
		}
		totalVCPUs := cores * sockets * threads
		cpusSummary := fmt.Sprintf("%d Cores, %d Sockets, %d Threads (%d vCPUs)", cores, sockets, threads, totalVCPUs)

		// CPU Requests & Limits
		cpuReq := 0.0
		cpuLim := 0.0
		if vSpec.Domain.Resources.Requests != nil {
			cpuReq = utils.QuantityToCores(vSpec.Domain.Resources.Requests.Cpu())
		}
		if vSpec.Domain.Resources.Limits != nil {
			cpuLim = utils.QuantityToCores(vSpec.Domain.Resources.Limits.Cpu())
		}

		// Host Node CPU Model
		hostNodeCPUModel := "N/A"
		if hostNode != "" {
			if n, ok := nodeMap[hostNode]; ok {
				hostNodeCPUModel = n.Status.NodeInfo.Architecture
			}
		}

		// Memory
		guestRAMGiB := 0.0
		memReqGiB := 0.0
		memLimGiB := 0.0
		if vSpec.Domain.Memory != nil && vSpec.Domain.Memory.Guest != nil {
			guestRAMGiB = utils.QuantityToGiB(vSpec.Domain.Memory.Guest)
		}
		if vSpec.Domain.Resources.Requests != nil {
			memReqGiB = utils.QuantityToGiB(vSpec.Domain.Resources.Requests.Memory())
			if guestRAMGiB == 0 {
				guestRAMGiB = memReqGiB
			}
		}
		if vSpec.Domain.Resources.Limits != nil {
			memLimGiB = utils.QuantityToGiB(vSpec.Domain.Resources.Limits.Memory())
		}

		overheadMiB := 0.0
		if memReqGiB > 0 && guestRAMGiB > 0 && memReqGiB > guestRAMGiB {
			overheadMiB = (memReqGiB - guestRAMGiB) * 1024.0
		}

		hugepages := "None"
		if vSpec.Domain.Memory != nil && vSpec.Domain.Memory.Hugepages != nil {
			hugepages = vSpec.Domain.Memory.Hugepages.PageSize
		}

		autoattachMemBalloon := true
		if vSpec.Domain.Devices.AutoattachMemBalloon != nil {
			autoattachMemBalloon = *vSpec.Domain.Devices.AutoattachMemBalloon
		}

		// Format labels & annotations
		var labelPairs []string
		for k, v := range vm.Labels {
			labelPairs = append(labelPairs, fmt.Sprintf("%s=%s", k, v))
		}
		sort.Strings(labelPairs)
		labelsStr := strings.Join(labelPairs, ", ")

		var annotPairs []string
		for k, v := range vm.Annotations {
			if strings.HasPrefix(k, "kubectl.kubernetes.io/") {
				continue
			}
			annotPairs = append(annotPairs, fmt.Sprintf("%s=%s", k, v))
		}
		sort.Strings(annotPairs)
		annotStr := strings.Join(annotPairs, ", ")

		// Uptime
		uptime := "N/A"
		if hasVMI && vmi.Status.Phase == virtv1.Running {
			uptime = utils.FormatUptime(&vmi.CreationTimestamp.Time)
		}

		creationDate := utils.FormatTimestamp(&vm.CreationTimestamp.Time)

		// Disks & Volume maps
		volMap := make(map[string]virtv1.Volume)
		for _, vol := range vSpec.Volumes {
			volMap[vol.Name] = vol
		}

		disksCount := len(vSpec.Domain.Devices.Disks)
		nicsCount := len(vSpec.Domain.Devices.Interfaces)
		totalDiskCapGB := 0.0

		// Affinity rules summary (RVTools: Cluster rules)
		var affinityParts []string
		if vSpec.Affinity != nil {
			if vSpec.Affinity.PodAntiAffinity != nil {
				affinityParts = append(affinityParts, "PodAntiAffinity")
			}
			if vSpec.Affinity.PodAffinity != nil {
				affinityParts = append(affinityParts, "PodAffinity")
			}
			if vSpec.Affinity.NodeAffinity != nil {
				affinityParts = append(affinityParts, "NodeAffinity")
			}
		}
		if len(vSpec.NodeSelector) > 0 {
			affinityParts = append(affinityParts, "NodeSelector")
		}
		clusterRules := strings.Join(affinityParts, ", ")

		cpuHotAddMax := uint32(0)
		if vSpec.Domain.CPU != nil && vSpec.Domain.CPU.MaxSockets > 0 {
			cpuHotAddMax = vSpec.Domain.CPU.MaxSockets
		}

		memHotAddMaxGiB := 0.0
		if vSpec.Domain.Memory != nil && vSpec.Domain.Memory.MaxGuest != nil {
			memHotAddMaxGiB = utils.QuantityToGiB(vSpec.Domain.Memory.MaxGuest)
		}

		// kvCPU
		report.CPU = append(report.CPU, KVCPURecord{
			VM:                    vm.Name,
			Powerstate:            powerState,
			Cluster:               raw.ClusterName,
			Namespace:             vm.Namespace,
			Host:                  hostNode,
			CPUs:                  totalVCPUs,
			Sockets:               sockets,
			CoresPerSocket:        cores,
			Threads:               threads,
			CPUModel:              cpuModel,
			DedicatedCPUPlacement: dedicatedPlacement,
			NUMANodes:             numaNodes,
			CPURequests:           cpuReq,
			CPULimits:             cpuLim,
			HostNodeCPUModel:      hostNodeCPUModel,
		})

		// kvMemory
		report.Memory = append(report.Memory, KVMemoryRecord{
			VM:                   vm.Name,
			Powerstate:           powerState,
			Cluster:              raw.ClusterName,
			Namespace:            vm.Namespace,
			Host:                 hostNode,
			SizeGiB:              guestRAMGiB,
			MemoryRequestsGiB:    memReqGiB,
			MemoryLimitsGiB:      memLimGiB,
			LauncherOverheadMiB:  overheadMiB,
			Hugepages:            hugepages,
			AutoattachMemBalloon: autoattachMemBalloon,
			MemoryDumpEnabled:    false,
		})

		// kvDisk & kvCD
		for _, disk := range vSpec.Domain.Devices.Disks {
			vol, hasVol := volMap[disk.Name]

			isCDRom := disk.DiskDevice.CDRom != nil
			busType := "virtio"
			if disk.DiskDevice.Disk != nil && disk.DiskDevice.Disk.Bus != "" {
				busType = string(disk.DiskDevice.Disk.Bus)
			} else if disk.DiskDevice.CDRom != nil && disk.DiskDevice.CDRom.Bus != "" {
				busType = string(disk.DiskDevice.CDRom.Bus)
			} else if disk.DiskDevice.LUN != nil && disk.DiskDevice.LUN.Bus != "" {
				busType = string(disk.DiskDevice.LUN.Bus)
			}

			volType := "Unknown"
			claimName := ""
			storageClass := "N/A"
			provSizeGiB := 0.0
			volMode := "N/A"
			accessMode := "N/A"
			csiDriver := "N/A"
			sourceImage := ""

			if hasVol {
				if vol.PersistentVolumeClaim != nil {
					volType = "PersistentVolumeClaim"
					claimName = vol.PersistentVolumeClaim.ClaimName
				} else if vol.DataVolume != nil {
					volType = "DataVolume"
					claimName = vol.DataVolume.Name
				} else if vol.ContainerDisk != nil {
					volType = "ContainerDisk"
					sourceImage = vol.ContainerDisk.Image
				} else if vol.HostDisk != nil {
					volType = "HostDisk"
					sourceImage = vol.HostDisk.Path
				} else if vol.CloudInitNoCloud != nil || vol.CloudInitConfigDrive != nil {
					volType = "CloudInit"
				} else if vol.ConfigMap != nil {
					volType = "ConfigMap"
				} else if vol.Secret != nil {
					volType = "Secret"
				} else if vol.ServiceAccount != nil {
					volType = "ServiceAccount"
				}

				// Resolve PVC details
				if claimName != "" {
					pvcKey := fmt.Sprintf("%s/%s", vm.Namespace, claimName)
					if pvc, ok := pvcMap[pvcKey]; ok {
						if pvc.Spec.StorageClassName != nil {
							storageClass = *pvc.Spec.StorageClassName
							if sc, scOk := scMap[storageClass]; scOk {
								csiDriver = sc.Provisioner
							}
						}
						if pvc.Spec.VolumeMode != nil {
							volMode = string(*pvc.Spec.VolumeMode)
						}
						if len(pvc.Spec.AccessModes) > 0 {
							var modes []string
							for _, m := range pvc.Spec.AccessModes {
								modes = append(modes, string(m))
							}
							accessMode = strings.Join(modes, ", ")
						}
						if storageReq, ok := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; ok {
							provSizeGiB = utils.QuantityToGiB(&storageReq)
							totalDiskCapGB += provSizeGiB
						}
					}
				}
			}

			cacheMode := string(disk.Cache)
			if cacheMode == "" {
				cacheMode = "none"
			}
			ioMode := string(disk.IO)
			if ioMode == "" {
				ioMode = "default"
			}
			dedicatedIO := disk.DedicatedIOThread != nil && *disk.DedicatedIOThread

			if isCDRom {
				bootOrder := "N/A"
				if disk.BootOrder != nil {
					bootOrder = fmt.Sprintf("%d", *disk.BootOrder)
				}
				report.CD = append(report.CD, KVCDRecord{
					VM:          vm.Name,
					Powerstate:  powerState,
					Cluster:     raw.ClusterName,
					Namespace:   vm.Namespace,
					Host:        hostNode,
					DeviceNode:  disk.Name,
					SourceType:  volType,
					SourceImage: sourceImage,
					BootOrder:   bootOrder,
					Connected:   "Connected",
				})
			} else {
				report.Disk = append(report.Disk, KVDiskRecord{
					VM:                vm.Name,
					Powerstate:        powerState,
					Cluster:           raw.ClusterName,
					Namespace:         vm.Namespace,
					Host:              hostNode,
					Disk:              disk.Name,
					VolumeType:        volType,
					ClaimName:         claimName,
					StorageClass:      storageClass,
					CapacityGiB:       provSizeGiB,
					VolumeMode:        volMode,
					AccessMode:        accessMode,
					BusType:           busType,
					CacheMode:         cacheMode,
					IOMode:            ioMode,
					DedicatedIOThread: dedicatedIO,
					CSIDriver:         csiDriver,
				})
			}
		}

		// kvInfo
		report.Info = append(report.Info, KVInfoRecord{
			VM:                  vm.Name,
			Powerstate:          powerState,
			Cluster:             raw.ClusterName,
			Namespace:           vm.Namespace,
			Host:                hostNode,
			PrimaryIPAddress:    primaryIP,
			DNSName:             dnsName,
			GuestOS:             guestOS,
			Firmware:            firmware,
			EFISecureBoot:       efiSecureBoot,
			TPMEnabled:          tpmEnabled,
			CPUs:                totalVCPUs,
			CPUsSummary:         cpusSummary,
			CPUHotAddMax:        cpuHotAddMax,
			MemoryGiB:           guestRAMGiB,
			MemoryHotAddMaxGiB:  memHotAddMaxGiB,
			Disks:               disksCount,
			TotalDiskCapacityGB: totalDiskCapGB,
			NICs:                nicsCount,
			RunStrategy:         runStrategy,
			ClusterRules:        clusterRules,
			CreationDate:        creationDate,
			Uptime:              uptime,
			Annotation:          annotStr,
			Labels:              labelsStr,
			VMUUID:              string(vm.UID),
		})

		// kvNetwork
		netMap := make(map[string]virtv1.Network)
		for _, net := range vSpec.Networks {
			netMap[net.Name] = net
		}

		for _, iface := range vSpec.Domain.Devices.Interfaces {
			bindingType := "masquerade"
			if iface.Bridge != nil {
				bindingType = "bridge"
			} else if iface.SRIOV != nil {
				bindingType = "sriov"
			} else if iface.DeprecatedMacvtap != nil {
				bindingType = "macvtap"
			} else if iface.DeprecatedSlirp != nil {
				bindingType = "slirp"
			} else if iface.Binding != nil {
				bindingType = iface.Binding.Name
			}

			networkName := "pod"
			if net, ok := netMap[iface.Name]; ok {
				if net.Multus != nil {
					networkName = net.Multus.NetworkName
				}
			}

			mac := iface.MacAddress
			macType := "Manual"
			if mac == "" {
				macType = "Generated"
			}
			podIP := ""
			var guestIPs []string

			if hasVMI {
				for _, vmiIface := range vmi.Status.Interfaces {
					if vmiIface.Name == iface.Name {
						if mac == "" {
							mac = vmiIface.MAC
						}
						podIP = vmiIface.IP
						guestIPs = append(guestIPs, vmiIface.IPs...)
					}
				}
			}

			report.Network = append(report.Network, KVNetworkRecord{
				VM:               vm.Name,
				Powerstate:       powerState,
				Cluster:          raw.ClusterName,
				Namespace:        vm.Namespace,
				Host:             hostNode,
				NICLabel:         iface.Name,
				Network:          networkName,
				BindingType:      bindingType,
				MacAddress:       mac,
				MacType:          macType,
				IPv4Address:      podIP,
				GuestReportedIPs: strings.Join(guestIPs, ", "),
				AdapterModel:     iface.Model,
				PCIAddress:       iface.PciAddress,
			})
		}

		// kvHardware (vUSB / vHardware)
		for _, gpu := range vSpec.Domain.Devices.GPUs {
			report.Hardware = append(report.Hardware, KVHardwareRecord{
				VM:           vm.Name,
				Cluster:      raw.ClusterName,
				Namespace:    vm.Namespace,
				Host:         hostNode,
				DeviceType:   "GPU",
				DeviceName:   gpu.Name,
				ResourceName: gpu.DeviceName,
			})
		}
		for _, hostDev := range vSpec.Domain.Devices.HostDevices {
			report.Hardware = append(report.Hardware, KVHardwareRecord{
				VM:           vm.Name,
				Cluster:      raw.ClusterName,
				Namespace:    vm.Namespace,
				Host:         hostNode,
				DeviceType:   "HostDevice",
				DeviceName:   hostDev.Name,
				ResourceName: hostDev.DeviceName,
			})
		}

		// kvPartition (from filesystem list)
		if fsList, ok := raw.FileSystemList[key]; ok && fsList != nil {
			for _, fs := range fsList.Items {
				totalGiB := utils.BytesToGiB(int64(fs.TotalBytes))
				usedGiB := utils.BytesToGiB(int64(fs.UsedBytes))
				freeGiB := 0.0
				freePct := 0.0
				if fs.TotalBytes > 0 {
					freeBytes := fs.TotalBytes - fs.UsedBytes
					if freeBytes < 0 {
						freeBytes = 0
					}
					freeGiB = utils.BytesToGiB(int64(freeBytes))
					freePct = float64(freeBytes) / float64(fs.TotalBytes) * 100.0
				}
				report.Partition = append(report.Partition, KVPartitionRecord{
					VM:          vm.Name,
					Powerstate:  powerState,
					Cluster:     raw.ClusterName,
					Namespace:   vm.Namespace,
					Host:        hostNode,
					MountPoint:  fs.MountPoint,
					FSType:      fs.FileSystemType,
					Disk:        fs.DiskName,
					CapacityGiB: totalGiB,
					ConsumedGiB: usedGiB,
					FreeGiB:     freeGiB,
					FreePercent: utils.BytesToGiB(int64(freePct * float64(utils.GiB))),
				})
			}
		}

		// kvGuestAgent (vTools in RVTools)
		if hasVMI && vmi.Status.Phase == virtv1.Running {
			agentConnected := false
			for _, cond := range vmi.Status.Conditions {
				if cond.Type == virtv1.VirtualMachineInstanceAgentConnected && cond.Status == corev1.ConditionTrue {
					agentConnected = true
					break
				}
			}

			agentVer := ""
			guestHost := ""
			prettyName := ""
			kernelRel := ""
			tz := ""
			fsFreeze := false
			userCount := 0

			if osInfo, ok := raw.GuestOSInfo[key]; ok && osInfo != nil {
				agentVer = osInfo.GAVersion
				guestHost = osInfo.Hostname
				prettyName = osInfo.OS.PrettyName
				kernelRel = osInfo.OS.KernelRelease
				tz = osInfo.Timezone
				fsFreeze = osInfo.FSFreezeStatus != ""
				userCount = len(osInfo.UserList)
			}

			report.GuestAgent = append(report.GuestAgent, KVGuestAgentRecord{
				VM:                 vm.Name,
				Powerstate:         powerState,
				Cluster:            raw.ClusterName,
				Namespace:          vm.Namespace,
				Host:               hostNode,
				AgentConnected:     agentConnected,
				AgentVersion:       agentVer,
				GuestHostname:      guestHost,
				GuestOS:            prettyName,
				KernelRelease:      kernelRel,
				Timezone:           tz,
				FSFreezeSupported:  fsFreeze,
				LoggedInUsersCount: userCount,
			})
		}
	}

	// 3. Transform kvSnapshot
	for _, snap := range raw.VMSnapshots {
		snapName := snap.GetName()
		snapNS := snap.GetNamespace()
		sourceVM, _, _ := unstructured.NestedString(snap.Object, "spec", "source", "name")
		readyToUse, _, _ := unstructured.NestedBool(snap.Object, "status", "readyToUse")
		creationTime := snap.GetCreationTimestamp().Time
		ageDays := utils.FormatAgeDays(creationTime)

		errorReason := ""
		if errObj, found, _ := unstructured.NestedMap(snap.Object, "status", "error"); found && errObj != nil {
			if msg, ok := errObj["message"].(string); ok {
				errorReason = msg
			}
		}

		report.Snapshot = append(report.Snapshot, KVSnapshotRecord{
			SnapshotName:           snapName,
			Cluster:                raw.ClusterName,
			Namespace:              snapNS,
			SourceVM:               sourceVM,
			ReadyToUse:             readyToUse,
			CreationDate:           utils.FormatTimestamp(&creationTime),
			AgeDays:                ageDays,
			VolumeSnapshotCount:    1,
			TotalRestorableSizeGiB: 0,
			ErrorReason:            errorReason,
		})
	}

	// 4. Transform kvStoragePool (vDatastore in RVTools)
	scPVCCount := make(map[string]int)
	scAllocSize := make(map[string]float64)
	for _, pvc := range raw.PVCs {
		if pvc.Spec.StorageClassName != nil {
			scName := *pvc.Spec.StorageClassName
			scPVCCount[scName]++
			if storageReq, ok := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; ok {
				scAllocSize[scName] += utils.QuantityToGiB(&storageReq)
			}
		}
	}

	for _, sc := range raw.StorageClasses {
		reclaimPolicy := "Delete"
		if sc.ReclaimPolicy != nil {
			reclaimPolicy = string(*sc.ReclaimPolicy)
		}
		bindingMode := "Immediate"
		if sc.VolumeBindingMode != nil {
			bindingMode = string(*sc.VolumeBindingMode)
		}
		allowExpansion := sc.AllowVolumeExpansion != nil && *sc.AllowVolumeExpansion
		isDefault := false
		if val, exists := sc.Annotations["storageclass.kubernetes.io/is-default-class"]; exists && val == "true" {
			isDefault = true
		}

		report.StoragePool = append(report.StoragePool, KVStoragePoolRecord{
			StorageClassName:      sc.Name,
			Cluster:               raw.ClusterName,
			ProvisionerCSIDriver:  sc.Provisioner,
			ReclaimPolicy:         reclaimPolicy,
			VolumeBindingMode:     bindingMode,
			AllowVolumeExpansion:  allowExpansion,
			IsDefaultClass:        isDefault,
			TotalBoundPVCCount:    scPVCCount[sc.Name],
			TotalAllocatedSizeGiB: scAllocSize[sc.Name],
		})
	}

	// 5. Transform kvHealth
	report.Health = RunHealthAudit(raw, report.Node)

	// 6. Summary metrics
	runningVMs := 0
	stoppedVMs := 0
	for _, info := range report.Info {
		if info.Powerstate == "poweredOn" || info.Powerstate == "Running" {
			runningVMs++
		} else {
			stoppedVMs++
		}
	}

	readyNodes := 0
	for _, nr := range report.Node {
		if nr.Status == "Ready" {
			readyNodes++
		}
	}

	criticals := 0
	warnings := 0
	for _, hr := range report.Health {
		if hr.Severity == "CRITICAL" {
			criticals++
		} else if hr.Severity == "WARNING" {
			warnings++
		}
	}

	report.Summary = ReportSummary{
		TotalVMs:          len(report.Info),
		RunningVMs:        runningVMs,
		StoppedVMs:        stoppedVMs,
		TotalNodes:        len(report.Node),
		ReadyNodes:        readyNodes,
		TotalDisks:        len(report.Disk),
		TotalSnapshots:    len(report.Snapshot),
		HealthIssuesCount: len(report.Health),
		CriticalIssues:    criticals,
		WarningIssues:     warnings,
	}

	return report
}
