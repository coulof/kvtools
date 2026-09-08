package engine

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	virtv1 "kubevirt.io/api/core/v1"

	"github.com/coulof/kvtools/pkg/collector"
)

func TestHealthAuditRules(t *testing.T) {
	liveMigrate := virtv1.EvictionStrategyLiveMigrate

	// VM 1: Has RWO PVC + LiveMigrate (HLTH-001) + host-passthrough CPU (HLTH-002) + HostDisk (HLTH-003) + single SRIOV (HLTH-004)
	vm1 := virtv1.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-vm-1",
			Namespace: "default",
			UID:       types.UID("uid-vm-1"),
		},
		Spec: virtv1.VirtualMachineSpec{
			Template: &virtv1.VirtualMachineInstanceTemplateSpec{
				Spec: virtv1.VirtualMachineInstanceSpec{
					EvictionStrategy: &liveMigrate,
					Domain: virtv1.DomainSpec{
						CPU: &virtv1.CPU{
							Model: "host-passthrough",
						},
						Devices: virtv1.Devices{
							Interfaces: []virtv1.Interface{
								{
									Name: "sriov-net",
									InterfaceBindingMethod: virtv1.InterfaceBindingMethod{
										SRIOV: &virtv1.InterfaceSRIOV{},
									},
								},
							},
						},
					},
					Volumes: []virtv1.Volume{
						{
							Name: "rwo-vol",
							VolumeSource: virtv1.VolumeSource{
								PersistentVolumeClaim: &virtv1.PersistentVolumeClaimVolumeSource{
									PersistentVolumeClaimVolumeSource: corev1.PersistentVolumeClaimVolumeSource{
										ClaimName: "rwo-pvc",
									},
								},
							},
						},
						{
							Name: "host-vol",
							VolumeSource: virtv1.VolumeSource{
								HostDisk: &virtv1.HostDisk{
									Path: "/var/lib/local.img",
								},
							},
						},
					},
				},
			},
		},
	}

	// VM 2: Running with AgentConnected: false (HLTH-007) + CPU limit > 4x req (HLTH-009) + Mem limit == req with no overhead (HLTH-010)
	cpuReq := resource.MustParse("500m")
	cpuLim := resource.MustParse("3000m")
	memReq := resource.MustParse("4Gi")
	memLim := resource.MustParse("4Gi")

	vm2 := virtv1.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-vm-2",
			Namespace: "default",
			UID:       types.UID("uid-vm-2"),
		},
		Spec: virtv1.VirtualMachineSpec{
			Template: &virtv1.VirtualMachineInstanceTemplateSpec{
				Spec: virtv1.VirtualMachineInstanceSpec{
					Domain: virtv1.DomainSpec{
						Resources: virtv1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    cpuReq,
								corev1.ResourceMemory: memReq,
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    cpuLim,
								corev1.ResourceMemory: memLim,
							},
						},
					},
				},
			},
		},
	}

	vmi2 := virtv1.VirtualMachineInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-vm-2",
			Namespace: "default",
			UID:       types.UID("uid-vm-2"),
		},
		Status: virtv1.VirtualMachineInstanceStatus{
			Phase: virtv1.Running,
			Conditions: []virtv1.VirtualMachineInstanceCondition{
				{
					Type:   virtv1.VirtualMachineInstanceAgentConnected,
					Status: corev1.ConditionFalse,
				},
			},
		},
	}

	// PVCs
	pvcRWO := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rwo-pvc",
			Namespace: "default",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
		},
	}

	// Zombie PVC (HLTH-005)
	pvcZombie := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-vm-old-disk-0",
			Namespace: "default",
			Labels: map[string]string{
				"kubevirt.io/created-by": "deleted-vm",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					UID:  types.UID("old-deleted-uid"),
					Name: "deleted-vm",
				},
			},
		},
	}

	// Stuck/Failed DataVolume (HLTH-006)
	dvFailed := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cdi.kubevirt.io/v1beta1",
			"kind":       "DataVolume",
			"metadata": map[string]interface{}{
				"name":      "failed-dv",
				"namespace": "default",
			},
			"status": map[string]interface{}{
				"phase": "Failed",
			},
		},
	}

	// Old snapshot (HLTH-008)
	twentyDaysAgo := metav1.NewTime(time.Now().Add(-20 * 24 * time.Hour))
	snapOld := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "snapshot.kubevirt.io/v1alpha1",
			"kind":       "VirtualMachineSnapshot",
			"metadata": map[string]interface{}{
				"name":              "aged-snap",
				"namespace":         "default",
				"creationTimestamp": twentyDaysAgo.Format(time.RFC3339),
			},
			"spec": map[string]interface{}{
				"source": map[string]interface{}{
					"name": "test-vm-1",
				},
			},
			"status": map[string]interface{}{
				"readyToUse": true,
			},
		},
	}
	snapOld.SetCreationTimestamp(twentyDaysAgo)

	// Node overcommit > 8:1 (HLTH-011)
	nodeOvercommit := KVNodeRecord{
		NodeName:            "node-1",
		VCPUOvercommitRatio: 9.5,
	}

	rawData := &collector.RawData{
		VMs:         []virtv1.VirtualMachine{vm1, vm2},
		VMIs:        []virtv1.VirtualMachineInstance{vmi2},
		PVCs:        []corev1.PersistentVolumeClaim{pvcRWO, pvcZombie},
		DataVolumes: []unstructured.Unstructured{dvFailed},
		VMSnapshots: []unstructured.Unstructured{snapOld},
	}

	healthFindings := RunHealthAudit(rawData, []KVNodeRecord{nodeOvercommit})

	rulesFound := make(map[string]bool)
	for _, f := range healthFindings {
		rulesFound[f.RuleID] = true
	}

	expectedRules := []string{
		"HLTH-001", "HLTH-002", "HLTH-003", "HLTH-004",
		"HLTH-005", "HLTH-006", "HLTH-007", "HLTH-008",
		"HLTH-009", "HLTH-010", "HLTH-011",
	}

	for _, r := range expectedRules {
		if !rulesFound[r] {
			t.Errorf("expected health audit rule %s to trigger, but it was not found", r)
		}
	}
}
