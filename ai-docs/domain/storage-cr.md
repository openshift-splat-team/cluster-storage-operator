# Storage (Operator CR)

**API Group**: `operator.openshift.io/v1`
**Kind**: `Storage`
**Scope**: Cluster (singleton: `cluster`)

## Purpose

Operator configuration resource for the Cluster Storage Operator. Installed by CVO at cluster bootstrap; controls CSO's management state and log levels.

**Key Principle**: This CR is created once by the installer (`release.openshift.io/create-only: "true"`) and never recreated by CSO itself. Deleting it is destructive.

## Spec Structure

```go
// From openshift/api operator/v1
type StorageSpec struct {
    OperatorSpec `json:",inline"`  // ManagementState, LogLevel, OperatorLogLevel
}
```

Inherits the standard `OperatorSpec`:

| Field | Values | Purpose |
|-------|--------|---------|
| `managementState` | `Managed` \| `Unmanaged` \| `Removed` | Controls whether CSO actively reconciles |
| `logLevel` | `Normal` \| `Debug` \| `Trace` \| `TraceAll` | Component log level |
| `operatorLogLevel` | Same | Operator binary log level |

## Key Concepts

### ManagementState

- **`Managed`** (default): CSO actively reconciles all CSI driver operators and resources
- **`Unmanaged`**: CSO stops reconciling but does not remove resources (manual takeover)
- **`Removed`**: CSO removes all managed resources (destructive; rarely used)

### Manifest Topology Annotations

The `Storage` CR manifest (`manifests/06_operator_cr.yaml`) carries topology annotations:

```yaml
include.release.openshift.io/hypershift: "true"
include.release.openshift.io/ibm-cloud-managed: "true"
include.release.openshift.io/self-managed-high-availability: "true"
include.release.openshift.io/single-node-developer: "true"
capability.openshift.io/name: Storage
```

A separate `06_operator_cr-hypershift.yaml` exists for HyperShift-specific overrides.

## Lifecycle

1. **Creation**: CVO applies `manifests/06_operator_cr.yaml` at cluster bootstrap (create-only; not updated by CVO after)
2. **Update**: Admin can change `logLevel` / `managementState` via `oc edit storage cluster`
3. **Never deleted** in normal operation

## Example

```yaml
apiVersion: operator.openshift.io/v1
kind: Storage
metadata:
  name: cluster
spec:
  managementState: Managed
  logLevel: Normal
  operatorLogLevel: Normal
```

## Related Concepts

- [ClusterCSIDriver](./clustercsidriver.md) — Per-driver config managed by CSO
- [architecture/components.md](../architecture/components.md) — How CSO consumes this CR
