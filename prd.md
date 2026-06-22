# PRD — KGen (Kubernetes & Helm Generator CLI)

## 1. Ringkasan

**KGen** adalah CLI interaktif untuk menghasilkan Helm Chart production-ready berdasarkan kebutuhan pengguna, tanpa harus memahami seluruh resource Kubernetes dan Helm.

KGen membantu developer, DevOps Engineer, dan Platform Engineer membuat standar deployment Kubernetes yang konsisten melalui wizard interaktif berbasis terminal.

Berbeda dengan generator Helm biasa, KGen menyediakan:

* Mode Simple, Advanced, dan Custom
* Helm Chart production-ready
* Best practice Kubernetes bawaan
* Import Kubernetes Manifest → Helm Chart
* GitOps-ready structure
* TUI (Terminal UI) menggunakan Bubble Tea

---

# 2. Problem Statement

Saat ini proses membuat Helm Chart sering mengalami masalah:

### Developer

* Tidak memahami Helm templating
* Copy-paste chart lama
* Banyak konfigurasi tidak terpakai

### DevOps Engineer

* Chart tidak konsisten antar proyek
* Naming convention berbeda-beda
* Resource penting sering terlupakan

Contoh:

* HPA tidak dibuat
* PDB tidak dibuat
* NetworkPolicy tidak dibuat
* Resource request/limit tidak diisi

### Platform Team

* Sulit menjaga standar deployment
* Review chart memakan waktu

---

# 3. Goals

### Primary Goals

* Mempercepat pembuatan Helm Chart
* Menstandarkan deployment Kubernetes
* Mengurangi copy-paste chart lama

### Success Metrics

* Generate chart < 30 detik
* 80% resource standar tersedia otomatis
* Chart lolos Helm lint
* Chart bisa langsung dipakai di cluster

---

# 4. Target User

## Persona 1 — Developer

Pengalaman Kubernetes minim.

Kebutuhan:

* Deploy aplikasi cepat
* Tidak ingin belajar Helm secara mendalam

---

## Persona 2 — DevOps Engineer

Mengelola banyak service.

Kebutuhan:

* Konsistensi chart
* Best practice otomatis

---

## Persona 3 — Platform Engineer

Membangun Internal Developer Platform.

Kebutuhan:

* Template reusable
* Standarisasi deployment

---

# 5. Technology Stack

## Language

Go

## UI Framework

[Bubble Tea](https://github.com/charmbracelet/bubbletea?utm_source=chatgpt.com)

## Additional Libraries

* [Bubbles](https://github.com/charmbracelet/bubbles?utm_source=chatgpt.com)
* [Lip Gloss](https://github.com/charmbracelet/lipgloss?utm_source=chatgpt.com)
* [Helm SDK](https://helm.sh/docs/sdk/?utm_source=chatgpt.com)
* [Sprig](https://github.com/Masterminds/sprig?utm_source=chatgpt.com)
* YAML v3

---

# 6. Core Features

## Feature 1 — Interactive Chart Generator

Command:

```bash
kgen create
```

User akan masuk ke wizard interaktif.

---

### Step 1

Application Information

```text
Application Name
Namespace
Container Image
Container Port
```

---

### Step 2

Deployment Mode

```text
Simple
Advanced
Custom
```

---

# 7. Deployment Modes

## Simple

Resource:

```text
✓ Deployment
✓ Service
```

Tidak termasuk:

```text
✗ Ingress
✗ HPA
✗ PDB
✗ ServiceMonitor
✗ NetworkPolicy
```

Target:

* Internal tools
* Development environment

---

## Advanced

Resource:

```text
✓ Deployment
✓ Service
✓ Ingress
✓ HPA
✓ PDB
✓ ServiceMonitor
✓ NetworkPolicy
```

Target:

* Production workloads

---

## Custom

User memilih resource satu per satu.

Contoh:

```text
[✓] Deployment
[✓] Service
[✓] Ingress
[ ] HPA
[✓] PDB
[ ] ServiceMonitor
[✓] NetworkPolicy
[ ] CronJob
[ ] StatefulSet
```

---

# 8. Supported Kubernetes Resources

## Workload

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

### NetworkPolicy

```yaml
networkpolicy.yaml
```

---

## Scaling

### HorizontalPodAutoscaler

```yaml
hpa.yaml
```

---

## Reliability

### PodDisruptionBudget

```yaml
pdb.yaml
```

### Pod Anti Affinity

Deployment patch:

```yaml
affinity:
```

---

## Monitoring

### ServiceMonitor

```yaml
servicemonitor.yaml
```

### PrometheusRule

```yaml
prometheusrule.yaml
```

---

## Secrets

### Secret

```yaml
secret.yaml
```

### ExternalSecret

```yaml
externalsecret.yaml
```

---

# 9. Helm Template Standards

Semua resource menggunakan conditional rendering.

Contoh:

```yaml
{{- if .Values.ingress.enabled }}
```

```yaml
{{- if .Values.hpa.enabled }}
```

```yaml
{{- if .Values.networkPolicy.enabled }}
```

---

# 10. Profiles

## Development

```bash
kgen create --profile dev
```

Default:

```yaml
replicaCount: 1
ingress:
  enabled: false

hpa:
  enabled: false
```

---

## Production

```bash
kgen create --profile prod
```

Default:

```yaml
replicaCount: 3

ingress:
  enabled: true

hpa:
  enabled: true

pdb:
  enabled: true
```

---

## Enterprise

```bash
kgen create --profile enterprise
```

Tambahan:

```yaml
networkPolicy:
  enabled: true

serviceMonitor:
  enabled: true

externalSecret:
  enabled: true
```

---

# 11. Import Existing Manifest

Command:

```bash
kgen import manifests/
```

Input:

```text
deployment.yaml
service.yaml
ingress.yaml
```

Output:

```text
chart/
├── Chart.yaml
├── values.yaml
└── templates/
```

Fungsi:

* Parse YAML
* Identifikasi field hardcoded
* Konversi menjadi Helm values

Contoh:

```yaml
image: nginx:latest
```

menjadi:

```yaml
image:
  repository: nginx
  tag: latest
```

dan template:

```yaml
image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
```

---

# 12. Validation Engine

Command:

```bash
kgen validate
```

Pemeriksaan:

### Resource Requests

```yaml
resources:
```

### Resource Limits

```yaml
limits:
```

### Liveness Probe

```yaml
livenessProbe:
```

### Readiness Probe

```yaml
readinessProbe:
```

### Security Context

```yaml
securityContext:
```

---

Output:

```text
WARN: No liveness probe found
WARN: No resource limit configured
PASS: Readiness probe found
```

---

# 13. Explain Generated Resources

Command:

```bash
kgen explain
```

Output:

```text
Deployment
├─ Menjalankan aplikasi

Service
├─ Mengekspos pod

Ingress
├─ Mengekspos aplikasi ke luar cluster

HPA
├─ Auto scaling pod
```

---

# 14. GitOps Generator

Command:

```bash
kgen gitops init
```

Output:

```text
clusters/
apps/
platform/
```

Support:

* Argo CD
* Flux

---

# 15. Future Features (v2)

## AI Generator

```bash
kgen ai
```

Prompt:

```text
Generate Laravel application with Redis and MySQL
```

Output:

```text
helm-chart/
```

---

## Kubernetes Best Practice Scanner

```bash
kgen scan chart/
```

Deteksi:

* Missing probes
* Missing limits
* Missing HPA
* Missing NetworkPolicy

---

# 16. CLI Architecture

```text
cmd/
internal/
├── tui/
├── templates/
├── generator/
├── importer/
├── validator/
├── profiles/
└── gitops/

charts/
```

---

# 17. MVP Scope (v0.1)

### Included

* Bubble Tea TUI
* Simple Mode
* Advanced Mode
* Custom Mode
* Deployment
* Service
* Ingress
* HPA
* values.yaml generation
* Helm lint validation

### Excluded

* AI
* Manifest Import
* GitOps Generator
* Enterprise Profiles

Target MVP: menghasilkan Helm Chart yang siap digunakan dalam waktu kurang dari 1 menit melalui wizard terminal yang nyaman dan konsisten.

