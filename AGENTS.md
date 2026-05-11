# Cluster Storage Operator - Agentic Documentation

**Component**: Cluster Storage Operator (CSO)
**Repository**: openshift/cluster-storage-operator
**Documentation Tier**: 2 (Component-specific)

> **Generic Platform Patterns**: See [Tier 1 Ecosystem Hub](https://github.com/openshift/enhancements/tree/master/ai-docs) for operator patterns, testing practices, security guidelines, and cross-repo ADRs.

> **Retrieval-first**: Read `ai-docs/architecture/components.md` before editing controller or asset code. Read `ai-docs/architecture/pr-review.md` before reviewing any PR.

## What is CSO?

Deploys and lifecycle-manages per-platform CSI driver operators (AWS EBS, Azure Disk, Azure File, GCP PD, IBM VPC Block, OpenStack Cinder/Manila, PowerVS Block, vSphere). Ensures a default StorageClass exists and enforces storage admission policies. CSO does **not** implement CSI drivers — it manages the *operator* for each driver.

**Key Principle**: Platform-specific storage is owned by its driver operator; CSO is the meta-operator that installs and watches them all.

## Core Components

- **csidriveroperator**: Starts per-platform driver managers; `CSIOperatorConfig` registry in `csioperatorclient/`
- **defaultstorageclass**: Ensures a default StorageClass exists per platform
- **vsphereproblemdetector**: Lifecycle-manages the vSphere problem detector on vSphere clusters
- **volumedatasourcevalidator**: Manages the namespace label `ValidatingAdmissionPolicy`
- **metrics**: StorageClass and VolumeAttributesClass Prometheus metrics

**Quick Start**: `oc describe clusteroperator/storage` | `oc get clustercsidriver`

## Documentation Structure

```text
ai-docs/
├── domain/
│   ├── storage-cr.md          # operator.openshift.io/v1 Storage (CSO config)
│   └── clustercsidriver.md    # operator.openshift.io/v1 ClusterCSIDriver (per-driver)
├── architecture/
│   ├── components.md          # Modes, CSIOperatorConfig, asset generation, image refs
│   └── pr-review.md           # PR review checklist and failure conditions
├── decisions/                 # CSO-specific ADRs
├── exec-plans/active/         # Active feature planning
├── references/
│   └── ecosystem.md           # Links to Tier 1
├── CSO_DEVELOPMENT.md         # Build, asset workflow, adding a driver
└── CSO_TESTING.md             # Unit and E2E test suites
```

**Platform Patterns (Tier 1)**: [Operator](https://github.com/openshift/enhancements/tree/master/ai-docs/platform/operator-patterns) | [Testing](https://github.com/openshift/enhancements/tree/master/ai-docs/practices/testing) | [Security](https://github.com/openshift/enhancements/tree/master/ai-docs/practices/security)

## Knowledge Graph

```text
                        [AGENTS.md] ← Start here
                              │
           ┌──────────────────┼──────────────────┐
           │                  │                  │
      [domain/]        [architecture/]      [decisions/]
   Storage CR          Modes + Config       Design history
   ClusterCSIDriver    PR review rules      Asset strategy
           │                  │                  │
           └──────────────────┼──────────────────┘
                              │
                    [references/ecosystem]
                      Links to Tier 1
```

**AI Agent Path**: domain/ → architecture/components.md → CSO_DEVELOPMENT.md → architecture/pr-review.md

## Drivers Managed

| Driver | Platform | HyperShift | Asset Type |
|--------|----------|-----------|------------|
| aws-ebs | AWS | ✅ | kustomize generated |
| azure-disk | Azure | ✅ | kustomize generated |
| azure-file | Azure | ✅ | kustomize generated |
| openstack-cinder | OpenStack | ✅ | kustomize generated |
| openstack-manila | OpenStack | ✅ | kustomize generated |
| gcp-pd | GCP | ❌ | static |
| ibm-vpc-block | IBM Cloud | ❌ | static |
| powervs-block | PowerVS | ✅ | static |
| vsphere | vSphere | ❌ | static |

---

**Tier 1 Hub**: https://github.com/openshift/enhancements/tree/master/ai-docs
