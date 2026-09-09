# Engineering Implementation Plan: `kvtools` (RVTools for KubeVirt)

This plan provides a complete architecture, technical specification, and step-by-step implementation guide to build `kvtools`—a standalone CLI utility written in Go that extracts, audits, and exports comprehensive KubeVirt cluster inventory to multi-tab Excel (`.xlsx`), JSON, and terminal tables, matching the structural style and operational depth of RVTools.

---

## 1. Project Overview & Architecture

### 1.1 Goal
Create a fast, concurrent, and resilient CLI tool (`kvtools`) that connects to any Kubernetes cluster running KubeVirt (Harvester, OpenShift Virtualization, SUSE Virtualization, or upstream KubeVirt), inspects all VM/VMI resources, nodes, storage, networks, snapshots, and guest agent metrics, runs health/hygiene audits (`kvHealth`), and produces a styled, multi-tab Excel workbook identical in spirit to RVTools.

### 1.2 Core Design Principles
1. **Zero Runtime Dependencies**: Compiled to a single static Go binary. Compatible with standard `kubeconfig` and in-cluster service accounts.
2. **Graceful Degradation**: If optional CRDs (e.g. VolumeSnapshots, CDI DataVolumes, Multus NADs) or guest agent subresources are not installed or RBAC-restricted, the tool must log a warning and populate available sheets rather than failing the run.
3. **High Concurrency & Low API Overhead**: Batch-fetch cluster-scoped and namespace-scoped objects in parallel using goroutines and worker pools, avoiding N+1 API roundtrips.
4. **Familiarity**: Maintain sheet naming, color-coded tabs, auto-filters, freeze panes, and column layout analogous to RVTools.

---

## 2. CLI Interface & Flags

Using `spf13/cobra`, the binary runs as a standalone CLI (`kvtools`):

```bash
# Basic usage (all namespaces, exports to auto-named Excel file)
kvtools [flags]

# Common execution patterns
kvtools -A -o excel --output-file ./kvtools-inventory.xlsx
kvtools -n default -o table kvInfo
kvtools --health-only
kvtools -A -o json > inventory.json
```

### Supported Flags

| Flag | Shorthand | Type | Default | Description |
|---|---|---|---|---|
| `--all-namespaces` | `-A` | `bool` | `true` | Query VMs across all namespaces |
| `--namespace` | `-n` | `string` | `""` | Filter by specific namespace |
| `--output` | `-o` | `string` | `excel` | Output format: `excel`, `json`, `csv`, `table` |
| `--output-file` | `-f` | `string` | `kvtools_<cluster>_<timestamp>.xlsx` | Path for export file |
| `--kubeconfig` | | `string` | `~/.kube/config` | Path to kubeconfig |
| `--context` | | `string` | current | Kubernetes context to use |
| `--guest-subresources` | | `bool` | `true` | Fetch guest agent subresources (`/guestosinfo`, `/filesystemlist`) |
| `--health-only` | | `bool` | `false` | Only run and display `kvHealth` audit |
| `--concurrency` | `-c` | `int` | `10` | Worker pool size for parallel subresource queries |
| `--quiet` | `-q` | `bool` | `false` | Suppress progress bars and summary output |

---

## 3. Go Repository Structure

```
kvtools/
├── cmd/
│   ├── root.go                 # Cobra CLI root command, global flags, kubeconfig initialization
│   ├── export.go               # 'export' command (default action)
│   ├── version.go              # Version and build metadata
│   └── health.go               # Quick 'health' terminal inspector
├── pkg/
│   ├── client/                 # K8s & KubeVirt client wrappers
│   │   ├── client.go           # Unified client struct (KubeVirt, K8s Core, Dynamic, Discovery)
│   │   └── discovery.go        # CRD & API capability discovery (CDI, Snapshot, Multus check)
│   ├── collector/              # Raw data retrieval from Kubernetes APIs
│   │   ├── collector.go        # Orchestrator running collectors concurrently
│   │   ├── vm.go               # VMs, VMIs, and guest subresources
│   │   ├── compute.go          # Nodes, CPU/Memory allocations
│   │   ├── storage.go          # PVCs, PVs, DataVolumes, StorageClasses
│   │   ├── network.go          # NetworkAttachmentDefinitions, Interfaces
│   │   ├── snapshot.go         # VirtualMachineSnapshots, VolumeSnapshots
│   │   └── types.go            # Internal normalized data structures
│   ├── engine/                 # Transformation, normalization, and auditing
│   │   ├── transformer.go      # Maps raw API objects to Sheet Record structs
│   │   ├── health.go           # kvHealth rules engine (migration blockers, zombies, QoS)
│   │   └── models.go           # Tab-specific record models (KVInfoRecord, KVCPURecord, etc.)
│   ├── exporter/               # Output formatting and serialization
│   │   ├── excel/
│   │   │   ├── excel.go        # Excelize orchestrator
│   │   │   ├── styles.go       # RVTools-style tab colors, header styles, alignments
│   │   │   └── sheets.go       # Per-tab row builders
│   │   ├── json/
│   │   │   └── json.go         # JSON marshaler
│   │   ├── csv/
│   │   │   └── csv.go          # Multi-file or single-sheet CSV exporter
│   │   └── table/
│   │       └── table.go        # Pretty terminal table generator (tablewriter/lipgloss)
│   └── utils/
│       ├── units.go            # MiB/GiB, core, and time formatting helpers
│       └── progress.go         # Terminal spinner / progress indicators
├── test/
│   ├── fixtures/               # Mock KubeVirt & K8s JSON/YAML manifests
│   └── fake/                   # Fake client helpers for unit & integration testing
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
└── README.md
```

---

## 4. Key Dependencies

```go
require (
    github.com/spf13/cobra v1.8.1
    github.com/xuri/excelize/v2 v2.9.0
    github.com/olekukonko/tablewriter v0.0.5
    k8s.io/api v0.31.x
    k8s.io/apimachinery v0.31.x
    k8s.io/client-go v0.31.x
    kubevirt.io/api v1.3.x
    kubevirt.io/client-go v1.3.x
)
```

---

## 5. Sheet Specifications & Field Mappings

Each sheet corresponds to an RVTools counterpart and maps to specific KubeVirt/K8s fields:

### Sheet 1: `kvInfo` (Tab Color: Jungle Green `#2ECC71`)
*Primary VM Inventory & Identity*
- **Columns**:
  1. `VM Name` (`metadata.name`)
  2. `Namespace` (`metadata.namespace`)
  3. `Power State` (`Running`, `Stopped`, `Starting`, `Paused`, `CrashLoopBackOff`)
  4. `Run Strategy` (`spec.runStrategy` or `spec.running`)
  5. `Node` (`status.nodeName`)
  6. `IP Address` (Primary Pod IP / Guest Agent IP)
  7. `Guest OS` (`status.guestOSInfo.prettyName` / `guestOSInfo.id`)
  8. `Firmware / Boot` (BIOS vs UEFI, SecureBoot enabled)
  9. `TPM Enabled` (bool: `spec.template.spec.domain.devices.tpm`)
  10. `CPUs (Cores/Sockets/Threads)`
  11. `Memory Configured (GiB)`
  12. `Disks Count`
  13. `NICs Count`
  14. `Created Time` (`metadata.creationTimestamp`)
  15. `Uptime` (calculated from `status.phaseTransitionTimestamps`)
  16. `Labels` (formatted string `k=v, ...`)
  17. `Annotations` (Description, metadata)
  18. `UID` (`metadata.uid`)

### Sheet 2: `kvCPU` (Tab Color: Sky Blue `#3498DB`)
*Compute & CPU Topology*
- **Columns**:
  1. `VM Name` / `Namespace`
  2. `Cores` (`spec.template.spec.domain.cpu.cores`)
  3. `Sockets` (`spec.template.spec.domain.cpu.sockets`)
  4. `Threads` (`spec.template.spec.domain.cpu.threads`)
  5. `Total vCPUs` (`cores * sockets * threads`)
  6. `CPU Model` (`host-passthrough`, `host-model`, or specific model)
  7. `Dedicated CPU Placement` (bool: `spec.template.spec.domain.cpu.dedicatedCpuPlacement`)
  8. `NUMA Nodes` (guest NUMA configuration)
  9. `CPU Requests` (`resources.requests.cpu`)
  10. `CPU Limits` (`resources.limits.cpu`)
  11. `Host Node CPU Model` (pulled from assigned `Node.status.nodeInfo.architecture`)

### Sheet 3: `kvMemory` (Tab Color: Royal Blue `#2980B9`)
*RAM Allocation & Overhead*
- **Columns**:
  1. `VM Name` / `Namespace`
  2. `Guest RAM (GiB)` (`domain.memory.guest`)
  3. `Memory Requests (GiB)` (`domain.resources.requests.memory`)
  4. `Memory Limits (GiB)` (`domain.resources.limits.memory`)
  5. `Launcher Overhead (MiB)` (`status.memory.overhead`)
  6. `Hugepages` (None, 2Mi, 1Gi)
  7. `Autoattach Memory Balloon` (bool: `devices.autoattachMemBalloon`)
  8. `Memory Dump enabled`

### Sheet 4: `kvDisk` (Tab Color: Amber `#F39C12`)
*Virtual Disks, PVCs, DataVolumes & CSI Storage*
- **Columns**:
  1. `VM Name` / `Namespace`
  2. `Disk Target Name` (e.g. `disk0`, `rootdisk`)
  3. `Volume Type` (`PersistentVolumeClaim`, `DataVolume`, `ContainerDisk`, `HostDisk`, `CloudInit`)
  4. `Claim Name` (PVC Name)
  5. `StorageClass` (`pvc.spec.storageClassName`)
  6. `Provisioned Size (GiB)` (`pvc.spec.resources.requests.storage`)
  7. `Volume Mode` (`Block` vs `Filesystem`)
  8. `Access Mode` (`ReadWriteOnce`, `ReadWriteMany`, `ReadWriteOncePod`)
  9. `Bus Type` (`virtio`, `scsi`, `sata`)
  10. `Cache Mode` (`none`, `writethrough`, `writeback`)
  11. `IO Mode` (`threads`, `native`)
  12. `Dedicated IO Thread` (bool)
  13. `CSI Driver` (resolved via StorageClass)

### Sheet 5: `kvPartition` (Tab Color: Orange `#E67E22`)
*Guest Filesystem Breakdown (requires QEMU Guest Agent)*
- **Columns**:
  1. `VM Name` / `Namespace`
  2. `Mount Point` (`/`, `C:\`, `/var`, `/home`)
  3. `FS Type` (`ext4`, `xfs`, `ntfs`, `apfs`)
  4. `Disk Name` / `Device`
  5. `Total Capacity (GiB)`
  6. `Used Space (GiB)`
  7. `Free Space (GiB)`
  8. `Free %`

### Sheet 6: `kvNetwork` (Tab Color: Teal `#1ABC9C`)
*Interfaces, Multus Networks & MACs*
- **Columns**:
  1. `VM Name` / `Namespace`
  2. `Interface Name` (`default`, `nic1`, `vlan100`)
  3. `Network Name` (`pod`, or Multus `NetworkAttachmentDefinition` name)
  4. `Binding Type` (`masquerade`, `bridge`, `sriov`, `macvtap`)
  5. `MAC Address` (Configured vs Guest Reported)
  6. `Pod IP`
  7. `Guest Reported IPs` (IPv4 / IPv6 array from guest agent)
  8. `Interface Model` (`virtio`, `e1000`, etc.)
  9. `PCI Address`

### Sheet 7: `kvCD` (Tab Color: Purple `#9B59B6`)
*CD-ROM & ISO Attachments*
- **Columns**:
  1. `VM Name` / `Namespace`
  2. `CD Device Name`
  3. `Source Type` (`ContainerDisk`, `PVC`, `DataVolume`)
  4. `Source Image / Reference` (e.g. `registry.example.com/iso/win2022:latest`)
  5. `Boot Order`
  6. `Connected / Mounted State`

### Sheet 8: `kvSnapshot` (Tab Color: Coral `#E74C3C`)
*VM Snapshots & Backup State*
- **Columns**:
  1. `Snapshot Name`
  2. `Namespace`
  3. `Source VM`
  4. `Ready to Use` (bool)
  5. `Creation Timestamp`
  6. `Age (Days)`
  7. `Volume Snapshot Count`
  8. `Total Restorable Size (GiB)`
  9. `Error / Failure Reason` (if any)

### Sheet 9: `kvGuestAgent` (Tab Color: Steel `#34495E`)
*QEMU Guest Agent Health & Capabilities*
- **Columns**:
  1. `VM Name` / `Namespace`
  2. `Agent Connected` (bool)
  3. `Agent Version`
  4. `Guest Hostname`
  5. `Guest OS Pretty Name`
  6. `Guest Kernel Release`
  7. `Timezone`
  8. `FS Freeze Supported` (bool)
  9. `Logged-in Users Count`

### Sheet 10: `kvNode` (Tab Color: Dark Slate `#2C3E50`)
*Kubernetes Compute Nodes Hosting KubeVirt*
- **Columns**:
  1. `Node Name`
  2. `Status` (`Ready`, `NotReady`, `SchedulingDisabled`)
  3. `Total Physical Cores / Sockets`
  4. `Total RAM (GiB)`
  5. `Allocatable CPU`
  6. `Allocatable RAM (GiB)`
  7. `Allocated VM vCPUs` (sum of running VM cores on node)
  8. `Allocated VM RAM (GiB)` (sum of running VM memory on node)
  9. `vCPU Overcommit Ratio` (`Allocated vCPUs / Allocatable Cores`)
  10. `Active VM Count`
  11. `KVM Hardware Acceleration Enabled` (bool / device plugin check)
  12. `Kubernetes Version` / `OS Image` / `Kernel Version`

### Sheet 11: `kvStoragePool` (Tab Color: Dark Amber `#D35400`)
*Storage Classes & CSI Drivers*
- **Columns**:
  1. `StorageClass Name`
  2. `Provisioner / CSI Driver`
  3. `Reclaim Policy` (`Delete` / `Retain`)
  4. `VolumeBindingMode` (`WaitForFirstConsumer` / `Immediate`)
  5. `AllowVolumeExpansion` (bool)
  6. `Is Default Class` (bool)
  7. `Total Bound PVC Count`
  8. `Total Allocated Capacity (GiB)`

### Sheet 12: `kvHardware` (Tab Color: Brown `#795548`)
*PCI Passthrough, GPUs & Host Devices*
- **Columns**:
  1. `VM Name` / `Namespace`
  2. `Device Type` (`GPU`, `vGPU`, `SRIOV NIC`, `USB`, `HostDevice`)
  3. `Device Name`
  4. `Resource Name` (e.g. `nvidia.com/GA102GL_A40`)
  5. `Node Selector / Assigned Node`

### Sheet 13: `kvHealth` (Tab Color: Crimson `#C0392B`)
*Automated Best-Practice, Hygiene & Migration Audit*
- **Columns**:
  1. `Category` (`Migration Blocker`, `Storage Zombie`, `Guest Visibility`, `Snapshot Sprawl`, `QoS Risk`)
  2. `Severity` (`CRITICAL`, `WARNING`, `INFO`)
  3. `Resource Kind` (`VirtualMachine`, `PersistentVolumeClaim`, `Snapshot`, `Node`)
  4. `Resource Name`
  5. `Namespace`
  6. `Issue Summary`
  7. `Remediation Recommendation`

---

## 6. `kvHealth` Built-in Audit Rules

Implement the following deterministic checks in `pkg/engine/health.go`:

| Rule ID | Category | Severity | Condition | Remediation |
|---|---|---|---|---|
| `HLTH-001` | Migration Blocker | `CRITICAL` | VM has RWO PVC and `liveMigration` is requested | Change storage to RWX (shared) or configure CDI block volume migration |
| `HLTH-002` | Migration Blocker | `WARNING` | VM uses `host-passthrough` CPU model | Switch to `host-model` or ensure all cluster nodes have identical CPU microarchitecture |
| `HLTH-003` | Migration Blocker | `CRITICAL` | VM mounts a local `HostDisk` | Replace HostDisk with PVC or ContainerDisk before attempting live migration |
| `HLTH-004` | Migration Blocker | `WARNING` | VM has direct SR-IOV NIC without bond failover | Configure failover secondary network interface |
| `HLTH-005` | Storage Zombie | `WARNING` | PVC name matches KubeVirt pattern (`*-disk-*`, `*-dv-*`) but owning VM/VMI UID no longer exists | Delete orphaned PVC to reclaim storage capacity |
| `HLTH-006` | Storage Zombie | `WARNING` | DataVolume in `Failed` or stuck `ImportInProgress` state for > 2 hours | Inspect CDI importer pod logs and clean up failed import DataVolume |
| `HLTH-007` | Guest Visibility | `WARNING` | Running VM has `AgentConnected: false` | Install and enable `qemu-guest-agent` inside guest OS for IP reporting and quiescing |
| `HLTH-008` | Snapshot Sprawl | `WARNING` | `VirtualMachineSnapshot` age > 14 days | Merge/delete old VM snapshots to avoid storage degradation and snapshot chain limits |
| `HLTH-009` | QoS / Sizing | `WARNING` | VM CPU limits > 4× CPU requests | Set consistent CPU limits or remove limits to prevent aggressive CFS scheduler throttling |
| `HLTH-010` | QoS / Sizing | `CRITICAL` | VM memory limits equal to memory request with 0 overhead margin | Add launcher overhead allowance to avoid kernel OOM killer terminating `virt-launcher` |
| `HLTH-011` | Node Imbalance | `WARNING` | Node vCPU overcommit ratio exceeds 8:1 | Rebalance VMs across cluster nodes or add compute capacity |

---

## 7. Parallel API Data Collection Workflow

To ensure high performance across large clusters (1000+ VMs):

```
                       ┌───────────────────────┐
                       │  Start kvtools CLI    │
                       └──────────┬────────────┘
                                  │
                   ┌──────────────┴──────────────┐
                   │ Discovery: Check installed  │
                   │ CRDs (CDI, Snapshot, CNI)   │
                   └──────────────┬──────────────┘
                                  │
         ┌────────────────────────┼────────────────────────┐
         ▼                        ▼                        ▼
┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐
│ List all VMs &   │    │ List Nodes &     │    │ List Storage &   │
│ VMIs (cluster)   │    │ Cluster Specs    │    │ StorageClasses   │
└────────┬─────────┘    └────────┬─────────┘    └────────┬─────────┘
         │                       │                       │
         ▼                       │                       │
┌──────────────────┐             │                       │
│ Worker Pool:     │             │                       │
│ Fetch Subresources             │                       │
│ (guestosinfo &   │             │                       │
│  filesystemlist) │             │                       │
└────────┬─────────┘             │                       │
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                                 ▼
                    ┌─────────────────────────┐
                    │ Engine: Normalization & │
                    │ kvHealth Rule Checks    │
                    └────────────┬────────────┘
                                 │
                                 ▼
                    ┌─────────────────────────┐
                    │ Exporter: Generate      │
                    │ Styled Excel (.xlsx)    │
                    └─────────────────────────┘
```

---

## 8. Excel Formatting Standards (`pkg/exporter/excel/styles.go`)

Using `github.com/xuri/excelize/v2`:
1. **Header Rows**:
   - Background Color: Matching tab theme color (Dark tint).
   - Font: White, Bold, 11pt, `Segoe UI` / `Arial`.
   - Row height: 26pt.
2. **Data Rows**:
   - Alternating zebra striping (`#FFFFFF` / `#F8F9FA`).
   - Font: 10pt, regular.
   - Text aligned Left; Numeric/Count/Bytes aligned Right; Booleans/States Centered.
3. **Features**:
   - `AutoFilter` enabled across all columns on every sheet.
   - `FreezePanes` on row 1 (keep headers visible while scrolling).
   - Auto-fit column widths based on maximum content length + padding.
   - Tab colors set via `SetSheetPrOptions` with custom hex RGB codes.

---

## 9. Implementation Milestones

### Milestone 1: Core Scaffolding & Basic Inventory (Days 1–2)
- [ ] Initialize Go module `github.com/<org>/kvtools`.
- [ ] Setup Cobra CLI commands (`export`, `version`, `health`).
- [ ] Implement KubeConfig loading & client factory in `pkg/client/client.go`.
- [ ] Build collectors for `VM`, `VMI`, and `Node`.
- [ ] Implement transformers for `kvInfo`, `kvCPU`, `kvMemory`, `kvNode`.
- [ ] Set up `excelize` engine to write basic multi-tab workbook.

### Milestone 2: Storage, Network & Snapshot Sheets (Days 3–4)
- [ ] Implement discovery checks for CDI and Snapshot CRDs in `pkg/client/discovery.go`.
- [ ] Build collectors for PVCs, PVs, StorageClasses, DataVolumes.
- [ ] Build collector for NetworkAttachmentDefinitions and VMI network interfaces.
- [ ] Build collector for `VirtualMachineSnapshot` and `VolumeSnapshot`.
- [ ] Implement transformers for `kvDisk`, `kvNetwork`, `kvCD`, `kvSnapshot`, `kvStoragePool`, `kvHardware`.

### Milestone 3: Guest Subresources & Concurrency (Days 5–6)
- [ ] Implement worker pool in `pkg/collector/vm.go` for VMI subresources:
  - `GET /apis/subresources.kubevirt.io/v1/namespaces/{ns}/virtualmachineinstances/{name}/guestosinfo`
  - `GET /apis/subresources.kubevirt.io/v1/namespaces/{ns}/virtualmachineinstances/{name}/filesystemlist`
- [ ] Implement transformers for `kvPartition` and `kvGuestAgent`.
- [ ] Handle graceful fallback when guest agent is not running or subresource call returns 404/500.

### Milestone 4: `kvHealth` Engine & CLI Exporters (Days 7–8)
- [ ] Implement all 11 audit rules in `pkg/engine/health.go`.
- [ ] Write `kvHealth` tab generator with severity color tags (Red for Critical, Yellow for Warning).
- [ ] Implement CLI table output (`tablewriter`) for terminal viewing.
- [ ] Implement JSON output (`--output json`).

### Milestone 5: Polishing, Testing & Packaging (Days 9–10)
- [ ] Add unit tests with fake client and mocked manifests in `test/fixtures/`.
- [ ] Test on live Harvester / OpenShift Virtualization / KubeVirt cluster.
- [ ] Add auto-fit column width calculation and Excel styling refinements.
- [ ] Write documentation and `Makefile` build targets.

---

## 10. Verification & Test Plan

1. **Unit Testing (`go test ./...`)**:
   - Use `client-go/kubernetes/fake` and `kubevirt.io/client-go/kubecli` mock client.
   - Load fixture YAMLs from `test/fixtures/` and verify that each transformer generates the exact expected column count and data values.
   - Test `kvHealth` rule engine with intentionally misconfigured VM fixtures (missing guest agent, RWO live migration blocker, zombie PVC).
2. **Missing CRD Resilience Test**:
   - Run collector against a cluster without CDI or Snapshot CRDs installed. Verify no panic, clean log warning, and empty/skipped optional tabs.
3. **Live Cluster Verification**:
   - Run `kvtools -A -o excel -f test-export.xlsx` against a development KubeVirt/Harvester cluster.
   - Open `test-export.xlsx` in Excel / LibreOffice / Google Sheets and verify:
     - All 13 tabs exist with correct colors.
     - Auto-filters work.
     - Row freeze works.
     - `kvHealth` accurately flags test VMs.
4. **Integration with `harvester-sizer`**:
   - Verify that the resulting `kvtools.json` or `.xlsx` can be read back into sizing tools for migration re-evaluation and cluster re-balancing.

---

### Handoff Note
This document contains the complete technical blueprint. The implementing agent can start immediately at **Milestone 1** by setting up the Go module, CLI structure with Cobra, and client-go connection logic.
