package collector

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

var (
	dataVolumeGVR = schema.GroupVersionResource{
		Group:    "cdi.kubevirt.io",
		Version:  "v1beta1",
		Resource: "datavolumes",
	}
)

// FetchPVCs retrieves PersistentVolumeClaims.
func FetchPVCs(ctx context.Context, k8sClient kubernetes.Interface, namespace string, allNamespaces bool) ([]corev1.PersistentVolumeClaim, error) {
	if k8sClient == nil {
		return nil, fmt.Errorf("kubernetes client is not available")
	}

	targetNS := namespace
	if allNamespaces {
		targetNS = metav1.NamespaceAll
	}

	pvcList, err := k8sClient.CoreV1().PersistentVolumeClaims(targetNS).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list PVCs: %w", err)
	}

	return pvcList.Items, nil
}

// FetchPVs retrieves PersistentVolumes.
func FetchPVs(ctx context.Context, k8sClient kubernetes.Interface) ([]corev1.PersistentVolume, error) {
	if k8sClient == nil {
		return nil, fmt.Errorf("kubernetes client is not available")
	}

	pvList, err := k8sClient.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list PVs: %w", err)
	}

	return pvList.Items, nil
}

// FetchStorageClasses retrieves StorageClasses.
func FetchStorageClasses(ctx context.Context, k8sClient kubernetes.Interface) ([]storagev1.StorageClass, error) {
	if k8sClient == nil {
		return nil, fmt.Errorf("kubernetes client is not available")
	}

	scList, err := k8sClient.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list storage classes: %w", err)
	}

	return scList.Items, nil
}

// FetchDataVolumes retrieves CDI DataVolumes if the CRD exists.
func FetchDataVolumes(ctx context.Context, dynamicClient dynamic.Interface, namespace string, allNamespaces bool, hasCDI bool) ([]unstructured.Unstructured, error) {
	if !hasCDI || dynamicClient == nil {
		return nil, nil
	}

	targetNS := namespace
	if allNamespaces {
		targetNS = metav1.NamespaceAll
	}

	var dvList *unstructured.UnstructuredList
	var err error
	if targetNS == metav1.NamespaceAll {
		dvList, err = dynamicClient.Resource(dataVolumeGVR).List(ctx, metav1.ListOptions{})
	} else {
		dvList, err = dynamicClient.Resource(dataVolumeGVR).Namespace(targetNS).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list DataVolumes: %w", err)
	}

	return dvList.Items, nil
}
