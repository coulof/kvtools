package engine

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	virtv1 "kubevirt.io/api/core/v1"

	"github.com/coulof/kvtools/pkg/collector"
)

func TestTransformerFullReport(t *testing.T) {
	node1 := corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-worker-01",
			Labels: map[string]string{
				"kubevirt.io/schedulable": "true",
			},
		},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
			},
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("32"),
				corev1.ResourceMemory: resource.MustParse("128Gi"),
			},
			Allocatable: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("30"),
				corev1.ResourceMemory: resource.MustParse("120Gi"),
			},
			NodeInfo: corev1.NodeSystemInfo{
				Architecture:   "amd64",
				KubeletVersion: "v1.30.2",
				OSImage:        "SLE-Micro 5.5",
				KernelVersion:  "5.14.21",
			},
		},
	}

	sc := storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "longhorn-fast",
			Annotations: map[string]string{
				"storageclass.kubernetes.io/is-default-class": "true",
			},
		},
		Provisioner: "driver.longhorn.io",
	}

	scName := "longhorn-fast"
	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "win2022-disk-0",
			Namespace: "vms",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &scName,
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("50Gi"),
				},
			},
		},
	}

	bootOrder := uint(1)
	vm := virtv1.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "win-server-2022",
			Namespace: "vms",
			UID:       types.UID("vm-uuid-12345"),
			CreationTimestamp: metav1.NewTime(time.Now().Add(-48 * time.Hour)),
		},
		Spec: virtv1.VirtualMachineSpec{
			Template: &virtv1.VirtualMachineInstanceTemplateSpec{
				Spec: virtv1.VirtualMachineInstanceSpec{
					Domain: virtv1.DomainSpec{
						CPU: &virtv1.CPU{
							Cores:   4,
							Sockets: 2,
							Threads: 1,
							Model:   "host-model",
						},
						Memory: &virtv1.Memory{
							Guest: resourcePtr(resource.MustParse("16Gi")),
						},
						Devices: virtv1.Devices{
							Disks: []virtv1.Disk{
								{
									Name:      "disk0",
									BootOrder: &bootOrder,
									DiskDevice: virtv1.DiskDevice{
										Disk: &virtv1.DiskTarget{
											Bus: "virtio",
										},
									},
								},
								{
									Name: "cdrom0",
									DiskDevice: virtv1.DiskDevice{
										CDRom: &virtv1.CDRomTarget{
											Bus: "sata",
										},
									},
								},
							},
							Interfaces: []virtv1.Interface{
								{
									Name:       "nic0",
									MacAddress: "52:54:00:12:34:56",
									InterfaceBindingMethod: virtv1.InterfaceBindingMethod{
										Bridge: &virtv1.InterfaceBridge{},
									},
								},
							},
							GPUs: []virtv1.GPU{
								{
									Name:       "gpu0",
									DeviceName: "nvidia.com/A40",
								},
							},
						},
					},
					Volumes: []virtv1.Volume{
						{
							Name: "disk0",
							VolumeSource: virtv1.VolumeSource{
								PersistentVolumeClaim: &virtv1.PersistentVolumeClaimVolumeSource{
									PersistentVolumeClaimVolumeSource: corev1.PersistentVolumeClaimVolumeSource{
										ClaimName: "win2022-disk-0",
									},
								},
							},
						},
						{
							Name: "cdrom0",
							VolumeSource: virtv1.VolumeSource{
								ContainerDisk: &virtv1.ContainerDiskSource{
									Image: "registry.example.com/iso/virtio-win:latest",
								},
							},
						},
					},
					Networks: []virtv1.Network{
						{
							Name: "nic0",
							NetworkSource: virtv1.NetworkSource{
								Pod: &virtv1.PodNetwork{},
							},
						},
					},
				},
			},
		},
	}

	vmi := virtv1.VirtualMachineInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "win-server-2022",
			Namespace: "vms",
			UID:       types.UID("vm-uuid-12345"),
			CreationTimestamp: metav1.NewTime(time.Now().Add(-24 * time.Hour)),
		},
		Spec: vm.Spec.Template.Spec,
		Status: virtv1.VirtualMachineInstanceStatus{
			Phase:    virtv1.Running,
			NodeName: "node-worker-01",
			Interfaces: []virtv1.VirtualMachineInstanceNetworkInterface{
				{
					Name: "nic0",
					IP:   "10.244.1.45",
					IPs:  []string{"10.244.1.45"},
					MAC:  "52:54:00:12:34:56",
				},
			},
			Conditions: []virtv1.VirtualMachineInstanceCondition{
				{
					Type:   virtv1.VirtualMachineInstanceAgentConnected,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}

	agentInfo := &virtv1.VirtualMachineInstanceGuestAgentInfo{
		GAVersion: "7.2.0",
		Hostname:  "win-server-prod",
		OS: virtv1.VirtualMachineInstanceGuestOSInfo{
			PrettyName:    "Windows Server 2022 Datacenter",
			KernelRelease: "10.0.20348",
		},
		Timezone: "UTC",
	}

	fsList := &virtv1.VirtualMachineInstanceFileSystemList{
		Items: []virtv1.VirtualMachineInstanceFileSystem{
			{
				MountPoint:     "C:\\",
				FileSystemType: "NTFS",
				TotalBytes:     50 * 1024 * 1024 * 1024,
				UsedBytes:      20 * 1024 * 1024 * 1024,
			},
		},
	}

	snap := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "snapshot.kubevirt.io/v1alpha1",
			"kind":       "VirtualMachineSnapshot",
			"metadata": map[string]interface{}{
				"name":      "win-snap-daily",
				"namespace": "vms",
			},
			"spec": map[string]interface{}{
				"source": map[string]interface{}{
					"name": "win-server-2022",
				},
			},
			"status": map[string]interface{}{
				"readyToUse": true,
			},
		},
	}

	rawData := &collector.RawData{
		ClusterName:    "harvester-prod",
		CollectedAt:    time.Now().UTC(),
		VMs:            []virtv1.VirtualMachine{vm},
		VMIs:           []virtv1.VirtualMachineInstance{vmi},
		Nodes:          []corev1.Node{node1},
		StorageClasses: []storagev1.StorageClass{sc},
		PVCs:           []corev1.PersistentVolumeClaim{pvc},
		VMSnapshots:    []unstructured.Unstructured{snap},
		GuestOSInfo: map[string]*virtv1.VirtualMachineInstanceGuestAgentInfo{
			"vms/win-server-2022": agentInfo,
		},
		FileSystemList: map[string]*virtv1.VirtualMachineInstanceFileSystemList{
			"vms/win-server-2022": fsList,
		},
	}

	report := Transform(rawData)

	// Verify Summary
	if report.Summary.TotalVMs != 1 || report.Summary.RunningVMs != 1 {
		t.Errorf("expected 1 total & 1 running VM, got total=%d running=%d", report.Summary.TotalVMs, report.Summary.RunningVMs)
	}

	// Verify kvInfo
	if len(report.Info) != 1 {
		t.Fatalf("expected 1 Info record, got %d", len(report.Info))
	}
	info := report.Info[0]
	if info.VMName != "win-server-2022" || info.PowerState != "Running" || info.GuestOS != "Windows Server 2022 Datacenter" {
		t.Errorf("unexpected Info values: %+v", info)
	}

	// Verify kvCPU
	if len(report.CPU) != 1 {
		t.Fatalf("expected 1 CPU record, got %d", len(report.CPU))
	}
	cpu := report.CPU[0]
	if cpu.TotalVCPUs != 8 || cpu.Cores != 4 || cpu.Sockets != 2 {
		t.Errorf("expected 8 vCPUs (4 cores * 2 sockets), got %d", cpu.TotalVCPUs)
	}

	// Verify kvMemory
	if len(report.Memory) != 1 || report.Memory[0].GuestRAMGiB != 16.0 {
		t.Errorf("expected 16 GiB RAM, got %+v", report.Memory)
	}

	// Verify kvDisk & kvCD
	if len(report.Disk) != 1 || report.Disk[0].DiskTargetName != "disk0" {
		t.Errorf("expected 1 Disk record (disk0), got %+v", report.Disk)
	}
	if len(report.CD) != 1 || report.CD[0].CDDeviceName != "cdrom0" {
		t.Errorf("expected 1 CD record (cdrom0), got %+v", report.CD)
	}

	// Verify kvNetwork
	if len(report.Network) != 1 || report.Network[0].BindingType != "bridge" {
		t.Errorf("expected 1 Network record with bridge binding, got %+v", report.Network)
	}

	// Verify kvHardware
	if len(report.Hardware) != 1 || report.Hardware[0].DeviceType != "GPU" {
		t.Errorf("expected 1 GPU hardware record, got %+v", report.Hardware)
	}

	// Verify kvPartition
	if len(report.Partition) != 1 || report.Partition[0].MountPoint != "C:\\" {
		t.Errorf("expected 1 Partition record for C:\\, got %+v", report.Partition)
	}

	// Verify kvGuestAgent
	if len(report.GuestAgent) != 1 || !report.GuestAgent[0].AgentConnected {
		t.Errorf("expected 1 GuestAgent record with connected agent, got %+v", report.GuestAgent)
	}

	// Verify kvNode
	if len(report.Node) != 1 || report.Node[0].ActiveVMCount != 1 || report.Node[0].AllocatedVMvCPUs != 8 {
		t.Errorf("expected 1 Node record with 1 VM and 8 vCPUs allocated, got %+v", report.Node)
	}

	// Verify kvStoragePool
	if len(report.StoragePool) != 1 || report.StoragePool[0].TotalBoundPVCCount != 1 {
		t.Errorf("expected 1 StoragePool record with 1 bound PVC, got %+v", report.StoragePool)
	}

	// Verify kvSnapshot
	if len(report.Snapshot) != 1 || report.Snapshot[0].SnapshotName != "win-snap-daily" {
		t.Errorf("expected 1 Snapshot record, got %+v", report.Snapshot)
	}
}

func resourcePtr(q resource.Quantity) *resource.Quantity {
	return &q
}
