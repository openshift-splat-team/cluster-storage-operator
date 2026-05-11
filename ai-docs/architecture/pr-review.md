# CSO PR Review Guide

CSO-specific review checklist and failure conditions. Read before reviewing any PR.

> **For generic review patterns**: See [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/practices)

## Critical Review Order

**Check in this order for every PR:**

1. **Generated assets up to date** — if `base/`, `standalone/`, or `hypershift/` source files changed, run `hack/verify-manifest.sh` or confirm CI passed. `generated/` must never be edited by hand.

2. **Standalone and HyperShift symmetry** — driver RBAC/Deployment/CR changes almost always need both paths. Check both `StandaloneStarter.populateConfigs()` and `HyperShiftStarter.populateConfigs()` in `operator_starter.go`.

3. **`image-references` complete** — every new `${...}_IMAGE` placeholder in a Deployment asset must have a matching entry in `manifests/image-references`.

4. **Manifest topology annotations correct** — new manifests in `manifests/` must carry `include.release.openshift.io/*` annotations or they are silently excluded from some topologies.

5. **RBAC least-privilege** — new or modified RBAC must use minimum required permissions with correct subjects and namespaces.

6. **Feature gate status** — new Tech Preview drivers must set `RequireFeatureGate`; GA drivers must not.

7. **CredentialsRequest changes** — if any `03_credentials_request_*.yaml` is modified, cross-repo coordination is required (see below).

## PR Review Failure Conditions

A PR **must not be approved** if any of the following are true:

- `generated/` assets are out of date relative to kustomize source overlays
- A new container image is used in an asset but absent from `manifests/image-references`
- A new manifest in `manifests/` is missing `include.release.openshift.io/*` annotations
- A new Tech Preview driver does not set `RequireFeatureGate`
- A GA driver still has `RequireFeatureGate` set
- RBAC rules grant broader permissions than the component demonstrably needs
- A `ValidatingAdmissionPolicy` change uses `failurePolicy: Fail` without upgrade safety analysis
- A driver is registered in `StandaloneStarter.populateConfigs()` but HyperShift support decision is undocumented
- Any `manifests/03_credentials_request_*.yaml` is modified without the PR description covering the required cross-repo coordination

## RBAC Review

- RBAC under `assets/csidriveroperators/<driver>/base/` governs the CSI driver **operator**, not the driver itself
- CSO itself runs with `cluster-admin` (`manifests/08_operator_rbac.yaml`) — known TODO, not a new concern per PR
- Sidecar RBAC in `manifests/09_sidecar-*.yaml` is **shared across all drivers** — changes affect every driver simultaneously

## Adding a New CSI Driver

Verify the PR includes **all** of:

1. `assets/csidriveroperators/<driver>/` with `base/`, `standalone/`, and optionally `hypershift/guest/` + `hypershift/mgmt/`
2. `GetXxxCSIOperatorConfig()` function in `pkg/operator/csidriveroperator/csioperatorclient/<driver>.go`
3. Registration in `StandaloneStarter.populateConfigs()` in `operator_starter.go`
4. Registration in `HyperShiftStarter.populateConfigs()` if HyperShift is supported (or explicit note in PR that it isn't)
5. Driver and operator images in `manifests/image-references`
6. Driver added to `drivers` array in `hack/generate-manifests.sh` if it uses kustomize generation
7. `manifests/03_credentials_request_<driver>.yaml` if cloud credentials are required
8. `RequireFeatureGate` set (all new drivers start as Tech Preview)

## CredentialsRequest Changes

`manifests/03_credentials_request_*.yaml` define cloud IAM permissions granted to each CSI driver via CCO.

**Any modification requires cross-repo coordination.** Post this comment when detected:

> **Action required — CredentialsRequest change detected.**
>
> Modifying a `CredentialsRequest` changes cloud IAM permissions for a CSI driver. Coordinate:
>
> 1. **CSI driver operator repo** — confirm whether the driver operator repo has its own copy of this CR or relies on CSO's. Update both if needed.
> 2. **AWS STS / manual-mode** — for STS or manual-mode CCO clusters, IAM policies must also be updated in the installer or customer-managed policy documents. Adding a new `ec2:*` action here is not sufficient.
> 3. **Cloud Credential Operator** — if a new `ProviderSpec` field or provider kind is used, CCO may need to be updated first.
> 4. **Release notes** — new IAM permissions are customer-visible; note them especially for manual-mode users managing their own policies.

**Per-cloud diff review:**

| Cloud | Pattern | What to Check |
|-------|---------|---------------|
| AWS (`AWSProviderSpec`) | `statementEntries[].action` | Confirm the new action is actually called by driver code; `resource: "*"` is standard |
| Azure (`AzureProviderSpec`) | `permissions: Microsoft.<svc>/<res>/<action>` | Confirm permissions match driver requirements |
| GCP (`GCPProviderSpec`) | `predefinedRoles` | Prefer fine-grained `permissions` entries over broad roles |
| All | Removed permissions | Confirm driver no longer calls the corresponding API |

## Admission Policy Changes

`manifests/13_validating_admission_policy.yaml` uses `failurePolicy: Fail`. Any CEL expression change must be reviewed for:

- **Correctness**: The policy validates `storage.openshift.io/fsgroup-change-policy` and `storage.openshift.io/selinux-change-policy` namespace labels
- **Upgrade safety**: A broken policy with `failurePolicy: Fail` blocks namespace creation cluster-wide
