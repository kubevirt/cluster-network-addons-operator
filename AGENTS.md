This file provides guidance when working with code in this repository.

## Repository Overview

Cluster Network Addons Operator (CNAO) is a Kubernetes operator that deploys and manages networking add-on components on Kubernetes and OpenShift clusters. This repository contains:

- The CNAO operator code
- Manifest generation for managed network components
- E2E test suites
- Dev/test environment setup (kubevirtci)
- Installation manifests for Operator Lifecycle Manager (OLM)

### How It Works

CNAO follows the standard Kubernetes operator pattern using controller-runtime. The operator watches a single cluster-scoped `NetworkAddonsConfig` CR and reconciles the desired state by rendering and applying manifests for each enabled network component.

### Core Components

- **cmd/manager/** - Operator entry point; sets up the controller-runtime manager, scheme, metrics server, and health probes
- **pkg/controller/networkaddonsconfig/** - Main controller loop (`networkaddonsconfig_controller.go`) and pod controller for status tracking
- **pkg/network/** - Per-component logic; each file (e.g., `multus.go`, `kubemacpool.go`, `ovs.go`) knows how to render manifests for one component
- **pkg/components/** - Manifest templates and default image references for all managed components
- **pkg/apply/** - Manifest application and three-way merge logic for safe upgrades
- **pkg/apis/networkaddonsoperator/** - CRD type definitions (shared types used by both `v1` and `v1alpha1` API versions)

### Managed Components

Each component has deployment manifests under `data/` and rendering logic under `pkg/network/`:

| Component | Description |
|-----------|-------------|
| Multus | Multi-network CNI meta-plugin |
| Multus Dynamic Networks | Hot-plug/hot-unplug of pod interfaces |
| Linux Bridge | Linux bridge CNI plugin |
| OVS | Open vSwitch CNI plugin |
| KubeMacPool | MAC address pool manager |
| Macvtap | Macvtap CNI plugin |
| KubeSecondaryDNS | Secondary DNS for VM interfaces |
| KubevirtIpamController | IPAM controller for secondary networks |

### Key Directories

```
cmd/manager/          Entry point for the operator binary
pkg/apis/             CRD types (NetworkAddonsConfig v1 and v1alpha1)
pkg/controller/       Reconciliation controllers
pkg/network/          Per-component manifest rendering
pkg/components/       Default images, manifest templates
pkg/apply/            Manifest application and merge logic
pkg/monitoring/       Prometheus metrics and alerts
data/                 Raw YAML templates for each managed component
test/e2e/             E2E test suites (lifecycle, workflow, monitoring, compliance)
hack/                 Build, release, and component-bump scripts
tools/                Build helpers (bumper, manifest-templator, metrics docs)
cluster/              Local dev cluster management (kubevirtci)
manifests/            Release manifests for OLM
templates/            ClusterServiceVersion templates for OLM (Operator Lifecycle Manager) deployment
automation/           CI script wrappers
```

## Development Commands

### Build

```bash
make manager               # Build the operator binary
make manifest-templator    # Build the manifest templator tool
make docker-build          # Build operator and registry container images
```

### Code Quality

```bash
make check                 # Full validation: whitespace, vet, goimports, gen-k8s, lint, unit tests
make fmt                   # Auto-format (whitespace + goimports)
make lint                  # Run golangci-lint
make vet                   # Run go vet
```

### Testing

```bash
make test/unit             # Run unit tests
make test/e2e/lifecycle    # E2E: operator deployment and upgrades
make test/e2e/workflow     # E2E: component deployment workflows
make test/e2e/monitoring   # E2E: Prometheus metrics and alerts
make test/e2e/compliance   # E2E: TLS compliance
```

### Code Generation

```bash
make gen-k8s               # Generate deepcopy methods from CRD types
make gen-k8s-check         # Verify generated code is up to date
make gen-manifests         # Generate operator deployment manifests
```

### Local Development Cluster

```bash
make cluster-up            # Start local kubevirtci cluster with cert-manager
make cluster-down          # Tear down local cluster
make cluster-sync          # Push operator image and install on cluster
make cluster-clean         # Remove operator from cluster
```

### Dependencies

```bash
make vendor                # Tidy and vendor Go modules (enforces max Go version)
```

### Component Bumps

```bash
make bump-<component>      # Bump a single component (e.g., bump-multus, bump-ovs)
make bump-all              # Bump all components
make bump-kubevirtci       # Update kubevirtci version
```

## Build Toolchain

- **Go version**: 1.25 (max allowed; auto-installed to `build/_output/bin/go/` via `hack/install-go.sh`)
- **Build flags**: `GOFLAGS=-mod=vendor GO111MODULE=on CGO_ENABLED=0`
- **Multi-arch**: Builds for `linux/amd64`, `linux/arm64`, `linux/s390x`
- **Container runtime**: Auto-detects podman or docker (`OCI_BIN`)

## Testing

### Unit Tests

- Located alongside source files in `pkg/`
- Framework: Ginkgo v2 with Gomega assertions
- Run with `make test/unit`

### E2E Tests

- Located in `test/e2e/` with four suites: lifecycle, workflow, monitoring, compliance
- Require a running cluster (local via kubevirtci or external via `KUBECONFIG`)
- Default timeout: 3 hours (`E2E_TEST_TIMEOUT`)
- Test infrastructure helpers in `test/libframework/`, `test/operations/`, `test/check/`

### Test Naming

Use Ginkgo-style descriptive names: `When <condition>... it should <behavior>...`

## CRD: NetworkAddonsConfig

- **API Group**: `networkaddonsoperator.network.kubevirt.io`
- **Versions**: `v1` (current), `v1alpha1` (legacy)
- **Scope**: Cluster-scoped (single instance per cluster)
- **Spec fields**: Each managed component has an optional field to enable/configure it
- **Type definitions**: `pkg/apis/networkaddonsoperator/shared/networkaddonsconfig_types.go`

After modifying types, run `make gen-k8s` to regenerate deepcopy methods.

## Key Conventions

- Dependencies are vendored; run `make vendor` after changing `go.mod`
- Do not modify `vendor/` directly
- Run `make check` before submitting PRs
- Component default images are defined in `pkg/components/components.go`
- Component version pins are tracked in `components.yaml` at the repo root
- Deployment manifest templates live in `data/<component>/` as raw YAML
- The operator renders final manifests at runtime by patching images and configuration into these templates

## Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `KUBECONFIG` | Path to kubeconfig | - |
| `KUBEVIRT_PROVIDER` | Cluster provider for local dev | - |
| `IMAGE_REGISTRY` | Container image registry | `quay.io/kubevirt` |
| `IMAGE_TAG` | Container image tag | `latest` |
| `OCI_BIN` | Container runtime (`podman`/`docker`) | auto-detected |
| `E2E_TEST_TIMEOUT` | E2E test timeout | `3h` |
| `CNAO_LOG_LEVEL` | Operator log level (-1=debug, 2=error) | - |
| `VERSION` | Build version | `99.0.0` |
