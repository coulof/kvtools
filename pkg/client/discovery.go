package client

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
)

// ClusterCapabilities holds flags indicating available CRDs and API groups.
type ClusterCapabilities struct {
	HasKubeVirt          bool
	HasCDI               bool
	HasVMSnapshot        bool
	HasVolumeSnapshot    bool
	HasMultus            bool
	HasMetrics           bool
	AvailableGVRs        map[string]bool
}

// DiscoverCapabilities queries the cluster discovery API to check available CRDs and APIs.
func DiscoverCapabilities(discoveryClient discovery.DiscoveryInterface) (*ClusterCapabilities, error) {
	caps := &ClusterCapabilities{
		AvailableGVRs: make(map[string]bool),
	}

	if discoveryClient == nil {
		return caps, fmt.Errorf("discovery client is nil")
	}

	// Fetch all API resource lists
	_, apiResourceLists, err := discoveryClient.ServerGroupsAndResources()
	if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
		// Even if some groups failed (e.g. broken webhooks), we can still inspect the rest
		return caps, fmt.Errorf("failed to discover server resources: %w", err)
	}

	for _, list := range apiResourceLists {
		if list == nil {
			continue
		}
		gv, parseErr := schema.ParseGroupVersion(list.GroupVersion)
		if parseErr != nil {
			continue
		}

		for _, resource := range list.APIResources {
			// Skip subresources (containing '/')
			if strings.Contains(resource.Name, "/") {
				continue
			}

			key := fmt.Sprintf("%s/%s/%s", gv.Group, gv.Version, resource.Name)
			caps.AvailableGVRs[key] = true

			// Check specific capabilities
			if gv.Group == "kubevirt.io" && (resource.Name == "virtualmachines" || resource.Name == "virtualmachineinstances") {
				caps.HasKubeVirt = true
			}
			if gv.Group == "cdi.kubevirt.io" && resource.Name == "datavolumes" {
				caps.HasCDI = true
			}
			if gv.Group == "snapshot.kubevirt.io" && resource.Name == "virtualmachinesnapshots" {
				caps.HasVMSnapshot = true
			}
			if gv.Group == "snapshot.storage.k8s.io" && resource.Name == "volumesnapshots" {
				caps.HasVolumeSnapshot = true
			}
			if gv.Group == "k8s.cni.cncf.io" && resource.Name == "network-attachment-definitions" {
				caps.HasMultus = true
			}
			if gv.Group == "metrics.k8s.io" && resource.Name == "nodes" {
				caps.HasMetrics = true
			}
		}
	}

	return caps, nil
}

// HasResource returns true if the specific group/version/resource is available in the cluster.
func (c *ClusterCapabilities) HasResource(group, version, resource string) bool {
	if c == nil || c.AvailableGVRs == nil {
		return false
	}
	key := fmt.Sprintf("%s/%s/%s", group, version, resource)
	return c.AvailableGVRs[key]
}
