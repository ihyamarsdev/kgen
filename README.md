# KGen (Kubernetes & Helm Generator CLI)

**KGen** adalah CLI interaktif untuk menghasilkan Helm Chart production-ready berdasarkan kebutuhan pengguna, tanpa harus memahami seluruh resource Kubernetes dan Helm secara mendalam.

KGen membantu developer, DevOps Engineer, dan Platform Engineer membuat standar deployment Kubernetes yang konsisten melalui wizard interaktif berbasis terminal yang indah.

---

## Fitur Utama

- **TUI (Terminal UI) Interaktif**: Menggunakan [Bubble Tea](https://github.com/charmbracelet/bubbletea) untuk memberikan pengalaman CLI yang modern dan intuitif, lengkap dengan *Help Menu* interaktif dengan menekan tombol `?`.
- **Tiga Mode Deployment**:
  - **Simple Mode**: Hanya membuat `Deployment` dan `Service` (cocok untuk development/internal tools).
  - **Advanced Mode**: Membuat `Deployment`, `Service`, `Ingress`, dan `HPA` (cocok untuk production workloads).
  - **Custom Mode**: Memungkinkan pengguna memilih sendiri resource yang ingin di-generate.
- **Konfigurasi Berbasis Profil**:
  - `--profile dev`: Default untuk environment development (replicaCount=1, Ingress/HPA disabled).
  - `--profile prod`: Default untuk environment production (replicaCount=3, Ingress/HPA enabled dengan best practice resource limits dan security context).
- **Validation Engine**: Periksa apakah konfigurasi Helm Chart Anda memenuhi standar best practices Kubernetes (limits, requests, probes, security context).
- **Explain Command**: Memberikan penjelasan sederhana mengenai fungsi masing-masing resource Kubernetes yang dibuat.

---

## Cara Instalasi

Pastikan Anda memiliki Go (versi 1.22 ke atas) yang terinstal di sistem Anda.

1. Clone repositori ini.
2. Build binary kgen:
   ```bash
   go build -o kgen main.go
   ```
3. Pindahkan binary ke PATH Anda (opsional):
   ```bash
   mv kgen /usr/local/bin/
   ```

---

## Panduan Penggunaan

### 1. Membuat Helm Chart Baru (`kgen create`)

Untuk memulai wizard interaktif pembuatan Helm Chart:
```bash
./kgen create
```

Atau pilih profil tertentu saat pembuatan:
```bash
./kgen create --profile prod
```

Anda juga dapat menentukan folder output secara spesifik:
```bash
./kgen create -o ./my-helm-chart
```

### 2. Validasi Best Practices (`kgen validate`)

Untuk melakukan pemindaian standar keamanan dan keandalan pada Helm Chart:
```bash
./kgen validate ./my-helm-chart
```

### 3. Penjelasan Resource (`kgen explain`)

Untuk melihat ringkasan deskripsi fungsionalitas resource Kubernetes yang didukung:
```bash
./kgen explain
```

---

## Struktur Proyek

```text
kgen/
├── cmd/
│   ├── root.go        # Cobra root command
│   ├── create.go      # 'kgen create' command
│   ├── validate.go    # 'kgen validate' command
│   └── explain.go     # 'kgen explain' command
├── internal/
│   ├── generator/
│   │   ├── generator.go # Core generator logic
│   │   └── templates.go # Helm templates (deployment, service, ingress, hpa)
│   ├── tui/
│   │   ├── styles.go    # Lipgloss styling tokens
│   │   └── wizard.go    # Bubble Tea TUI implementation
│   └── validator/
│       └── validator.go # Best practices validator
├── main.go
└── prd.md
```
