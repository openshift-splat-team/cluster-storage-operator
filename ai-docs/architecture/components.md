# CSO Component Architecture

## Overview

CSO is a meta-operator: it does not implement storage directly. It installs and lifecycle-manages the operator for each platform's CSI driver via a registry of `CSIOperatorConfig` structs.

## Repository Layout

```text
cmd/
└── cluster-storage-operator/   # Operator binary entrypoint

pkg/operator/
├── starter.go                  # RunOperator: selects Standalone vs HyperShift
├── operator_starter.go         # StandaloneStarter + HyperShiftStarter + CSIOperatorConfig registry
├── csidriveroperator/
│   ├── csioperatorclient/      # Per-driver CSIOperatorConfig (aws.go, azure-disk.go, …)
│   │   └── types.go            # CSIOperatorConfig struct definition
│   ├── driver_starter.go       # shouldRunController() + per-driver manager launch
│   ├── deploymentcontroller.go # Reconciles driver operator Deployment (standalone)
│   ├── hypershift_deployment_controller.go  # Reconciles Deployment on mgmt cluster
│   └── crcontroller.go         # Syncs log level → ClusterCSIDriver CRs
├── defaultstorageclass/        # Ensures a default StorageClass exists
├── vsphereproblemdetector/     # vSphere problem detector lifecycle
├── volumedatasourcevalidator/  # ValidatingAdmissionPolicy for namespace labels
├── metrics/                    # StorageClass + VolumeAttributesClass Prometheus metrics
└── configobservation/          # Proxy config observation

assets/
└── csidriveroperators/<driver>/
    ├── base/                   # Shared SA, RBAC, ClusterCSIDriver CR, Deployment
    ├── standalone/
    │   └── generated/          # ← DO NOT EDIT: output of `make update`
    └── hypershift/
        ├── guest/generated/    # ← DO NOT EDIT
        └── mgmt/generated/     # ← DO NOT EDIT
```

## Two Deployment Modes

### Standalone (self-managed OCP)

CSO runs on the same cluster it manages. All resources (Deployment, ServiceAccount, RBAC, ClusterCSIDriver CR) live in `openshift-cluster-csi-drivers`.

```
CSO (openshift-cluster-storage-operator)
  └── deploys driver operator Deployment → openshift-cluster-csi-drivers
         └── driver operator manages CSI driver DaemonSet/Deployment
```

### HyperShift (hosted clusters)

CSO runs on a **management cluster**, managing a **guest cluster**. Assets are split:

```
Management cluster:
  CSO → deploys driver operator Deployment (per-tenant control plane namespace)

Guest cluster:
  CSO → applies ServiceAccount, RBAC, ClusterCSIDriver CR
```

Mode is selected in `starter.go` based on `--guest-kubeconfig` flag presence. HyperShift assets use `${CONTROLPLANE_NAMESPACE}` placeholder substituted at runtime by `namespaceReplacer()` in `driver_starter.go`.

## CSIOperatorConfig Registry

Each driver is described by a `CSIOperatorConfig` struct, registered in `operator_starter.go`'s `populateConfigs()`. Fields:

| Field | Purpose |
|-------|---------|
| `CSIDriverName` | CSI driver ID (e.g., `ebs.csi.aws.com`) = ClusterCSIDriver CR name |
| `CSIDriverDeploymentName` | Deployment name (e.g., `aws-ebs-csi-driver-operator`) |
| `ConditionPrefix` | Prefix for ClusterOperator conditions (e.g., `AWSEBSCSIDriverOperator`) |
| `Platform` | Target platform, or `AllPlatforms` |
| `StatusFilter` | Optional sub-platform callback (e.g., Azure Stack Hub exclusion) |
| `StaticAssets` | YAML assets applied to guest/standalone cluster |
| `MgmtStaticAssets` | YAML assets applied to management cluster (HyperShift) |
| `DeploymentAsset` | Path to driver operator Deployment asset |
| `CRAsset` | Path to ClusterCSIDriver CR asset |
| `ImageReplacer` | `strings.Replacer` substituting `${DRIVER_IMAGE}`, `${OPERATOR_IMAGE}`, etc. |
| `AllowDisabled` | `true` = Disabled condition does not degrade CSO (sub-platform variance) |
| `RequireFeatureGate` | Feature gate name; if set, driver is Tech Preview |

## Platform Filtering

`shouldRunController()` in `driver_starter.go` starts a driver only when ALL of:
1. Cluster platform matches `cfg.Platform` (or `AllPlatforms`)
2. `cfg.StatusFilter` returns `true` (if non-nil)
3. `cfg.RequireFeatureGate` is enabled (if set)
4. No unmanaged third-party CSI driver with the same name exists (no `csi.openshift.io/managed` annotation)

## Asset Generation (Kustomize)

Assets under `generated/` are produced by `hack/generate-manifests.sh` running `oc kustomize`. **Never edit `generated/` by hand.**

| Driver | Generation |
|--------|-----------|
| aws-ebs, azure-disk, azure-file, openstack-cinder, openstack-manila | kustomize generated |
| gcp-pd, ibm-vpc-block, powervs-block, vsphere | static (no generated/) |

Workflow: edit `base/`, `standalone/`, or `hypershift/` → run `make update` → commit `generated/`.
CI enforces freshness via `hack/verify-manifest.sh`.

## Image Substitution Pattern

Deployment assets contain placeholders: `${DRIVER_IMAGE}`, `${OPERATOR_IMAGE}`, `${PROVISIONER_IMAGE}`, etc.

At runtime, CSO builds a `strings.NewReplacer` from env vars (defined as constants in `csioperatorclient/<driver>.go`) and applies it to the Deployment YAML before applying it to the cluster.

Adding a new placeholder requires:
1. Constant + `os.Getenv()` in `csioperatorclient/<driver>.go`
2. Entry in `manifests/image-references`
3. Env var in `manifests/10_deployment.yaml`

## Image References

`manifests/image-references` is an `ImageStream` consumed by OpenShift's ART build system to pin image digests in the release payload. Every container image used in any asset must have an entry here.

## Manifest Topology Annotations

All manifests in `manifests/` must carry topology annotations or they are silently excluded:

```yaml
include.release.openshift.io/hypershift: "true"
include.release.openshift.io/ibm-cloud-managed: "true"
include.release.openshift.io/self-managed-high-availability: "true"
include.release.openshift.io/single-node-developer: "true"
capability.openshift.io/name: Storage
```

HyperShift-specific variants (e.g., `10_deployment-hypershift.yaml`) omit the `self-managed-*` annotation.

## Condition Naming Convention

Conditions follow `<ConditionPrefix><ConditionType>`:
- Example: `AWSEBSCSIDriverOperatorDeploymentAvailable`
- Prefix set in `CSIOperatorConfig.ConditionPrefix`
- Stale conditions cleaned up by `staleconditions.NewRemoveStaleConditionsController`

## Default StorageClass Controller

`pkg/operator/defaultstorageclass/controller.go` — currently all major platforms return `supportedByCSIError`, meaning the default StorageClass is the CSI driver operator's responsibility. Only add a new `case` here if CSO genuinely owns the default StorageClass for a platform (not the driver operator).

## CredentialsRequests

`manifests/03_credentials_request_*.yaml` define cloud IAM permissions per driver via CCO. These are applied at cluster install time. Modifying them requires cross-repo coordination — see [pr-review.md](./pr-review.md).
