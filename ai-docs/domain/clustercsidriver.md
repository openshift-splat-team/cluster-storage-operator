# ClusterCSIDriver

**API Group**: `operator.openshift.io/v1`
**Kind**: `ClusterCSIDriver`
**Scope**: Cluster (one per CSI driver, e.g., `ebs.csi.aws.com`)

## Purpose

Per-CSI-driver configuration and status resource. CSO creates one ClusterCSIDriver CR per managed driver, sets its log level, and reads its status to roll up into the `storage` ClusterOperator conditions.

**Key Principle**: CSO creates and owns ClusterCSIDriver CRs. Each CSI driver operator reads its own ClusterCSIDriver to determine log level and reports status back through it.

## Spec Structure

```go
type ClusterCSIDriverSpec struct {
    OperatorSpec      `json:",inline"`    // managementState, logLevel, operatorLogLevel
    StorageClassState StorageClassStateName // Managed | Unmanaged | Removed
    DriverConfig      CSIDriverConfigSpec  // Cloud-specific tuning
}
```

### StorageClassState

| Value | Behavior |
|-------|----------|
| `Managed` (default) | CSI driver operator manages the default StorageClass |
| `Unmanaged` | StorageClass management paused (manual control) |
| `Removed` | CSI driver operator deletes the default StorageClass |

### DriverConfig (Cloud-Specific)

Allows per-driver tuning without modifying the driver operator itself:

| Cloud | Fields |
|-------|--------|
| AWS | `aws.efsVolumeMetrics`, `aws.kmsKeyARN` |
| Azure | `azure.diskEncryptionSet` |
| GCP | `gcp.kmsKey` |
| vSphere | `vsphere.topologyCategories`, `vsphere.globalMaxSnapshotsPerBlockVolume` |

## Status

```go
type ClusterCSIDriverStatus struct {
    OperatorStatus `json:",inline"`  // conditions, generations, observedGeneration
}
```

### Key Conditions

| Condition | Meaning |
|-----------|---------|
| `<Prefix>DeploymentAvailable` | Driver operator Deployment is available |
| `<Prefix>DeploymentProgressing` | Driver operator Deployment is rolling out |
| `<Prefix>DeploymentDegraded` | Driver operator Deployment has failed |
| `Disabled` | Driver is intentionally not running on this platform/sub-platform |

`<Prefix>` comes from `CSIOperatorConfig.ConditionPrefix` (e.g., `AWSEBSCSIDriverOperator`).

The `Disabled` condition is allowed to be `True` without degrading CSO when `CSIOperatorConfig.AllowDisabled` is set (used for sub-platform variance, e.g., Azure Stack Hub).

## Lifecycle

1. **Creation**: CSO creates the ClusterCSIDriver CR from the `CRAsset` path in `CSIOperatorConfig` during driver startup
2. **Update**: `crcontroller.go` reconciles log level from the `Storage` CR into each ClusterCSIDriver
3. **Deletion**: CSO removes the CR when its driver is removed from the platform config

## Example

```yaml
apiVersion: operator.openshift.io/v1
kind: ClusterCSIDriver
metadata:
  name: ebs.csi.aws.com
spec:
  managementState: Managed
  logLevel: Normal
  storageClassState: Managed
```

## Relationship to CSIOperatorConfig

Each `ClusterCSIDriver` CR name corresponds to `CSIOperatorConfig.CSIDriverName`. CSO's `crcontroller.go` reconciles log level from the `Storage` CR into all ClusterCSIDriver CRs it manages.

## Related Concepts

- [storage-cr.md](./storage-cr.md) — Parent operator config that drives log level propagation
- [architecture/components.md](../architecture/components.md) — How CSO creates and manages these CRs
