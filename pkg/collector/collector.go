package collector

import (
	"context"
	"fmt"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	virtv1 "kubevirt.io/api/core/v1"

	"github.com/coulof/kvtools/pkg/client"
	"github.com/coulof/kvtools/pkg/utils"
)

// Collector coordinates concurrent fetching of all cluster resources.
type Collector struct {
	client *client.Client
}

// NewCollector creates a new Collector instance with the given Client.
func NewCollector(c *client.Client) *Collector {
	return &Collector{
		client: c,
	}
}

// Collect fetches all relevant cluster data concurrently.
func (c *Collector) Collect(ctx context.Context, opts CollectorOptions, prog *utils.Progress) (*RawData, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client is nil")
	}

	raw := &RawData{
		ClusterName:    c.client.ClusterName,
		CollectedAt:    time.Now().UTC(),
		GuestOSInfo:    make(map[string]*virtv1.VirtualMachineInstanceGuestAgentInfo),
		FileSystemList: make(map[string]*virtv1.VirtualMachineInstanceFileSystemList),
	}

	var mu sync.Mutex
	var warnings []string
	addWarning := func(w string) {
		mu.Lock()
		defer mu.Unlock()
		warnings = append(warnings, w)
	}

	if prog != nil {
		prog.Start("Connecting to cluster and fetching inventory")
	}

	var wg sync.WaitGroup

	// Task 1: Fetch VMs and VMIs
	var vms []virtv1.VirtualMachine
	var vmis []virtv1.VirtualMachineInstance
	var vmErr, vmiErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		if c.client.KV != nil {
			var innerWg sync.WaitGroup
			innerWg.Add(2)
			go func() {
				defer innerWg.Done()
				vms, vmErr = FetchVMs(ctx, c.client.KV, opts.Namespace, opts.AllNamespaces)
			}()
			go func() {
				defer innerWg.Done()
				vmis, vmiErr = FetchVMIs(ctx, c.client.KV, opts.Namespace, opts.AllNamespaces)
			}()
			innerWg.Wait()
		} else {
			addWarning("KubeVirt client unavailable; skipping VM/VMI collection")
		}
	}()

	// Task 2: Fetch Nodes
	var nodes []corev1.Node
	var nodeErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		if c.client.K8s != nil {
			nodes, nodeErr = FetchNodes(ctx, c.client.K8s)
		}
	}()

	// Task 3: Fetch Storage (PVCs, PVs, StorageClasses, DataVolumes)
	var pvcs []corev1.PersistentVolumeClaim
	var pvs []corev1.PersistentVolume
	var scs []storagev1.StorageClass
	var dvs []unstructured.Unstructured
	var pvcErr, pvErr, scErr, dvErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		if c.client.K8s != nil {
			var sWg sync.WaitGroup
			sWg.Add(3)
			go func() {
				defer sWg.Done()
				pvcs, pvcErr = FetchPVCs(ctx, c.client.K8s, opts.Namespace, opts.AllNamespaces)
			}()
			go func() {
				defer sWg.Done()
				pvs, pvErr = FetchPVs(ctx, c.client.K8s)
			}()
			go func() {
				defer sWg.Done()
				scs, scErr = FetchStorageClasses(ctx, c.client.K8s)
			}()
			sWg.Wait()
		}
		if c.client.Dynamic != nil && c.client.Capabilities != nil && c.client.Capabilities.HasCDI {
			dvs, dvErr = FetchDataVolumes(ctx, c.client.Dynamic, opts.Namespace, opts.AllNamespaces, true)
		}
	}()

	// Task 4: Fetch NetworkAttachmentDefinitions
	var nads []unstructured.Unstructured
	var nadErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		if c.client.Dynamic != nil && c.client.Capabilities != nil && c.client.Capabilities.HasMultus {
			nads, nadErr = FetchNetworkAttachmentDefinitions(ctx, c.client.Dynamic, opts.Namespace, opts.AllNamespaces, true)
		}
	}()

	// Task 5: Fetch Snapshots
	var vmSnaps []unstructured.Unstructured
	var volSnaps []unstructured.Unstructured
	var vmSnapErr, volSnapErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		if c.client.Dynamic != nil && c.client.Capabilities != nil {
			var snapWg sync.WaitGroup
			if c.client.Capabilities.HasVMSnapshot {
				snapWg.Add(1)
				go func() {
					defer snapWg.Done()
					vmSnaps, vmSnapErr = FetchVMSnapshots(ctx, c.client.Dynamic, opts.Namespace, opts.AllNamespaces, true)
				}()
			}
			if c.client.Capabilities.HasVolumeSnapshot {
				snapWg.Add(1)
				go func() {
					defer snapWg.Done()
					volSnaps, volSnapErr = FetchVolumeSnapshots(ctx, c.client.Dynamic, opts.Namespace, opts.AllNamespaces, true)
				}()
			}
			snapWg.Wait()
		}
	}()

	wg.Wait()

	// Process errors / warnings
	if vmErr != nil {
		addWarning(fmt.Sprintf("Failed to list VirtualMachines: %v", vmErr))
	}
	if vmiErr != nil {
		addWarning(fmt.Sprintf("Failed to list VirtualMachineInstances: %v", vmiErr))
	}
	if nodeErr != nil {
		addWarning(fmt.Sprintf("Failed to list Nodes: %v", nodeErr))
	}
	if pvcErr != nil {
		addWarning(fmt.Sprintf("Failed to list PVCs: %v", pvcErr))
	}
	if pvErr != nil {
		addWarning(fmt.Sprintf("Failed to list PVs: %v", pvErr))
	}
	if scErr != nil {
		addWarning(fmt.Sprintf("Failed to list StorageClasses: %v", scErr))
	}
	if dvErr != nil {
		addWarning(fmt.Sprintf("Failed to list DataVolumes: %v", dvErr))
	}
	if nadErr != nil {
		addWarning(fmt.Sprintf("Failed to list NetworkAttachmentDefinitions: %v", nadErr))
	}
	if vmSnapErr != nil {
		addWarning(fmt.Sprintf("Failed to list VirtualMachineSnapshots: %v", vmSnapErr))
	}
	if volSnapErr != nil {
		addWarning(fmt.Sprintf("Failed to list VolumeSnapshots: %v", volSnapErr))
	}

	raw.VMs = vms
	raw.VMIs = vmis
	raw.Nodes = nodes
	raw.PVCs = pvcs
	raw.PVs = pvs
	raw.StorageClasses = scs
	raw.DataVolumes = dvs
	raw.NetworkAttachmentDefinitions = nads
	raw.VMSnapshots = vmSnaps
	raw.VolumeSnapshots = volSnaps

	// Task 6: Guest Agent Subresources (if requested and VMIs are found)
	if opts.FetchGuestSubresources && len(vmis) > 0 && c.client.KV != nil {
		if prog != nil {
			prog.Update(fmt.Sprintf("Fetching guest agent subresources for %d VMIs", len(vmis)))
		}
		osMap, fsMap, subWarnings := FetchGuestSubresources(ctx, c.client.KV, vmis, opts.Concurrency)
		raw.GuestOSInfo = osMap
		raw.FileSystemList = fsMap
		for _, sw := range subWarnings {
			addWarning(sw)
		}
	}

	raw.Warnings = warnings

	if prog != nil {
		prog.Success(fmt.Sprintf("Collected %d VMs, %d VMIs, %d Nodes, %d PVCs", len(raw.VMs), len(raw.VMIs), len(raw.Nodes), len(raw.PVCs)))
	}

	return raw, nil
}
