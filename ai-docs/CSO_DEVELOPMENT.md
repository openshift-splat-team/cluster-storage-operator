# CSO Development Guide

> **Generic Go standards and controller-runtime patterns**: See [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/platform/operator-patterns)

## Quick Start

**Prerequisites**: Go 1.21+, `oc` CLI (for `make update`), OpenShift cluster (for E2E)

```bash
make              # build the operator binary
make update       # regenerate kustomize assets + generated code (requires oc on PATH)
make test-unit    # run unit tests
make verify       # static checks (formatting, generated code freshness)
make check        # verify + test-unit
hack/verify-manifest.sh  # CI check: confirm generated/ is up to date
```

## Repository Structure

```text
cmd/cluster-storage-operator/   # Binary entrypoint

pkg/operator/
├── starter.go                  # RunOperator: mode selection (standalone vs HyperShift)
├── operator_starter.go         # StandaloneStarter + HyperShiftStarter + populateConfigs()
├── csidriveroperator/
│   ├── csioperatorclient/      # Per-driver CSIOperatorConfig files
│   │   ├── types.go            # CSIOperatorConfig struct
│   │   ├── aws.go              # GetAWSEBSCSIOperatorConfig()
│   │   ├── azure-disk.go       # GetAzureDiskCSIOperatorConfig()
│   │   └── ...                 # One file per driver
│   ├── driver_starter.go       # shouldRunController() + per-driver manager launch
│   ├── deploymentcontroller.go # Reconciles Deployment (standalone)
│   ├── hypershift_deployment_controller.go
│   └── crcontroller.go         # Syncs logLevel → ClusterCSIDriver
├── defaultstorageclass/
├── vsphereproblemdetector/
├── volumedatasourcevalidator/
├── metrics/
└── configobservation/

assets/csidriveroperators/<driver>/
├── base/           # Shared SA, RBAC, CR, Deployment (edit here)
├── standalone/     # Standalone overlay + patches (edit here)
│   └── generated/  # ← DO NOT EDIT — output of make update
└── hypershift/
    ├── guest/generated/   # ← DO NOT EDIT
    └── mgmt/generated/    # ← DO NOT EDIT

manifests/          # CVO-managed cluster manifests (applied at install)
test/e2e/           # E2E tests
```

## Development Workflow

### Build

```bash
make
# Produces: ./cluster-storage-operator binary
```

### Regenerate Assets

After editing any file in `assets/csidriveroperators/<driver>/base/`, `standalone/`, or `hypershift/`:

```bash
make update
# Runs: hack/generate-manifests.sh (calls oc kustomize) → updates generated/ dirs
git diff assets/  # review generated output before committing
```

### Local Testing (On-Cluster)

```bash
# Scale down CVO and CSO to avoid conflicts
oc scale --replicas=0 deploy/cluster-version-operator -n openshift-cluster-version
oc scale --replicas=0 deploy/cluster-storage-operator -n openshift-cluster-storage-operator

# Export required image env vars (see README.md for the full list)
export OPERATOR_IMAGE=quay.io/openshift/origin-cluster-storage-operator:latest
export AWS_EBS_DRIVER_OPERATOR_IMAGE=...
# ... etc.

make
./cluster-storage-operator start \
  --kubeconfig $KUBECONFIG \
  --namespace openshift-cluster-storage-operator
```

### Debugging

```bash
# Check CSO status
oc describe clusteroperator/storage

# Check per-driver ClusterCSIDriver conditions
oc get clustercsidriver
oc describe clustercsidriver ebs.csi.aws.com

# CSO logs
oc logs -f deployment/cluster-storage-operator \
  -n openshift-cluster-storage-operator

# Check if assets are fresh
hack/verify-manifest.sh
```

## Adding a New CSI Driver

1. **Create asset directory**:
   ```
   assets/csidriveroperators/<driver>/base/
   assets/csidriveroperators/<driver>/standalone/
   assets/csidriveroperators/<driver>/hypershift/guest/  (if HyperShift supported)
   assets/csidriveroperators/<driver>/hypershift/mgmt/   (if HyperShift supported)
   ```

2. **Create CSIOperatorConfig**:
   ```go
   // pkg/operator/csidriveroperator/csioperatorclient/<driver>.go
   func GetMyDriverCSIOperatorConfig() csioperatorclient.CSIOperatorConfig {
       return csioperatorclient.CSIOperatorConfig{
           CSIDriverName:           "mydriver.csi.example.com",
           CSIDriverDeploymentName: "my-driver-operator",
           ConditionPrefix:         "MyDriverCSIDriverOperator",
           Platform:                configv1.SomePlatformType,
           StaticAssets:            []string{"assets/csidriveroperators/mydriver/standalone/generated/..."},
           DeploymentAsset:         "assets/csidriveroperators/mydriver/standalone/generated/deployment.yaml",
           CRAsset:                 "assets/csidriveroperators/mydriver/standalone/generated/clustercsidriver.yaml",
           ImageReplacer:           strings.NewReplacer("${MYDRIVER_OPERATOR_IMAGE}", os.Getenv("MYDRIVER_OPERATOR_IMAGE")),
           RequireFeatureGate:      "MyDriverCSIDriver",  // all new drivers start as Tech Preview
       }
   }
   ```

3. **Register in** `operator_starter.go` `StandaloneStarter.populateConfigs()` and `HyperShiftStarter.populateConfigs()`

4. **Add images** to `manifests/image-references` and env vars to `manifests/10_deployment.yaml`

5. **Add to kustomize generator** in `hack/generate-manifests.sh` if using kustomize generation

6. **Add CredentialsRequest** at `manifests/03_credentials_request_<driver>.yaml` if cloud IAM needed

7. **Run** `make update && make check`

## Common Tasks

| Task | Command |
|------|---------|
| Regenerate assets after source change | `make update` |
| Verify assets are fresh (CI check) | `hack/verify-manifest.sh` |
| Run unit tests | `make test-unit` |
| Run all checks | `make check` |
| Build binary | `make` |

## Component-Specific Notes

- **`oc` required for `make update`**: The kustomize generation step uses `oc kustomize` (not standalone kustomize) to support OpenShift-specific patches
- **HyperShift namespace placeholder**: Management cluster assets must use `${CONTROLPLANE_NAMESPACE}` for any namespace reference — substituted at runtime
- **Image env vars**: All container images are passed as env vars to the CSO Deployment. The full list is in `README.md` and `manifests/10_deployment.yaml`
- **Condition prefix**: Choose a unique `ConditionPrefix` per driver — it prefixes all ClusterOperator conditions and must not conflict with existing prefixes
