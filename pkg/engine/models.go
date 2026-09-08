package engine

import "time"

// InventoryReport is the top-level container holding transformed records for all 13 sheets.
type InventoryReport struct {
	ClusterName  string                 `json:"clusterName"`
	GeneratedAt  time.Time              `json:"generatedAt"`
	Summary      ReportSummary          `json:"summary"`
	Info         []KVInfoRecord         `json:"info"`
	CPU          []KVCPURecord          `json:"cpu"`
	Memory       []KVMemoryRecord       `json:"memory"`
	Disk         []KVDiskRecord         `json:"disk"`
	Partition    []KVPartitionRecord    `json:"partition"`
	Network      []KVNetworkRecord      `json:"network"`
	CD           []KVCDRecord           `json:"cd"`
	Snapshot     []KVSnapshotRecord     `json:"snapshot"`
	GuestAgent   []KVGuestAgentRecord   `json:"guestAgent"`
	Node         []KVNodeRecord         `json:"node"`
	StoragePool  []KVStoragePoolRecord  `json:"storagePool"`
	Hardware     []KVHardwareRecord     `json:"hardware"`
	Health       []KVHealthRecord       `json:"health"`
}

// ReportSummary contains high-level cluster virtualization metrics.
type ReportSummary struct {
	TotalVMs          int `json:"totalVMs"`
	RunningVMs        int `json:"runningVMs"`
	StoppedVMs        int `json:"stoppedVMs"`
	TotalNodes        int `json:"totalNodes"`
	ReadyNodes        int `json:"readyNodes"`
	TotalDisks        int `json:"totalDisks"`
	TotalSnapshots    int `json:"totalSnapshots"`
	HealthIssuesCount int `json:"healthIssuesCount"`
	CriticalIssues    int `json:"criticalIssues"`
	WarningIssues     int `json:"warningIssues"`
}

// KVInfoRecord represents a row in Sheet 1 (kvInfo).
type KVInfoRecord struct {
	VMName              string  `json:"vmName"`
	Namespace           string  `json:"namespace"`
	PowerState          string  `json:"powerState"`
	RunStrategy         string  `json:"runStrategy"`
	Node                string  `json:"node"`
	IPAddress           string  `json:"ipAddress"`
	GuestOS             string  `json:"guestOS"`
	FirmwareBoot        string  `json:"firmwareBoot"`
	TPMEnabled          bool    `json:"tpmEnabled"`
	CPUsSummary         string  `json:"cpusSummary"`
	CPUHotplugMax       uint32  `json:"cpuHotplugMax"`
	MemoryConfigGiB     float64 `json:"memoryConfigGiB"`
	MemoryHotplugMaxGiB float64 `json:"memoryHotplugMaxGiB"`
	DisksCount          int     `json:"disksCount"`
	NICsCount           int     `json:"nicsCount"`
	AffinityRules       string  `json:"affinityRules"`
	CreatedTime         string  `json:"createdTime"`
	Uptime              string  `json:"uptime"`
	Labels              string  `json:"labels"`
	Annotations         string  `json:"annotations"`
	UID                 string  `json:"uid"`
}

// KVCPURecord represents a row in Sheet 2 (kvCPU).
type KVCPURecord struct {
	VMName                string  `json:"vmName"`
	Namespace             string  `json:"namespace"`
	Cores                 uint32  `json:"cores"`
	Sockets               uint32  `json:"sockets"`
	Threads               uint32  `json:"threads"`
	TotalVCPUs            uint32  `json:"totalVCPUs"`
	CPUModel              string  `json:"cpuModel"`
	DedicatedCPUPlacement bool    `json:"dedicatedCPUPlacement"`
	NUMANodes             int     `json:"numaNodes"`
	CPURequests           float64 `json:"cpuRequests"`
	CPULimits             float64 `json:"cpuLimits"`
	HostNodeCPUModel      string  `json:"hostNodeCPUModel"`
}

// KVMemoryRecord represents a row in Sheet 3 (kvMemory).
type KVMemoryRecord struct {
	VMName                string  `json:"vmName"`
	Namespace             string  `json:"namespace"`
	GuestRAMGiB           float64 `json:"guestRAMGiB"`
	MemoryRequestsGiB     float64 `json:"memoryRequestsGiB"`
	MemoryLimitsGiB       float64 `json:"memoryLimitsGiB"`
	LauncherOverheadMiB   float64 `json:"launcherOverheadMiB"`
	Hugepages             string  `json:"hugepages"`
	AutoattachMemBalloon  bool    `json:"autoattachMemBalloon"`
	MemoryDumpEnabled     bool    `json:"memoryDumpEnabled"`
}

// KVDiskRecord represents a row in Sheet 4 (kvDisk).
type KVDiskRecord struct {
	VMName            string  `json:"vmName"`
	Namespace         string  `json:"namespace"`
	DiskTargetName    string  `json:"diskTargetName"`
	VolumeType        string  `json:"volumeType"`
	ClaimName         string  `json:"claimName"`
	StorageClass      string  `json:"storageClass"`
	ProvisionedSizeGiB float64 `json:"provisionedSizeGiB"`
	VolumeMode        string  `json:"volumeMode"`
	AccessMode        string  `json:"accessMode"`
	BusType           string  `json:"busType"`
	CacheMode         string  `json:"cacheMode"`
	IOMode            string  `json:"ioMode"`
	DedicatedIOThread bool    `json:"dedicatedIOThread"`
	CSIDriver         string  `json:"csiDriver"`
}

// KVPartitionRecord represents a row in Sheet 5 (kvPartition).
type KVPartitionRecord struct {
	VMName          string  `json:"vmName"`
	Namespace       string  `json:"namespace"`
	MountPoint      string  `json:"mountPoint"`
	FSType          string  `json:"fsType"`
	DiskName        string  `json:"diskName"`
	TotalCapacityGiB float64 `json:"totalCapacityGiB"`
	UsedSpaceGiB    float64 `json:"usedSpaceGiB"`
	FreeSpaceGiB    float64 `json:"freeSpaceGiB"`
	FreePercent     float64 `json:"freePercent"`
}

// KVNetworkRecord represents a row in Sheet 6 (kvNetwork).
type KVNetworkRecord struct {
	VMName           string `json:"vmName"`
	Namespace        string `json:"namespace"`
	InterfaceName    string `json:"interfaceName"`
	NetworkName      string `json:"networkName"`
	BindingType      string `json:"bindingType"`
	MACAddress       string `json:"macAddress"`
	PodIP            string `json:"podIP"`
	GuestReportedIPs string `json:"guestReportedIPs"`
	InterfaceModel   string `json:"interfaceModel"`
	PCIAddress       string `json:"pciAddress"`
}

// KVCDRecord represents a row in Sheet 7 (kvCD).
type KVCDRecord struct {
	VMName         string `json:"vmName"`
	Namespace      string `json:"namespace"`
	CDDeviceName   string `json:"cdDeviceName"`
	SourceType     string `json:"sourceType"`
	SourceImage    string `json:"sourceImage"`
	BootOrder      string `json:"bootOrder"`
	ConnectedState string `json:"connectedState"`
}

// KVSnapshotRecord represents a row in Sheet 8 (kvSnapshot).
type KVSnapshotRecord struct {
	SnapshotName          string  `json:"snapshotName"`
	Namespace             string  `json:"namespace"`
	SourceVM              string  `json:"sourceVM"`
	ReadyToUse            bool    `json:"readyToUse"`
	CreationTimestamp     string  `json:"creationTimestamp"`
	AgeDays               int     `json:"ageDays"`
	VolumeSnapshotCount   int     `json:"volumeSnapshotCount"`
	TotalRestorableSizeGiB float64 `json:"totalRestorableSizeGiB"`
	ErrorReason           string  `json:"errorReason"`
}

// KVGuestAgentRecord represents a row in Sheet 9 (kvGuestAgent).
type KVGuestAgentRecord struct {
	VMName             string `json:"vmName"`
	Namespace          string `json:"namespace"`
	AgentConnected     bool   `json:"agentConnected"`
	AgentVersion       string `json:"agentVersion"`
	GuestHostname      string `json:"guestHostname"`
	GuestOSPrettyName  string `json:"guestOSPrettyName"`
	GuestKernelRelease string `json:"guestKernelRelease"`
	Timezone           string `json:"timezone"`
	FSFreezeSupported  bool   `json:"fsFreezeSupported"`
	LoggedInUsersCount int    `json:"loggedInUsersCount"`
}

// KVNodeRecord represents a row in Sheet 10 (kvNode).
type KVNodeRecord struct {
	NodeName            string  `json:"nodeName"`
	Status              string  `json:"status"`
	TotalPhysicalCores  int64   `json:"totalPhysicalCores"`
	TotalRAMGiB         float64 `json:"totalRAMGiB"`
	AllocatableCPU      float64 `json:"allocatableCPU"`
	AllocatableRAMGiB   float64 `json:"allocatableRAMGiB"`
	AllocatedVMvCPUs    float64 `json:"allocatedVMvCPUs"`
	AllocatedVMRAMGiB   float64 `json:"allocatedVMRAMGiB"`
	VCPUOvercommitRatio float64 `json:"vcpuOvercommitRatio"`
	ActiveVMCount       int     `json:"activeVMCount"`
	KVMHardwareAccel    bool    `json:"kvmHardwareAccel"`
	KubernetesVersion   string  `json:"kubernetesVersion"`
	OSImage             string  `json:"osImage"`
	KernelVersion       string  `json:"kernelVersion"`
	NodeTaints          string  `json:"nodeTaints"`
	NodeConditions      string  `json:"nodeConditions"`
}

// KVStoragePoolRecord represents a row in Sheet 11 (kvStoragePool).
type KVStoragePoolRecord struct {
	StorageClassName      string  `json:"storageClassName"`
	ProvisionerCSIDriver  string  `json:"provisionerCSIDriver"`
	ReclaimPolicy         string  `json:"reclaimPolicy"`
	VolumeBindingMode     string  `json:"volumeBindingMode"`
	AllowVolumeExpansion  bool    `json:"allowVolumeExpansion"`
	IsDefaultClass        bool    `json:"isDefaultClass"`
	TotalBoundPVCCount    int     `json:"totalBoundPVCCount"`
	TotalAllocatedSizeGiB float64 `json:"totalAllocatedSizeGiB"`
}

// KVHardwareRecord represents a row in Sheet 12 (kvHardware).
type KVHardwareRecord struct {
	VMName       string `json:"vmName"`
	Namespace    string `json:"namespace"`
	DeviceType   string `json:"deviceType"`
	DeviceName   string `json:"deviceName"`
	ResourceName string `json:"resourceName"`
	AssignedNode string `json:"assignedNode"`
}

// KVHealthRecord represents a row in Sheet 13 (kvHealth).
type KVHealthRecord struct {
	RuleID         string `json:"ruleId"`
	Category       string `json:"category"`
	Severity       string `json:"severity"` // "CRITICAL", "WARNING", "INFO"
	ResourceKind   string `json:"resourceKind"`
	ResourceName   string `json:"resourceName"`
	Namespace      string `json:"namespace"`
	IssueSummary   string `json:"issueSummary"`
	Remediation    string `json:"remediation"`
}
