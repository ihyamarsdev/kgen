# KGen (Kubernetes & Helm Generator CLI)

**KGen** is an interactive command-line interface (CLI) designed to generate production-ready Helm Charts based on user requirements—without requiring deep, initial expertise in all Kubernetes resources and Helm structures.

KGen helps developers, DevOps, and Platform Engineers establish consistent Kubernetes deployment standards through an interactive and aesthetic terminal-based wizard.

---

## Key Features

- **Interactive Terminal UI (TUI)**: Utilizes [Bubble Tea](https://github.com/charmbracelet/bubbletea) to deliver a modern, intuitive CLI experience, complete with an interactive *Help Menu* available by pressing `?`.
- **Three Deployment Modes**:
  - **Simple Mode**: Generates only `Deployment` and `Service` (perfect for development or internal tooling).
  - **Advanced Mode**: Generates `Deployment`, `Service`, `Ingress`, `HPA`, `PDB`, `ServiceMonitor`, and `NetworkPolicy` (tailored for production workloads).
  - **Custom Mode**: Empowers users to select individual resources from a checklist of categories (Workloads, Storage, RBAC, Networking, Scaling, Configuration, Monitoring, GitOps).
- **Profile-Based Configuration**:
  - `--profile dev`: Standard development profile (replicaCount=1, Ingress/HPA disabled).
  - `--profile prod`: Standard production profile (replicaCount=3, Ingress/HPA enabled, custom resource limits, and anti-affinity/topology constraints).
- **Smart Dependency Engine**: Automatically recommends or toggles dependencies when selecting resources in Custom Mode (e.g. `StatefulSet` prompts PVC creation, `ServiceMonitor` requires a `Service`, `RoleBinding` triggers `Role` and `ServiceAccount`).
- **Production Readiness Score**: Computes and displays a colored readiness score (out of 100) detailing compliance checks (probes, requests/limits, security policies) after generation.
- **Best Practices Validator**: Run `kgen validate` to inspect generated files against best practice guidelines.
- **Resource Explainer**: Run `kgen explain` to see functional descriptions of Kubernetes resources in clear, readable terms.
- **Chart Diff**: Run `kgen diff` to compare two generated Helm chart directories and see exactly what changed.
- **Template Preview**: Run `kgen preview` to render and display Helm chart templates in the terminal without writing to disk.
- **One-Click Deploy**: Run `kgen deploy` to install or upgrade a Helm chart directly to your Kubernetes cluster via Helm CLI.
- **Release Status**: Run `kgen status` to check the health and status of a deployed release.

---

## Installation

### 1. Automatic Installation (Bash Script)

You can download and install the precompiled binary for Linux/macOS automatically with the following command:

```bash
curl -sSfL https://raw.githubusercontent.com/ihyamarsdev/kgen/main/install.sh | bash
```

### 2. Manual Installation (Go CLI)

If you have Go installed (version 1.22 or above), you can install KGen directly from source:

```bash
go install github.com/ihyamarsdev/kgen@latest
```

### 3. Build from Source

1. Clone this repository:
   ```bash
   git clone https://github.com/ihyamarsdev/kgen.git
   cd kgen
   ```
2. Build the binary:
   ```bash
   go build -o kgen main.go
   ```
3. Move the binary to your execution path (optional):
   ```bash
   mv kgen /usr/local/bin/
   ```

---

## Usage Guide

### 1. Generate a New Helm Chart (`kgen create`)

To start the interactive terminal wizard:
```bash
kgen create
```
*Note: By default, generated charts are stored in your home directory at `~/kgen/<app-name>`.*

To run using the production profile:
```bash
kgen create --profile prod
```

To specify a custom output directory:
```bash
kgen create -o ./my-helm-chart
```

### 2. Validate Best Practices (`kgen validate`)

To run security and reliability checks on an existing Helm Chart:
```bash
kgen validate ./my-helm-chart
```

### 3. Resource Explanations (`kgen explain`)

To view descriptions and purposes of supported Kubernetes resources:
```bash
kgen explain
```

### 4. View or Edit Generated Charts (`kgen edit`)

To list and interactively view or edit generated files in a Helm chart:
```bash
kgen edit [chart-directory]
```
If `chart-directory` is omitted, KGen will check if the current directory is a valid Helm chart, or list the generated charts inside your default `~/kgen/` folder to choose from.

### 5. Update KGen (`kgen update`)

To check for the latest release and update kgen in-place:
```bash
kgen update
```

To skip the confirmation prompt (useful for automation):
```bash
kgen update -y
```

### 6. Uninstall KGen (`kgen uninstall`)

To remove the kgen binary, generated charts (`~/kgen/`), and configuration (`~/.config/kgen/`):
```bash
kgen uninstall
```

To skip the confirmation prompt:
```bash
kgen uninstall -y
```

### 7. Compare Charts (`kgen diff`)

To compare two generated Helm chart directories and see differences:
```bash
kgen diff [chart-a] [chart-b]
```

If paths are omitted, KGen will let you select from charts in `~/kgen/`.
Color-coded output shows removed lines (red `-`) and added lines (green `+`).

### 8. Preview Templates (`kgen preview`)

To render and display Helm chart templates in the terminal:
```bash
kgen preview [chart-directory]
```

If no directory is specified, KGen will auto-select or let you choose from `~/kgen/`.
Go-template files (Chart.yaml, values.yaml) are rendered with default values;
static template files (templates/*.yaml) are displayed as-is.

### 9. Deploy to Kubernetes (`kgen deploy`)

To deploy a generated Helm chart to your Kubernetes cluster:
```bash
kgen deploy [chart-directory]
```

If the release already exists, KGen will perform an upgrade. Otherwise, it performs a fresh install.

Specify a namespace, release name, or override values:
```bash
kgen deploy ./my-chart -n production -r my-app-release
kgen deploy ./my-chart --set image.tag=v2 --set replicaCount=5
kgen deploy ./my-chart -f overrides.yaml --wait --timeout 10m
```

Dry-run to preview what would be deployed:
```bash
kgen deploy ./my-chart --dry-run
```

If `chart-directory` is omitted, KGen will auto-select from `~/kgen/`.

**Prerequisite**: Helm CLI must be installed and available in your PATH.

### 10. Undeploy from Kubernetes (`kgen undeploy`)

To uninstall a Helm release from your cluster:
```bash
kgen undeploy [chart-directory]
```

Specify namespace and release name:
```bash
kgen undeploy ./my-chart -n production -r my-app-release
```

Skip the confirmation prompt:
```bash
kgen undeploy ./my-chart -y
```

Dry-run to preview what would be uninstalled:
```bash
kgen undeploy ./my-chart --dry-run
```

### 11. Check Release Status (`kgen status`)

To view the status of a deployed Helm release:
```bash
kgen status [chart-directory]
```

Specify namespace and release name:
```bash
kgen status ./my-chart -n production -r my-app-release
```

If `chart-directory` is omitted, KGen will auto-select from `~/kgen/`.

---

## Release Process

KGen uses **automated GitHub Actions** for releases. When a new release is published on GitHub:

1. **Trigger**: Publishing a release with tag `vX.Y.Z` triggers the [release workflow](.github/workflows/release.yml).
2. **Build**: Binaries are built for 3 platforms automatically:
   - `kgen-linux-amd64`
   - `kgen-darwin-amd64`
   - `kgen-darwin-arm64`
3. **Checksums**: SHA256 checksums are generated for each binary.
4. **Upload**: All binaries + checksums are uploaded to the release assets.

To create a new release:
1. Update `internal/version/version.go` with the new version.
2. Update `CHANGELOG.md` with release notes.
3. Commit and push.
4. Create a GitHub release with the matching tag (e.g., `v0.7.0`).
5. The workflow runs automatically — binaries appear as release assets within minutes.

Users install via:
```bash
curl -sSfL https://raw.githubusercontent.com/ihyamarsdev/kgen/main/install.sh | bash
```
The installer detects OS/arch, downloads the correct binary, and verifies its checksum.

---

## Project Structure

```text
kgen/
│   ├── charts.go      # Shared chart listing, selection, and file scanning helpers
│   ├── common.go      # Shared helpers (confirmation, error printing, helm utils)
│   ├── deploy.go      # 'kgen deploy', 'kgen undeploy', 'kgen status' commands
│   ├── diff.go        # 'kgen diff' command
│   ├── preview.go     # 'kgen preview' command
│   ├── root.go        # Cobra root command (includes --version flag)
│   ├── create.go      # 'kgen create' command
│   ├── edit.go        # 'kgen edit' command
│   ├── uninstall.go   # 'kgen uninstall' command
│   ├── update.go      # 'kgen update' command
│   ├── validate.go    # 'kgen validate' command
│   └── explain.go     # 'kgen explain' command
├── internal/
│   ├── generator/
│   │   ├── generator.go # Core generator logic
│   │   └── templates.go # Helm templates catalog
│   ├── tui/
│   │   ├── styles.go    # Lipgloss styling tokens
│   │   ├── listmodel.go # Reusable cursor-based list Bubble Tea model
│   │   ├── selector.go  # File selector/editor Bubble Tea TUI
│   │   ├── chartlist.go # Chart folder selection Bubble Tea TUI
│   │   └── wizard.go    # Bubble Tea TUI implementation
│   ├── validator/
│   │   └── validator.go # Best practices validator
│   └── version/
│       └── version.go   # Version, repo metadata, and build-time constants
├── dist/              # Distribution directory for precompiled binaries
├── install.sh         # Installer shell script
├── main.go            # Entrypoint
└── prd.md             # Project requirements and specification
```

