# ADR-0003: Split Asset Topology for HyperShift Support

**Date**: 2022-01-01
**Status**: Accepted

## Context

HyperShift places the control plane (including CSI driver operator Deployment) on a management cluster, while the data plane (guest cluster) receives ServiceAccount, RBAC, and ClusterCSIDriver CR. A single asset set cannot serve both clusters.

CSO needed a way to apply different resources to two different clusters from a single operator instance without changing the generic controller logic.

## Decision

Split assets into two categories per driver:

- **`StaticAssets`** / kustomize `standalone/` or `hypershift/guest/` → applied to the guest/standalone cluster (via `commonClients`)
- **`MgmtStaticAssets`** / `hypershift/mgmt/` → applied to the management cluster (via a separate management client)

Assets in `hypershift/mgmt/` use `${CONTROLPLANE_NAMESPACE}` as a namespace placeholder, substituted at runtime by `namespaceReplacer()` in `driver_starter.go`.

The `HyperShiftStarter` wires up a second `csoclients.Clients` instance pointed at the management cluster for applying `MgmtStaticAssets`.

## Rationale

Keeping the split at the asset configuration level (fields in `CSIOperatorConfig`) means the generic `driver_starter.go` and `deploymentcontroller.go` need no per-driver knowledge. The guest vs. management split is entirely data-driven.

### Alternatives Considered

| Option | Pros | Cons |
|--------|------|------|
| Split in CSIOperatorConfig (chosen) | Data-driven, no controller changes per driver | `MgmtStaticAssets` field adds config complexity |
| Separate operator instances | Clean isolation | Operational complexity; state not shared |
| Single asset set with topology selectors | One file | Complex selector logic; different resources per cluster |

## Consequences

**Positive**: HyperShift support for a driver is a config change (add `MgmtStaticAssets`), not a controller change. The management client is set up once in `HyperShiftStarter`; all drivers reuse it.

**Negative**: Adding a new driver for HyperShift requires understanding which resources go to each cluster — documentation gap that has caused bugs (RBAC applied to wrong cluster). Drivers not supporting HyperShift must be explicitly registered with empty `MgmtStaticAssets` or excluded from `HyperShiftStarter.populateConfigs()`.
