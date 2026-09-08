# kvtools (RVTools for KubeVirt)

[![Go Version](https://img.shields.io/github/go-mod/go-version/coulof/kvtools)](https://golang.org)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

`kvtools` is a fast, concurrent CLI utility and `kubectl` plugin (`kubectl-kvtools`) that inspects, audits, and exports comprehensive KubeVirt cluster inventory to a styled, multi-tab Excel (`.xlsx`) workbook, JSON, CSV, or formatted terminal tables—mirroring the operational depth and health audits of VMware's RVTools.

Compatible with **SUSE Harvester**, **OpenShift Virtualization**, **SUSE Virtualization**, and upstream **KubeVirt**.

---

## Features

- 📑 **13 Multi-Tab Excel Sheets**: `kvInfo`, `kvCPU`, `kvMemory`, `kvDisk`, `kvPartition`, `kvNetwork`, `kvCD`, `kvSnapshot`, `kvGuestAgent`, `kvNode`, `kvStoragePool`, `kvHardware`, and `kvHealth`.
- 🔍 **`kvHealth` Best-Practice Engine**: 11 deterministic audit rules detecting live migration blockers, storage zombies, snapshot sprawl, CPU CFS throttling risks, and node vCPU overcommit imbalance.
- ⚡ **High Concurrency & Graceful Degradation**: Concurrently queries cluster objects and guest agent subresources (`/guestosinfo`, `/filesystemlist`) with worker pools; runs safely on clusters lacking optional CRDs (CDI, Snapshot, Multus).
- 📦 **Zero Runtime Dependencies**: Compiled to a single static binary.
- 🔌 **Dual Mode Execution**: Run directly as `kvtools` or as `kubectl kvtools`.

---

## 13 Multi-Tab Inventory Breakdown

| Sheet | Description |
|---|---|
| **`kvInfo`** | Primary VM inventory, power state, node placement, guest OS, firmware/boot, TPM, uptime, and UID. |
| **`kvCPU`** | Cores, sockets, threads, vCPUs, CPU model, dedicated placement, NUMA, requests, and limits. |
| **`kvMemory`** | Guest RAM, requests, limits, launcher overhead, hugepages, and ballooning. |
| **`kvDisk`** | Disks, PVCs, DataVolumes, StorageClasses, provisioned sizes, volume modes, bus types, and CSI drivers. |
| **`kvPartition`** | Guest OS filesystem mount points, filesystem types, capacity, used space, and free %. |
| **`kvNetwork`** | Interfaces, Multus networks, binding modes (masquerade/bridge/sriov), MACs, and Pod/Guest IPs. |
| **`kvCD`** | Attached CD-ROMs, container disks, ISOs, boot orders, and mount states. |
| **`kvSnapshot`** | VM snapshots, volume snapshots, readiness, age, and restorable capacity. |
| **`kvGuestAgent`** | QEMU guest agent connectivity, agent version, hostname, guest kernel, and timezone. |
| **`kvNode`** | Node physical cores, allocatable resources, allocated VM vCPUs/RAM, and vCPU overcommit ratio. |
| **`kvStoragePool`** | StorageClasses, provisioners, reclaim policies, binding modes, and bound PVC totals. |
| **`kvHardware`** | PCI passthrough devices, GPU/vGPU allocations, SR-IOV NICs, and host devices. |
| **`kvHealth`** | Automated migration blockers, storage zombies, and sizing risk findings. |

---

## Installation

### From Source
```bash
git clone https://github.com/coulof/kvtools.git
cd kvtools
make build
# Binary is generated in bin/kvtools and bin/kubectl-kvtools
```

### Install to `$GOPATH/bin`
```bash
make install
```

---

## Usage & Examples

### 1. Export Full Cluster to Excel (`.xlsx`)
```bash
# Auto-names: kvtools_<cluster>_<timestamp>.xlsx
kvtools

# Or specify a custom output file
kvtools -A -o excel -f ./cluster-inventory.xlsx
```

### 2. Run Health & Migration Readiness Audit in Terminal
```bash
kvtools health
# Or
kvtools --health-only
```

### 3. Display Terminal Table for Specific Sheet
```bash
# View VM summary inventory
kvtools -o table kvInfo

# View compute node resource allocation & overcommit
kvtools -o table kvNode
```

### 4. Filter by Namespace
```bash
kvtools -n production -o excel -f prod-vms.xlsx
```

### 5. Export JSON / CSV
```bash
# JSON output to stdout
kvtools -A -o json > inventory.json

# Export all 13 sheets to separate CSV files
kvtools -A -o csv -f ./csv-exports/
```

---

## CLI Flags

| Flag | Shorthand | Default | Description |
|---|---|---|---|
| `--all-namespaces` | `-A` | `true` | Query resources across all namespaces |
| `--namespace` | `-n` | `""` | Filter resources by a specific namespace |
| `--output` | `-o` | `excel` | Output format: `excel`, `json`, `csv`, `table` |
| `--output-file` | `-f` | `""` | Output destination file or directory path |
| `--kubeconfig` | | `""` | Path to kubeconfig file |
| `--context` | | `""` | Kubernetes context to use |
| `--guest-subresources` | | `true` | Query guest agent subresources (`/guestosinfo`, `/filesystemlist`) |
| `--health-only` | | `false` | Only run and display `kvHealth` audit findings |
| `--concurrency` | `-c` | `10` | Worker pool concurrency for parallel subresource queries |
| `--quiet` | `-q` | `false` | Suppress spinner and progress output |

---

## Built-in `kvHealth` Rules

- **`HLTH-001` (CRITICAL)**: VM configured for live migration with ReadWriteOnce (RWO) PVC.
- **`HLTH-002` (WARNING)**: VM configured with `host-passthrough` CPU model.
- **`HLTH-003` (CRITICAL)**: VM mounts local `HostDisk` blocking live migration.
- **`HLTH-004` (WARNING)**: VM has direct SR-IOV NIC without bond failover.
- **`HLTH-005` (WARNING)**: Storage zombie PVC with KubeVirt metadata whose owner UID no longer exists.
- **`HLTH-006` (WARNING)**: CDI DataVolume in `Failed` or stuck `ImportInProgress` state (> 2 hours).
- **`HLTH-007` (WARNING)**: Running VM without active `qemu-guest-agent` connectivity.
- **`HLTH-008` (WARNING)**: `VirtualMachineSnapshot` age exceeding 14 days (snapshot sprawl).
- **`HLTH-009` (WARNING)**: VM CPU limits exceeding 4× CPU requests (CFS throttling risk).
- **`HLTH-010` (CRITICAL)**: VM memory limits equal to memory request with 0 overhead margin (OOM risk).
- **`HLTH-011` (WARNING)**: Node vCPU overcommit ratio exceeding 8:1.

---

## Development & Testing

```bash
# Run unit tests
make test

# Run vet
make vet

# Build local binaries
make build
```

---

## Roadmap

- [ ] **Advanced `kvHealth` Audit Rules**:
  - **Duplicate MAC Detection**: Detect duplicate MAC address assignments across running interfaces and VM specs.
  - **Windows Hyper-V Enlightenments**: Flag Windows guests missing hypervisor enlightenments (`synic`, `relaxed`, `spinlocks`, `vapic`).
  - **Oversized vCPU Allocation**: Flag VMs assigned more vCPUs than any single physical node possesses.
  - **vCPU Hotplug Ceiling Check**: Detect hotplug requests exceeding max socket topology limits.
  - **Retained Snapshot Content Auditing**: Surface orphaned `VolumeSnapshotContent` objects holding SAN/CSI storage with `deletionPolicy: Retain`.
  - **IOThreads & VirtIO-RNG Tuning**: Identify high-throughput disk workloads lacking dedicated IO threads or missing entropy devices.
- [ ] **Additional Inventory Sheets**:
  - **`kvMigration`**: Live migration history, migration durations, source/target nodes, and failure reasons.
  - **`kvEvents`**: Aggregated VM and virt-launcher warning events (OOMKills, scheduling failures, disk attachment timeouts) over the last 24h.
- [ ] **Krew Plugin Index**: Publish `kubectl-kvtools` to the official [Krew index](https://krew.sigs.k8s.io/) (`kubectl krew install kvtools`).

---

## License

Apache License 2.0.
