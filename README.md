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

---

## Project Structure

```text
kgen/
├── cmd/
│   ├── root.go        # Cobra root command
│   ├── create.go      # 'kgen create' command
│   ├── edit.go        # 'kgen edit' command
│   ├── validate.go    # 'kgen validate' command
│   └── explain.go     # 'kgen explain' command
├── internal/
│   ├── generator/
│   │   ├── generator.go # Core generator logic
│   │   └── templates.go # Helm templates catalog
│   ├── tui/
│   │   ├── styles.go    # Lipgloss styling tokens
│   │   ├── selector.go  # File selector/editor Bubble Tea TUI
│   │   └── wizard.go    # Bubble Tea TUI implementation
│   └── validator/
│       └── validator.go # Best practices validator
├── dist/              # Distribution directory for precompiled binaries
├── install.sh         # Installer shell script
├── main.go            # Entrypoint
└── prd.md             # Project requirements and specification
```

