package collector

import (
	"context"
	"fmt"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	virtv1 "kubevirt.io/api/core/v1"
	"kubevirt.io/client-go/kubecli"
)

// FetchVMs retrieves VirtualMachine resources from the cluster.
func FetchVMs(ctx context.Context, kvClient kubecli.KubevirtClient, namespace string, allNamespaces bool) ([]virtv1.VirtualMachine, error) {
	if kvClient == nil {
		return nil, fmt.Errorf("kubevirt client is not available")
	}

	targetNS := namespace
	if allNamespaces {
		targetNS = metav1.NamespaceAll
	}

	vmList, err := kvClient.VirtualMachine(targetNS).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list virtual machines: %w", err)
	}

	return vmList.Items, nil
}

// FetchVMIs retrieves VirtualMachineInstance resources from the cluster.
func FetchVMIs(ctx context.Context, kvClient kubecli.KubevirtClient, namespace string, allNamespaces bool) ([]virtv1.VirtualMachineInstance, error) {
	if kvClient == nil {
		return nil, fmt.Errorf("kubevirt client is not available")
	}

	targetNS := namespace
	if allNamespaces {
		targetNS = metav1.NamespaceAll
	}

	vmiList, err := kvClient.VirtualMachineInstance(targetNS).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list virtual machine instances: %w", err)
	}

	return vmiList.Items, nil
}

type subresourceResult struct {
	key     string
	osInfo  *virtv1.VirtualMachineInstanceGuestAgentInfo
	fsList  *virtv1.VirtualMachineInstanceFileSystemList
	warning string
}

// FetchGuestSubresources concurrently queries guestosinfo and filesystemlist for running VMIs.
func FetchGuestSubresources(
	ctx context.Context,
	kvClient kubecli.KubevirtClient,
	vmis []virtv1.VirtualMachineInstance,
	concurrency int,
) (map[string]*virtv1.VirtualMachineInstanceGuestAgentInfo, map[string]*virtv1.VirtualMachineInstanceFileSystemList, []string) {
	osInfoMap := make(map[string]*virtv1.VirtualMachineInstanceGuestAgentInfo)
	fsListMap := make(map[string]*virtv1.VirtualMachineInstanceFileSystemList)
	var warnings []string

	if kvClient == nil || len(vmis) == 0 {
		return osInfoMap, fsListMap, warnings
	}

	if concurrency <= 0 {
		concurrency = 10
	}

	// Filter VMIs that are running
	var runningVMIs []virtv1.VirtualMachineInstance
	for _, vmi := range vmis {
		if vmi.Status.Phase == virtv1.Running {
			runningVMIs = append(runningVMIs, vmi)
		}
	}

	if len(runningVMIs) == 0 {
		return osInfoMap, fsListMap, warnings
	}

	jobs := make(chan virtv1.VirtualMachineInstance, len(runningVMIs))
	results := make(chan subresourceResult, len(runningVMIs))

	var wg sync.WaitGroup
	workerCount := concurrency
	if len(runningVMIs) < workerCount {
		workerCount = len(runningVMIs)
	}

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for vmi := range jobs {
				key := fmt.Sprintf("%s/%s", vmi.Namespace, vmi.Name)
				res := subresourceResult{key: key}

				// Create a per-vmi timeout to prevent slow agent queries from blocking the whole run
				subCtx, cancel := context.WithTimeout(ctx, 10*time.Second)

				// Fetch guest OS info
				osInfo, err := kvClient.VirtualMachineInstance(vmi.Namespace).GuestOsInfo(subCtx, vmi.Name)
				if err == nil {
					res.osInfo = &osInfo
				}

				// Fetch filesystem list
				fsList, err := kvClient.VirtualMachineInstance(vmi.Namespace).FilesystemList(subCtx, vmi.Name)
				if err == nil {
					res.fsList = &fsList
				}

				cancel()
				results <- res
			}
		}()
	}

	for _, vmi := range runningVMIs {
		jobs <- vmi
	}
	close(jobs)

	wg.Wait()
	close(results)

	for res := range results {
		if res.osInfo != nil {
			osInfoMap[res.key] = res.osInfo
		}
		if res.fsList != nil {
			fsListMap[res.key] = res.fsList
		}
		if res.warning != "" {
			warnings = append(warnings, res.warning)
		}
	}

	return osInfoMap, fsListMap, warnings
}
