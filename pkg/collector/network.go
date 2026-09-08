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
	netAttachDefGVR = schema.GroupVersionResource{
		Group:    "k8s.cni.cncf.io",
		Version:  "v1",
		Resource: "network-attachment-definitions",
	}
)

// FetchNetworkAttachmentDefinitions retrieves Multus NetworkAttachmentDefinitions if the CRD exists.
func FetchNetworkAttachmentDefinitions(ctx context.Context, dynamicClient dynamic.Interface, namespace string, allNamespaces bool, hasMultus bool) ([]unstructured.Unstructured, error) {
	if !hasMultus || dynamicClient == nil {
		return nil, nil
	}

	targetNS := namespace
	if allNamespaces {
		targetNS = metav1.NamespaceAll
	}

	var nadList *unstructured.UnstructuredList
	var err error
	if targetNS == metav1.NamespaceAll {
		nadList, err = dynamicClient.Resource(netAttachDefGVR).List(ctx, metav1.ListOptions{})
	} else {
		nadList, err = dynamicClient.Resource(netAttachDefGVR).Namespace(targetNS).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list NetworkAttachmentDefinitions: %w", err)
	}

	return nadList.Items, nil
}
