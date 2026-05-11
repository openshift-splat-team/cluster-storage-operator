# ADR-0002: Kustomize Asset Generation for CSI Driver Manifests

**Date**: 2021-01-01
**Status**: Accepted

## Context

Each CSI driver operator needs a Deployment, ServiceAccount, ClusterRoleBinding, and ClusterCSIDriver CR. Standalone and HyperShift deployments share a base but differ in namespace, image env vars, and which resources are applied to which cluster.

Initially, assets were maintained as independent YAML files per driver per topology, leading to drift between standalone and HyperShift variants and between drivers.

## Decision

Use `kustomize` with a `base/` layer and `standalone/`, `hypershift/guest/`, and `hypershift/mgmt/` overlays per driver. The `generated/` output (produced by `hack/generate-manifests.sh` running `oc kustomize`) is committed to the repo and enforced in CI by `hack/verify-manifest.sh`.

Drivers with simple or non-standard assets (gcp-pd, ibm-vpc-block, powervs-block, vsphere) use static YAML without kustomize generation.

## Rationale

Kustomize base/overlay enforces that standalone and HyperShift share the same SA name, RBAC structure, and CR format — differences are patches, not forks. CI freshness enforcement prevents manual edits to `generated/` from drifting.

### Alternatives Considered

| Option | Pros | Cons |
|--------|------|------|
| Kustomize + committed generated/ (chosen) | Diff reviewable in PRs, CI-enforced | generated/ can go stale; requires `oc` on PATH for make update |
| Helm | Familiar to many contributors | Adds dependency; overkill for YAML composition |
| Pure Go templating | No external tools | Hard to read and maintain complex YAML patches |
| Per-topology independent YAMLs | Maximum flexibility | Drift between standalone/HyperShift; duplicated bugs |

## Consequences

**Positive**: Standalone and HyperShift patches are visible as kustomize overlays. `hack/verify-manifest.sh` catches forgotten `make update` runs in CI. New driver variants require only new patches, not full YAML duplication.

**Negative**: Contributors must run `make update` (which requires `oc` on PATH) after any source file change. Reviewing `generated/` diffs adds noise to PRs. Drivers not using kustomize are inconsistent with generated drivers and require manual sync discipline.
