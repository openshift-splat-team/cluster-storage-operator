# CSO Testing Guide

> **Test pyramid philosophy and E2E framework patterns**: See [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/practices/testing)

## Test Organization

```text
pkg/operator/**/*_test.go     # Unit tests (co-located with source)
test/e2e/                     # E2E tests (build tag: e2e)
test/manifest/                # Test manifests
cmd/cluster-storage-operator-tests-ext/  # OTE external test binary entrypoint
```

## Unit Tests

```bash
make test-unit
# Or directly:
go test -v ./pkg/...
go test -v ./pkg/operator/csidriveroperator/...
go test -v ./pkg/operator/defaultstorageclass/...
```

**CSO-Specific Unit Test Patterns**:

- **Fake clients**: Use `controller-runtime/pkg/client/fake` for Kubernetes API interactions; no real cluster needed
- **CSIOperatorConfig injection**: Tests construct `CSIOperatorConfig` directly and pass it to controllers, bypassing `populateConfigs()`
- **Mode testing**: Standalone and HyperShift code paths are tested by wiring different client sets
- **Asset loading**: Tests verify bindata assets are valid YAML and contain expected fields; use `assets.Asset()` directly
- **Platform filtering**: Unit tests exercise `shouldRunController()` with different platform/feature-gate combinations

**Key test files**:
- `pkg/operator/csidriveroperator/*_test.go` — controller unit tests
- `pkg/operator/defaultstorageclass/*_test.go` — StorageClass controller tests
- `pkg/operator/volumedatasourcevalidator/*_test.go` — admission policy tests

## E2E Tests

**Prerequisites**: Running OpenShift cluster with `KUBECONFIG` set

```bash
go test -tags e2e -v ./test/e2e/...
```

**Tests cover**:
- CSI driver operator Deployment becomes Available
- ClusterCSIDriver CR created and conditions healthy
- Default StorageClass exists for the cluster platform
- vSphere problem detector running on vSphere clusters

### Extended Tests (OTE)

```bash
# Build the external test binary
go build ./cmd/cluster-storage-operator-tests-ext/...

# Run via openshift-tests
openshift-tests run --provider <platform> ./cluster-storage-operator-tests-ext
```

## Manifest Freshness Verification

```bash
# CI check — same command run in CI
hack/verify-manifest.sh
# Fails if generated/ assets are out of date relative to kustomize sources
```

This is technically a static check, not a test, but it is one of the most common CI failures.

## Test Coverage

**Stronger coverage**:
- `csidriveroperator` controller logic (shouldRunController, deployment reconciliation)
- `defaultstorageclass` controller (platform case coverage)
- `volumedatasourcevalidator` (CEL expression correctness)

**Known gaps**:
- HyperShift-specific controller paths (require dual-client test harness)
- `vsphereproblemdetector` (limited unit coverage; mostly E2E)
- Per-driver `CSIOperatorConfig` functions (no unit tests verify image replacer env var names match deployment YAML placeholders)

## Debugging Test Failures

### Unit Test Failures

```bash
go test -v -run TestMyController ./pkg/operator/csidriveroperator/...
go test -race ./pkg/...  # check for data races
```

### E2E Test Failures

```bash
# Check ClusterCSIDriver conditions
oc describe clustercsidriver

# Check driver operator pod
oc get pods -n openshift-cluster-csi-drivers
oc logs -n openshift-cluster-csi-drivers deployment/aws-ebs-csi-driver-operator

# Must-gather
oc adm must-gather -- /usr/bin/gather
```

### Common Failures

| Symptom | Likely Cause |
|---------|-------------|
| `hack/verify-manifest.sh` fails | `make update` not run after editing source assets |
| `<Prefix>DeploymentDegraded` | Image pull failure (wrong image env var) or RBAC missing |
| Default StorageClass missing | Driver operator not Running; check ClusterCSIDriver status |
| HyperShift driver not starting | `MgmtStaticAssets` missing or `${CONTROLPLANE_NAMESPACE}` placeholder issue |

## CI Integration

- `make check` (verify + test-unit) runs on every PR
- E2E tests run in cloud-specific CI lanes per driver
- `hack/verify-manifest.sh` runs in CI to enforce generated asset freshness
