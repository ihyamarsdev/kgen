# KGen P1 Feature Enhancement Plan

## Status Overview

| Feature | Status | Notes |
|---------|--------|-------|
| Category-Based Selection | ✅ **Already Implemented** | Wizard has full category/resource navigation |
| Smart Dependency Engine | 🔄 **Enhanced** | Added 8 new dependency rules |
| Production Readiness Score | 🔄 **Enhanced** | Expanded validator + recommendations |

---

## Phase 1: Enhanced Smart Dependency Engine

### Current Dependencies (Already Working)
- [x] ServiceMonitor → auto-enable Service
- [x] Deselect Service → deselect ServiceMonitor
- [x] StatefulSet → auto-enable PVC
- [x] Deselect StatefulSet → deselect PVC
- [x] RoleBinding → auto-enable Role + ServiceAccount
- [x] ClusterRoleBinding → auto-enable ClusterRole + ServiceAccount

### New Dependencies To Add
- [x] Deployment → auto-enable Service (Deployment without Service is unusual)
- [x] Deselect Service → deselect Ingress, Gateway, ServiceMonitor
- [x] Ingress → auto-enable Service (Ingress requires a Service)
- [x] Gateway → auto-enable Service
- [x] CronJob → auto-enable ConfigMap
- [x] Job → auto-enable ConfigMap
- [x] HPA → suggest upgrading TemplateQuality if basic quality
- [x] VPA + HPA selected → warn (conflicting autoscalers)

### Dependency Hints Display
- [x] Show dependency hints in category resource view when toggling items

---

## Phase 2: Expanded Validator

### Current Checks (5)
- [x] Resource Limits
- [x] Resource Requests
- [x] Liveness Probe
- [x] Readiness Probe
- [x] Security Context

### New Checks To Add (5)
- [x] HPA (check values.yaml for autoscaling.enabled)
- [x] PDB (check for pdb.yaml existence or values.yaml pdb section)
- [x] NetworkPolicy (check for networkpolicy.yaml existence)
- [x] Topology Spread Constraints (check values.yaml for topologySpreadConstraints)
- [x] Pod Anti Affinity (check values.yaml for affinity.podAntiAffinity)

---

## Phase 3: Enhanced Score Display

- [x] Show actionable recommendations for items scoring 0 (what to add next)
- [x] Display improvement path suggestions after the score
- [x] Color-coded score bar

---

## Implementation Order

1. Phase 1: Enhanced Smart Dependency Engine → `internal/tui/wizard.go`
2. Phase 2: Expanded Validator → `internal/validator/validator.go`
3. Phase 3: Enhanced Score Display → `cmd/create.go`
4. Update Tests → `internal/validator/validator_test.go`, `internal/tui/wizard_test.go`
5. Run test suite and fix issues
