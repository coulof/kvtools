package collector

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var (
	vmSnapshotGVR = schema.GroupVersionResource{
		Group:    "snapshot.kubevirt.io",
		Version:  "v1alpha1",
		Resource: "virtualmachinesnapshots",
	}
	volumeSnapshotGVR = schema.GroupVersionResource{
		Group:    "snapshot.storage.k8s.io",
		Version:  "v1",
		Resource: "volumesnapshots",
	}
)

// FetchVMSnapshots retrieves VirtualMachineSnapshot CRs.
func FetchVMSnapshots(ctx context.Context, dynamicClient dynamic.Interface, namespace string, allNamespaces bool, hasVMSnapshot bool) ([]unstructured.Unstructured, error) {
	if !hasVMSnapshot || dynamicClient == nil {
		return nil, nil
	}

	targetNS := namespace
	if allNamespaces {
		targetNS = metav1.NamespaceAll
	}

	var snapList *unstructured.UnstructuredList
	var err error
	if targetNS == metav1.NamespaceAll {
		snapList, err = dynamicClient.Resource(vmSnapshotGVR).List(ctx, metav1.ListOptions{})
	} else {
		snapList, err = dynamicClient.Resource(vmSnapshotGVR).Namespace(targetNS).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list VirtualMachineSnapshots: %w", err)
	}

	return snapList.Items, nil
}

// FetchVolumeSnapshots retrieves VolumeSnapshot CRs.
func FetchVolumeSnapshots(ctx context.Context, dynamicClient dynamic.Interface, namespace string, allNamespaces bool, hasVolumeSnapshot bool) ([]unstructured.Unstructured, error) {
	if !hasVolumeSnapshot || dynamicClient == nil {
		return nil, nil
	}

	targetNS := namespace
	if allNamespaces {
		targetNS = metav1.NamespaceAll
	}

	var snapList *unstructured.UnstructuredList
	var err error
	if targetNS == metav1.NamespaceAll {
		snapList, err = dynamicClient.Resource(volumeSnapshotGVR).List(ctx, metav1.ListOptions{})
	} else {
		snapList, err = dynamicClient.Resource(volumeSnapshotGVR).Namespace(targetNS).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list VolumeSnapshots: %w", err)
	}

	return snapList.Items, nil
}
