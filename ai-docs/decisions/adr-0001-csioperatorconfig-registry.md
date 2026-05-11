# ADR-0001: CSIOperatorConfig Registry Pattern

**Date**: 2020-01-01
**Status**: Accepted

## Context

CSO must manage 9+ CSI driver operators across multiple cloud platforms. Each driver has a different Deployment, RBAC, ClusterCSIDriver CR, set of images, and platform conditions. The logic for starting each driver is similar but parameterized differently.

A naive approach would be a large switch statement in the controller with per-driver logic duplicated throughout.

## Decision

Define a `CSIOperatorConfig` struct that captures all per-driver configuration declaratively. Register all drivers in a `populateConfigs()` function in `operator_starter.go`. The generic `driver_starter.go` iterates the registry and starts each driver using only the config struct — no per-driver conditional logic in the controller.

## Rationale

The registry pattern decouples driver-specific data from driver-lifecycle logic. Adding a new driver requires only:
1. A new `GetXxxCSIOperatorConfig()` function in `csioperatorclient/`
2. Registration in `populateConfigs()`

No changes to any controller code.

### Alternatives Considered

| Option | Pros | Cons |
|--------|------|------|
| CSIOperatorConfig registry (chosen) | Clean, extensible, no controller changes per driver | Struct must be kept stable; new capabilities require adding fields |
| Per-driver controllers | Maximum per-driver flexibility | 9× code duplication, independent bugs |
| Plugin/dynamic loading | True runtime extensibility | Excessive complexity for a compile-time problem |

## Consequences

**Positive**: Adding a driver is a ~50-line config function. The generic controller is unchanged per new driver. Testing the controller doesn't require per-driver test cases for lifecycle.

**Negative**: `CSIOperatorConfig` has grown many fields over time (`AllowDisabled`, `RequireFeatureGate`, `MgmtStaticAssets`, `ExtraControllers`). The struct is effectively a mini-DSL and new contributors must understand all fields before adding a driver.
