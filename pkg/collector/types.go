package collector

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	virtv1 "kubevirt.io/api/core/v1"
)

// RawData encapsulates all raw Kubernetes and KubeVirt objects retrieved from the cluster.
type RawData struct {
	ClusterName                 string
	CollectedAt                 time.Time
	VMs                         []virtv1.VirtualMachine
	VMIs                        []virtv1.VirtualMachineInstance
	GuestOSInfo                 map[string]*virtv1.VirtualMachineInstanceGuestAgentInfo // keyed by "namespace/name"
	FileSystemList              map[string]*virtv1.VirtualMachineInstanceFileSystemList // keyed by "namespace/name"
	Nodes                       []corev1.Node
	PVCs                        []corev1.PersistentVolumeClaim
	PVs                         []corev1.PersistentVolume
	StorageClasses              []storagev1.StorageClass
	DataVolumes                 []unstructured.Unstructured
	VMSnapshots                 []unstructured.Unstructured
	VolumeSnapshots             []unstructured.Unstructured
	NetworkAttachmentDefinitions []unstructured.Unstructured
	Warnings                    []string
}

// CollectorOptions defines configuration parameters for data collection.
type CollectorOptions struct {
	Namespace               string
	AllNamespaces           bool
	FetchGuestSubresources  bool
	Concurrency             int
	Timeout                 time.Duration
}

// DefaultCollectorOptions returns sensible defaults for data collection.
func DefaultCollectorOptions() CollectorOptions {
	return CollectorOptions{
		Namespace:              "",
		AllNamespaces:          true,
		FetchGuestSubresources: true,
		Concurrency:            10,
		Timeout:                60 * time.Second,
	}
}
