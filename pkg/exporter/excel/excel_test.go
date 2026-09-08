package excel

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/coulof/kvtools/pkg/engine"
	"github.com/xuri/excelize/v2"
)

func TestExportExcelAllSheets(t *testing.T) {
	tempDir := t.TempDir()
	excelPath := filepath.Join(tempDir, "test-inventory.xlsx")

	report := &engine.InventoryReport{
		ClusterName: "test-cluster",
		GeneratedAt: time.Now().UTC(),
		Info: []engine.KVInfoRecord{
			{
				VMName:          "vm-prod-01",
				Namespace:       "default",
				PowerState:      "Running",
				RunStrategy:     "Always",
				Node:            "node-1",
				IPAddress:       "192.168.1.10",
				GuestOS:         "Ubuntu 22.04 LTS",
				FirmwareBoot:    "UEFI",
				TPMEnabled:      true,
				CPUsSummary:     "4 Cores, 1 Socket, 1 Thread (4 vCPUs)",
				MemoryConfigGiB: 8.0,
				DisksCount:      1,
				NICsCount:       1,
				CreatedTime:     "2026-09-01T00:00:00Z",
				Uptime:          "7d 12h",
				Labels:          "app=web",
				Annotations:     "desc=production web",
				UID:             "uid-1234",
			},
		},
		CPU: []engine.KVCPURecord{
			{
				VMName:      "vm-prod-01",
				Namespace:   "default",
				Cores:       4,
				Sockets:     1,
				Threads:     1,
				TotalVCPUs:  4,
				CPUModel:    "host-model",
				CPURequests: 2.0,
				CPULimits:   4.0,
			},
		},
		Memory: []engine.KVMemoryRecord{
			{
				VMName:            "vm-prod-01",
				Namespace:         "default",
				GuestRAMGiB:       8.0,
				MemoryRequestsGiB: 8.0,
				MemoryLimitsGiB:   8.0,
			},
		},
		Disk: []engine.KVDiskRecord{
			{
				VMName:             "vm-prod-01",
				Namespace:          "default",
				DiskTargetName:     "disk0",
				VolumeType:         "PersistentVolumeClaim",
				ClaimName:          "pvc-root",
				StorageClass:       "longhorn",
				ProvisionedSizeGiB: 40.0,
				VolumeMode:         "Block",
				AccessMode:         "ReadWriteMany",
				BusType:            "virtio",
			},
		},
		Partition: []engine.KVPartitionRecord{
			{
				VMName:           "vm-prod-01",
				Namespace:        "default",
				MountPoint:       "/",
				FSType:           "ext4",
				TotalCapacityGiB: 40.0,
				UsedSpaceGiB:     12.0,
				FreeSpaceGiB:     28.0,
				FreePercent:      70.0,
			},
		},
		Network: []engine.KVNetworkRecord{
			{
				VMName:        "vm-prod-01",
				Namespace:     "default",
				InterfaceName: "nic0",
				NetworkName:   "pod",
				BindingType:   "masquerade",
				MACAddress:    "52:54:00:ab:cd:ef",
				PodIP:         "10.244.0.15",
			},
		},
		CD: []engine.KVCDRecord{
			{
				VMName:         "vm-prod-01",
				Namespace:      "default",
				CDDeviceName:   "cdrom0",
				SourceType:     "ContainerDisk",
				ConnectedState: "Connected",
			},
		},
		Snapshot: []engine.KVSnapshotRecord{
			{
				SnapshotName: "snap-backup-1",
				Namespace:    "default",
				SourceVM:     "vm-prod-01",
				ReadyToUse:   true,
				AgeDays:      3,
			},
		},
		GuestAgent: []engine.KVGuestAgentRecord{
			{
				VMName:            "vm-prod-01",
				Namespace:         "default",
				AgentConnected:    true,
				AgentVersion:      "8.0.0",
				GuestHostname:     "web-prod",
				GuestOSPrettyName: "Ubuntu 22.04 LTS",
			},
		},
		Node: []engine.KVNodeRecord{
			{
				NodeName:            "node-1",
				Status:              "Ready",
				TotalPhysicalCores:  16,
				TotalRAMGiB:         64.0,
				AllocatableCPU:      15.0,
				AllocatableRAMGiB:   60.0,
				AllocatedVMvCPUs:    4.0,
				AllocatedVMRAMGiB:   8.0,
				VCPUOvercommitRatio: 0.27,
				ActiveVMCount:       1,
				KVMHardwareAccel:    true,
			},
		},
		StoragePool: []engine.KVStoragePoolRecord{
			{
				StorageClassName:     "longhorn",
				ProvisionerCSIDriver: "driver.longhorn.io",
				ReclaimPolicy:        "Delete",
				VolumeBindingMode:    "Immediate",
				TotalBoundPVCCount:   1,
			},
		},
		Hardware: []engine.KVHardwareRecord{
			{
				VMName:       "vm-prod-01",
				Namespace:    "default",
				DeviceType:   "GPU",
				DeviceName:   "gpu0",
				ResourceName: "nvidia.com/A100",
			},
		},
		Health: []engine.KVHealthRecord{
			{
				RuleID:       "HLTH-007",
				Category:     "Guest Visibility",
				Severity:     "WARNING",
				ResourceKind: "VirtualMachine",
				ResourceName: "vm-prod-01",
				Namespace:    "default",
				IssueSummary: "Guest agent not reporting",
				Remediation:  "Install qemu-guest-agent",
			},
		},
	}

	err := ExportExcel(report, excelPath)
	if err != nil {
		t.Fatalf("failed to export Excel: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(excelPath); os.IsNotExist(err) {
		t.Fatalf("Excel file was not created: %s", excelPath)
	}

	// Read file back using excelize
	f, err := excelize.OpenFile(excelPath)
	if err != nil {
		t.Fatalf("failed to read back created Excel file: %v", err)
	}
	defer f.Close()

	expectedSheets := []string{
		"kvInfo", "kvCPU", "kvMemory", "kvDisk", "kvPartition",
		"kvNetwork", "kvCD", "kvSnapshot", "kvGuestAgent",
		"kvNode", "kvStoragePool", "kvHardware", "kvHealth",
	}

	sheetList := f.GetSheetList()
	sheetMap := make(map[string]bool)
	for _, s := range sheetList {
		sheetMap[s] = true
	}

	for _, expected := range expectedSheets {
		if !sheetMap[expected] {
			t.Errorf("missing expected sheet '%s' in Excel file", expected)
		}
	}

	// Verify kvInfo has row data
	cellVal, err := f.GetCellValue("kvInfo", "A2")
	if err != nil || cellVal != "vm-prod-01" {
		t.Errorf("expected cell A2 on kvInfo to be 'vm-prod-01', got '%s', err: %v", cellVal, err)
	}
}
