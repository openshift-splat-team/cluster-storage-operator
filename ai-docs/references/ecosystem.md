# CSO Ecosystem References

Links to Tier 1 platform documentation. All generic patterns live there; this file is CSO's index into that hub.

> **Tier 1 Hub**: https://github.com/openshift/enhancements/tree/master/ai-docs

## Operator Patterns (Tier 1)

| Pattern | Link | CSO Usage |
|---------|------|-----------|
| Controller-runtime reconcile loop | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/platform/operator-patterns) | driver_starter.go, deploymentcontroller.go |
| Status conditions semantics | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/platform/operator-patterns/status-conditions.md) | `<Prefix>DeploymentAvailable` conditions |
| ClusterOperator status rollup | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/platform/operator-patterns) | CSO rolls all driver conditions → `storage` ClusterOperator |
| Stale conditions cleanup | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/platform/operator-patterns) | `staleconditions.NewRemoveStaleConditionsController` |
| FeatureGate access | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/platform/operator-patterns) | `featuregates.NewFeatureGateAccess` for Tech Preview drivers |
| Management state controller | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/platform/operator-patterns) | `managementstatecontroller` in operator_starter.go |

## Testing Practices (Tier 1)

| Practice | Link | CSO Usage |
|----------|------|-----------|
| Test pyramid | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/practices/testing) | Unit: pkg/.../\*_test.go; E2E: test/e2e/ |
| Mock vs real strategies | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/practices/testing) | Fake clients in controller unit tests |
| E2E framework | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/practices/testing) | Go test with build tags in test/e2e/ |

## Security Practices (Tier 1)

| Practice | Link | CSO Relevance |
|----------|------|---------------|
| RBAC guidelines | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/practices/security) | CSO runs cluster-admin (known TODO); driver operators have scoped RBAC |
| Secrets management | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/practices/security) | CredentialsRequests grant cloud IAM to driver operators |

## Kubernetes / OpenShift Fundamentals (Tier 1)

| Concept | Link | CSO Usage |
|---------|------|-----------|
| ClusterOperator | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/domain/openshift) | `storage` ClusterOperator is CSO's health signal |
| CSI (Container Storage Interface) | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/domain/kubernetes) | CSO manages operators for all CSI drivers |
| StorageClass | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/domain/kubernetes) | defaultstorageclass controller ensures one exists |
| ValidatingAdmissionPolicy | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/domain/kubernetes) | volumedatasourcevalidator uses CEL-based VAP |

## Cross-Repo ADRs (Tier 1)

| ADR | Relevance |
|-----|-----------|
| CVO operator lifecycle | CSO is CVO-managed; its manifests follow the `include.release.openshift.io/*` annotation convention |
| Release image bootstrapping | CSO manifests (including CredentialsRequests) are part of the release image |
| HyperShift control plane separation | CSO implements dual-cluster management; see [adr-0003](../decisions/adr-0003-hypershift-split-assets.md) |

## CSO-Specific Cross-Repo Dependencies

| Repo | Relationship |
|------|-------------|
| [openshift/aws-ebs-csi-driver-operator](https://github.com/openshift/aws-ebs-csi-driver-operator) | CSO deploys its Deployment; CredentialsRequest may be owned by CSO or driver |
| [openshift/azure-disk-csi-driver-operator](https://github.com/openshift/azure-disk-csi-driver-operator) | Same pattern |
| [openshift/cloud-credential-operator](https://github.com/openshift/cloud-credential-operator) | Processes `manifests/03_credentials_request_*.yaml` |
| [openshift/api](https://github.com/openshift/api) | Provides `Storage` and `ClusterCSIDriver` types (operator/v1) |
| [openshift/library-go](https://github.com/openshift/library-go) | Provides controller framework, FeatureGate, staleconditions, etc. |
