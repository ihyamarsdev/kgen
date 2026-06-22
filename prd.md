## Revisi PRD - Custom Mode

### Custom Mode Flow

```bash
kgen create --mode custom
```

Tampilan Bubble Tea:

```text
┌────────────────────────────────────┐
│ Select Resources                   │
├────────────────────────────────────┤
│ [✓] Deployment                     │
│ [✓] Service                        │
│ [✓] Ingress                        │
│ [ ] Gateway API                    │
│ [✓] ConfigMap                      │
│ [✓] ExternalSecret                 │
│ [✓] HPA                            │
│ [✓] ServiceMonitor                 │
│ [✓] PDB                            │
│ [ ] VPA                            │
│ [ ] KEDA                           │
│ [ ] StatefulSet                    │
│ [ ] CronJob                        │
│ [ ] ArgoCD Application             │
│ [ ] Istio VirtualService           │
└────────────────────────────────────┘

[ Space ] Toggle
[ Enter ] Continue
```

---

# Resource Catalog

## Workloads

### Deployment

```yaml
deployment.yaml
```

### StatefulSet

```yaml
statefulset.yaml
```

### DaemonSet

```yaml
daemonset.yaml
```

### Job

```yaml
job.yaml
```

### CronJob

```yaml
cronjob.yaml
```

---

## Networking

### Service

```yaml
service.yaml
```

### Ingress

```yaml
ingress.yaml
```

### Gateway API

```yaml
gateway.yaml
httproute.yaml
```

### NetworkPolicy

```yaml
networkpolicy.yaml
```

---

## Storage

### PersistentVolumeClaim

```yaml
pvc.yaml
```

### StorageClass Reference

```yaml
storageClassName:
```

### CSI Volume

```yaml
volumes:
```

---

## Configuration

### ConfigMap

```yaml
configmap.yaml
```

### Secret

```yaml
secret.yaml
```

### ExternalSecret

```yaml
externalsecret.yaml
```

### SealedSecret

```yaml
sealedsecret.yaml
```

---

## Autoscaling

### HorizontalPodAutoscaler

```yaml
hpa.yaml
```

### VerticalPodAutoscaler

```yaml
vpa.yaml
```

### KEDA

```yaml
scaledobject.yaml
triggerauthentication.yaml
```

---

## Security

### ServiceAccount

```yaml
serviceaccount.yaml
```

### Role

```yaml
role.yaml
```

### RoleBinding

```yaml
rolebinding.yaml
```

### ClusterRole

```yaml
clusterrole.yaml
```

### ClusterRoleBinding

```yaml
clusterrolebinding.yaml
```

### SecurityContext

Diinject ke Deployment:

```yaml
securityContext:
```

---

## Reliability

### PodDisruptionBudget

```yaml
pdb.yaml
```

### Pod Anti Affinity

```yaml
affinity:
```

### Topology Spread Constraints

```yaml
topologySpreadConstraints:
```

### PriorityClass

```yaml
priorityClassName:
```

---

## Monitoring

### ServiceMonitor

```yaml
servicemonitor.yaml
```

### PodMonitor

```yaml
podmonitor.yaml
```

### PrometheusRule

```yaml
prometheusrule.yaml
```

### GrafanaDashboard

```yaml
grafanadashboard.yaml
```

---

## Certificates

### Issuer

```yaml
issuer.yaml
```

### ClusterIssuer

```yaml
clusterissuer.yaml
```

### Certificate

```yaml
certificate.yaml
```

---

## GitOps

### Argo CD

Resource:

```yaml
application.yaml
applicationset.yaml
appproject.yaml
```

### Flux

Resource:

```yaml
gitrepository.yaml
helmrepository.yaml
helmrelease.yaml
kustomization.yaml
```

---

## Service Mesh

### Istio

```yaml
virtualservice.yaml
destinationrule.yaml
gateway.yaml
authorizationpolicy.yaml
requestauthentication.yaml
peerauthentication.yaml
```

---

# Smart Dependency Engine

KGen harus otomatis mendeteksi dependency.

Contoh:

Jika user memilih:

```text
✓ ServiceMonitor
```

maka otomatis:

```text
✓ Service
```

karena ServiceMonitor membutuhkan Service.

---

Jika memilih:

```text
✓ HPA
```

maka KGen menyarankan:

```text
✓ Resource Requests
```

karena HPA tanpa request CPU/Memory kurang optimal.

---

Jika memilih:

```text
✓ StatefulSet
```

maka KGen menawarkan:

```text
✓ PVC
```

---

Jika memilih:

```text
✓ ExternalSecret
```

maka KGen menanyakan:

```text
Secret Backend:
- Vault
- AWS Secrets Manager
- GCP Secret Manager
- Azure Key Vault
```

---

# Template Quality Level

Tambahkan pilihan kualitas template:

```text
Template Quality

( ) Basic
(*) Production
( ) Enterprise
```

### Basic

* Resource minimum

### Production

* Requests/Limits
* Probes
* HPA
* PDB

### Enterprise

* NetworkPolicy
* TopologySpreadConstraints
* PodSecurityContext
* ServiceMonitor
* AntiAffinity

---

# Estimasi Resource untuk v1.0

Target sekitar **40–50 template resource** yang mencakup:

* Core Kubernetes
* Monitoring
* Security
* Autoscaling
* GitOps
* Service Mesh
* Certificate Management

Dengan cakupan ini, KGen bukan hanya "Helm Generator", tetapi mulai mendekati **Platform Engineering Bootstrap Tool** yang bisa menjadi alternatif modern untuk `helm create` dengan pengalaman TUI yang jauh lebih baik menggunakan Bubble Tea.

