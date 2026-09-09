package engine

import "time"

// InventoryReport is the top-level container holding transformed records for all 13 sheets.
type InventoryReport struct {
	ClusterName string                `json:"clusterName"`
	GeneratedAt time.Time             `json:"generatedAt"`
	Summary     ReportSummary         `json:"summary"`
	Info        []KVInfoRecord        `json:"info"`
	CPU         []KVCPURecord         `json:"cpu"`
	Memory      []KVMemoryRecord      `json:"memory"`
	Disk        []KVDiskRecord        `json:"disk"`
	Partition   []KVPartitionRecord   `json:"partition"`
	Network     []KVNetworkRecord     `json:"network"`
	CD          []KVCDRecord          `json:"cd"`
	Snapshot    []KVSnapshotRecord    `json:"snapshot"`
	GuestAgent  []KVGuestAgentRecord  `json:"guestAgent"`
	Node        []KVNodeRecord        `json:"node"`
	StoragePool []KVStoragePoolRecord `json:"storagePool"`
	Hardware    []KVHardwareRecord    `json:"hardware"`
	Health      []KVHealthRecord      `json:"health"`
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

// KVInfoRecord represents a row in Sheet 1 (kvInfo / vInfo).
type KVInfoRecord struct {
	VM                  string  `json:"vm"`
	Powerstate          string  `json:"powerstate"`
	Cluster             string  `json:"cluster"`
	Namespace           string  `json:"namespace"`
	Host                string  `json:"host"`
	PrimaryIPAddress    string  `json:"primaryIPAddress"`
	DNSName             string  `json:"dnsName"`
	GuestOS             string  `json:"guestOS"`
	Firmware            string  `json:"firmware"`
	EFISecureBoot       bool    `json:"efiSecureBoot"`
	TPMEnabled          bool    `json:"tpmEnabled"`
	CPUs                uint32  `json:"cpus"`
	CPUsSummary         string  `json:"cpusSummary"`
	CPUHotAddMax        uint32  `json:"cpuHotAddMax"`
	MemoryGiB           float64 `json:"memoryGiB"`
	MemoryHotAddMaxGiB  float64 `json:"memoryHotAddMaxGiB"`
	Disks               int     `json:"disks"`
	TotalDiskCapacityGB float64 `json:"totalDiskCapacityGB"`
	NICs                int     `json:"nics"`
	RunStrategy         string  `json:"runStrategy"`
	ClusterRules        string  `json:"clusterRules"`
	CreationDate        string  `json:"creationDate"`
	Uptime              string  `json:"uptime"`
	Annotation          string  `json:"annotation"`
	Labels              string  `json:"labels"`
	VMUUID              string  `json:"vmUUID"`
}

// KVCPURecord represents a row in Sheet 2 (kvCPU / vCpu).
type KVCPURecord struct {
	VM                    string  `json:"vm"`
	Powerstate            string  `json:"powerstate"`
	Cluster               string  `json:"cluster"`
	Namespace             string  `json:"namespace"`
	Host                  string  `json:"host"`
	CPUs                  uint32  `json:"cpus"`
	Sockets               uint32  `json:"sockets"`
	CoresPerSocket        uint32  `json:"coresPerSocket"`
	Threads               uint32  `json:"threads"`
	CPUModel              string  `json:"cpuModel"`
	DedicatedCPUPlacement bool    `json:"dedicatedCPUPlacement"`
	NUMANodes             int     `json:"numaNodes"`
	CPURequests           float64 `json:"cpuRequests"`
	CPULimits             float64 `json:"cpuLimits"`
	HostNodeCPUModel      string  `json:"hostNodeCPUModel"`
}

// KVMemoryRecord represents a row in Sheet 3 (kvMemory / vMemory).
type KVMemoryRecord struct {
	VM                   string  `json:"vm"`
	Powerstate           string  `json:"powerstate"`
	Cluster              string  `json:"cluster"`
	Namespace            string  `json:"namespace"`
	Host                 string  `json:"host"`
	SizeGiB              float64 `json:"sizeGiB"`
	MemoryRequestsGiB    float64 `json:"memoryRequestsGiB"`
	MemoryLimitsGiB      float64 `json:"memoryLimitsGiB"`
	LauncherOverheadMiB  float64 `json:"launcherOverheadMiB"`
	Hugepages            string  `json:"hugepages"`
	AutoattachMemBalloon bool    `json:"autoattachMemBalloon"`
	MemoryDumpEnabled    bool    `json:"memoryDumpEnabled"`
}

// KVDiskRecord represents a row in Sheet 4 (kvDisk / vDisk).
type KVDiskRecord struct {
	VM                string  `json:"vm"`
	Powerstate        string  `json:"powerstate"`
	Cluster           string  `json:"cluster"`
	Namespace         string  `json:"namespace"`
	Host              string  `json:"host"`
	Disk              string  `json:"disk"`
	VolumeType        string  `json:"volumeType"`
	ClaimName         string  `json:"claimName"`
	StorageClass      string  `json:"storageClass"`
	CapacityGiB       float64 `json:"capacityGiB"`
	VolumeMode        string  `json:"volumeMode"`
	AccessMode        string  `json:"accessMode"`
	BusType           string  `json:"busType"`
	CacheMode         string  `json:"cacheMode"`
	IOMode            string  `json:"ioMode"`
	DedicatedIOThread bool    `json:"dedicatedIOThread"`
	CSIDriver         string  `json:"csiDriver"`
}

// KVPartitionRecord represents a row in Sheet 5 (kvPartition / vPartition).
type KVPartitionRecord struct {
	VM          string  `json:"vm"`
	Powerstate  string  `json:"powerstate"`
	Cluster     string  `json:"cluster"`
	Namespace   string  `json:"namespace"`
	Host        string  `json:"host"`
	MountPoint  string  `json:"mountPoint"`
	FSType      string  `json:"fsType"`
	Disk        string  `json:"disk"`
	CapacityGiB float64 `json:"capacityGiB"`
	ConsumedGiB float64 `json:"consumedGiB"`
	FreeGiB     float64 `json:"freeGiB"`
	FreePercent float64 `json:"freePercent"`
}

// KVNetworkRecord represents a row in Sheet 6 (kvNetwork / vNetwork).
type KVNetworkRecord struct {
	VM               string `json:"vm"`
	Powerstate       string `json:"powerstate"`
	Cluster          string `json:"cluster"`
	Namespace        string `json:"namespace"`
	Host             string `json:"host"`
	NICLabel         string `json:"nicLabel"`
	Network          string `json:"network"`
	BindingType      string `json:"bindingType"`
	MacAddress       string `json:"macAddress"`
	MacType          string `json:"macType"`
	IPv4Address      string `json:"ipv4Address"`
	GuestReportedIPs string `json:"guestReportedIPs"`
	AdapterModel     string `json:"adapterModel"`
	PCIAddress       string `json:"pciAddress"`
}

// KVCDRecord represents a row in Sheet 7 (kvCD / vCD).
type KVCDRecord struct {
	VM          string `json:"vm"`
	Powerstate  string `json:"powerstate"`
	Cluster     string `json:"cluster"`
	Namespace   string `json:"namespace"`
	Host        string `json:"host"`
	DeviceNode  string `json:"deviceNode"`
	SourceType  string `json:"sourceType"`
	SourceImage string `json:"sourceImage"`
	BootOrder   string `json:"bootOrder"`
	Connected   string `json:"connected"`
}

// KVSnapshotRecord represents a row in Sheet 8 (kvSnapshot / vSnapshot).
type KVSnapshotRecord struct {
	SnapshotName           string  `json:"snapshotName"`
	Cluster                string  `json:"cluster"`
	Namespace              string  `json:"namespace"`
	SourceVM               string  `json:"sourceVM"`
	ReadyToUse             bool    `json:"readyToUse"`
	CreationDate           string  `json:"creationDate"`
	AgeDays                int     `json:"ageDays"`
	VolumeSnapshotCount    int     `json:"volumeSnapshotCount"`
	TotalRestorableSizeGiB float64 `json:"totalRestorableSizeGiB"`
	ErrorReason            string  `json:"errorReason"`
}

// KVGuestAgentRecord represents a row in Sheet 9 (kvGuestAgent / vTools).
type KVGuestAgentRecord struct {
	VM                 string `json:"vm"`
	Powerstate         string `json:"powerstate"`
	Cluster            string `json:"cluster"`
	Namespace          string `json:"namespace"`
	Host               string `json:"host"`
	AgentConnected     bool   `json:"agentConnected"`
	AgentVersion       string `json:"agentVersion"`
	GuestHostname      string `json:"guestHostname"`
	GuestOS            string `json:"guestOS"`
	KernelRelease      string `json:"kernelRelease"`
	Timezone           string `json:"timezone"`
	FSFreezeSupported  bool   `json:"fsFreezeSupported"`
	LoggedInUsersCount int    `json:"loggedInUsersCount"`
}

// KVNodeRecord represents a row in Sheet 10 (kvNode / vHost).
type KVNodeRecord struct {
	Host                string  `json:"host"`
	Cluster             string  `json:"cluster"`
	Status              string  `json:"status"`
	PhysicalCores       int64   `json:"physicalCores"`
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

// KVStoragePoolRecord represents a row in Sheet 11 (kvStoragePool / vDatastore).
type KVStoragePoolRecord struct {
	StorageClassName      string  `json:"storageClassName"`
	Cluster               string  `json:"cluster"`
	ProvisionerCSIDriver  string  `json:"provisionerCSIDriver"`
	ReclaimPolicy         string  `json:"reclaimPolicy"`
	VolumeBindingMode     string  `json:"volumeBindingMode"`
	AllowVolumeExpansion  bool    `json:"allowVolumeExpansion"`
	IsDefaultClass        bool    `json:"isDefaultClass"`
	TotalBoundPVCCount    int     `json:"totalBoundPVCCount"`
	TotalAllocatedSizeGiB float64 `json:"totalAllocatedSizeGiB"`
}

// KVHardwareRecord represents a row in Sheet 12 (kvHardware / vUSB / vHardware).
type KVHardwareRecord struct {
	VM           string `json:"vm"`
	Cluster      string `json:"cluster"`
	Namespace    string `json:"namespace"`
	Host         string `json:"host"`
	DeviceType   string `json:"deviceType"`
	DeviceName   string `json:"deviceName"`
	ResourceName string `json:"resourceName"`
}

// KVHealthRecord represents a row in Sheet 13 (kvHealth / vHealth).
type KVHealthRecord struct {
	RuleID       string `json:"ruleId"`
	Severity     string `json:"severity"` // "CRITICAL", "WARNING", "INFO"
	Category     string `json:"category"`
	ResourceKind string `json:"resourceKind"`
	ResourceName string `json:"resourceName"`
	Cluster      string `json:"cluster"`
	Namespace    string `json:"namespace"`
	IssueSummary string `json:"issueSummary"`
	Remediation  string `json:"remediation"`
}
