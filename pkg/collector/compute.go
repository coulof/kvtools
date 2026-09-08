package collector

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// FetchNodes retrieves all Kubernetes Nodes in the cluster.
func FetchNodes(ctx context.Context, k8sClient kubernetes.Interface) ([]corev1.Node, error) {
	if k8sClient == nil {
		return nil, fmt.Errorf("kubernetes client is not available")
	}

	nodeList, err := k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	return nodeList.Items, nil
}
