package client

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"kubevirt.io/client-go/kubecli"
)

// Client holds all initialized Kubernetes and KubeVirt client handles.
type Client struct {
	K8s          kubernetes.Interface
	KV           kubecli.KubevirtClient
	Dynamic      dynamic.Interface
	Discovery    discovery.DiscoveryInterface
	RestConfig   *rest.Config
	ClusterName  string
	Capabilities *ClusterCapabilities
}

// ConfigOptions encapsulates parameters for creating a new Client.
type ConfigOptions struct {
	KubeconfigPath string
	Context        string
	QPS            float32
	Burst          int
}

// DefaultConfigOptions returns the standard config options.
func DefaultConfigOptions() ConfigOptions {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}
	return ConfigOptions{
		KubeconfigPath: kubeconfig,
		Context:        "",
		QPS:            50.0,
		Burst:          100,
	}
}

// NewClient initializes a new Client with K8s, KubeVirt, Dynamic, and Discovery clients.
func NewClient(opts ConfigOptions) (*Client, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if opts.KubeconfigPath != "" {
		loadingRules.ExplicitPath = opts.KubeconfigPath
	}

	configOverrides := &clientcmd.ConfigOverrides{}
	if opts.Context != "" {
		configOverrides.CurrentContext = opts.Context
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	rawConfig, err := clientConfig.RawConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load raw kubeconfig: %w", err)
	}

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build rest config: %w", err)
	}

	// Set high QPS and Burst for efficient bulk queries
	if opts.QPS > 0 {
		restConfig.QPS = opts.QPS
	} else {
		restConfig.QPS = 50.0
	}
	if opts.Burst > 0 {
		restConfig.Burst = opts.Burst
	} else {
		restConfig.Burst = 100
	}

	// Determine cluster name
	clusterName := "kubernetes"
	currentContextName := rawConfig.CurrentContext
	if opts.Context != "" {
		currentContextName = opts.Context
	}
	if ctx, exists := rawConfig.Contexts[currentContextName]; exists && ctx != nil {
		if ctx.Cluster != "" {
			clusterName = ctx.Cluster
		} else {
			clusterName = currentContextName
		}
	}

	// Initialize standard K8s clientset
	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	// Initialize dynamic client
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// Initialize discovery client
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery client: %w", err)
	}

	// Initialize KubeVirt client
	kvClient, err := kubecli.GetKubevirtClientFromRESTConfig(restConfig)
	if err != nil {
		// Even if KubeVirt client fails init (e.g. non-kubevirt cluster), we can still fallback gracefully
		kvClient = nil
	}

	// Discover cluster capabilities
	caps, err := DiscoverCapabilities(discoveryClient)
	if err != nil {
		// Log capability discovery warning but continue
		caps = &ClusterCapabilities{AvailableGVRs: make(map[string]bool)}
	}

	return &Client{
		K8s:          k8sClient,
		KV:           kvClient,
		Dynamic:      dynamicClient,
		Discovery:    discoveryClient,
		RestConfig:   restConfig,
		ClusterName:  clusterName,
		Capabilities: caps,
	}, nil
}
