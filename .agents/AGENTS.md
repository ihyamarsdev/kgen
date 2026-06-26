# KGen Workspace Rules & Guidelines

Welcome to the **KGen** codebase. Below are the rules, guidelines, and context required for any autonomous agents developing or maintaining this project.

---

## 🛠 Tech Stack & Dependencies

- **Language**: Go 1.22
- **CLI Framework**: [spf13/cobra](https://github.com/spf13/cobra)
- **Terminal UI Framework**: [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)
- **CLI Styling**: [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss)
- **Diff Library**: [sergi/go-diff](https://github.com/sergi/go-diff) (LCS-based diff algorithm)
- **YAML Parsing**: [go-yaml/yaml.v3](https://github.com/go-yaml/yaml)

**Module path**: `github.com/ihyamarsdev/kgen`

---

## 📁 Project Structure

```
kgen/
├── cmd/
│   ├── charts.go      # Chart listing, path resolution, file scanning helpers
│   ├── common.go      # Shared helpers: confirm(), printErr(), findEditor(),
│   │                  # homeDir(), findHelm(), helmOutput(), helmRun(),
│   │                  # releaseNameFromChart(), helmReleaseExists()
│   ├── create.go      # kgen create — interactive wizard + chart generation
│   ├── deploy.go      # kgen deploy / undeploy / status — Helm CLI integration
│   ├── diff.go        # kgen diff — compare two chart directories
│   ├── edit.go        # kgen edit — interactive file selector + editor
│   ├── explain.go     # kgen explain — Kubernetes resource descriptions
│   ├── preview.go     # kgen preview — render chart templates in terminal
│   ├── root.go        # Cobra root command with --version flag
│   ├── uninstall.go   # kgen uninstall — remove binary, charts, config
│   ├── update.go      # kgen update — self-update from GitHub releases
│   └── validate.go    # kgen validate — best practices checker
├── internal/
│   ├── generator/
│   │   ├── generator.go  # Table-driven chart generation
│   │   └── templates.go  # Helm template string constants (~35 templates)
│   ├── tui/
│   │   ├── listmodel.go   # Reusable cursor-based list Bubble Tea model
│   │   ├── selector.go    # File picker (wraps ListModel)
│   │   ├── chartlist.go   # Chart folder picker (wraps ListModel)
│   │   ├── styles.go      # Lipgloss styling tokens
│   │   └── wizard.go      # 14-step Helm chart generation wizard
│   ├── validator/
│   │   └── validator.go  # Table-driven best practices validator
│   └── version/
│       └── version.go    # Version string (overridable via -ldflags)
├── .github/workflows/ci.yml      # GitHub Actions: test, vet, lint
├── .github/workflows/release.yml # GitHub Actions: auto-build binaries on release
├── .golangci.yml                 # golangci-lint configuration (v2)
├── CHANGELOG.md              # Keep a Changelog format
├── install.sh                # Auto-installer with checksum verification
├── main.go                   # Entrypoint
└── prd.md                    # Project requirements document
```

---

## 📌 Core Architecture & Guidelines

### 1. TUI Models (Bubble Tea)

- All Bubble Tea models **MUST** use pointer receivers (`*Model`) for `Init()`, `Update()`, and `View()`.
- Returning pointer receivers ensures `tea.Program.Run()` returns the correct concrete type, preventing type-assertion failures.
- The `ListModel` in `internal/tui/listmodel.go` is a **reusable, generic cursor-based list model**. Both `SelectorModel` (file picker) and `ChartListModel` (chart folder picker) are thin wrappers around it.
- When wrapping `ListModel`, delegate `Init()`, `Update()`, and `View()` to the embedded `*ListModel` and sync state back to the wrapper struct.

### 2. Generator & Templates

- Templates are stored as string constants in `internal/generator/templates.go`.
- **Table-driven generation**: `internal/generator/generator.go` uses a struct slice to conditionally write ~35 template files — do NOT add new `if cfg.GenerateX { os.WriteFile(...) }` blocks; add entries to the `templates` slice instead.
- Only `Chart.yaml` and `values.yaml` are evaluated as Go templates via `text/template` by KGen.
- The `quote` helper **MUST** be in the `FuncMap` inside `renderAndWrite()` to prevent execution errors.
- All `templates/*.yaml` files are written as-is (Helm template syntax, rendered by Helm at install time).

### 3. Chart Path Resolution

- Use the shared helpers in `cmd/charts.go`:
  - `listAvailableCharts()` — scans `~/kgen/` for valid Helm charts
  - `resolveChartPath(path)` — resolves absolute, relative, or chart-name paths
  - `scanAllChartFiles(dir)` — returns `map[relPath]content` for a chart directory
  - `isHidden(rel)` — excludes dotfiles and dot-directories
  - `promptChartChoice(charts)` — interactive number-based chart selection
- Default chart storage: `~/kgen/<app-name>` unless `-o` flag is used.

### 4. Shared Helpers (`cmd/common.go`)

| Helper | Purpose |
|--------|---------|
| `confirm(prompt)` | Interactive y/N prompt (returns false on non-terminal) |
| `printErr(format, args...)` | Formatted error to stderr |
| `findEditor()` | Returns `$EDITOR` or fallback (nano/vim/vi) |
| `homeDir()` | Returns home directory (tries user.Current, then os.UserHomeDir) |
| `findHelm()` | Returns helm binary path or empty string |
| `helmOutput(args...)` | Runs helm command, captures stdout+stderr |
| `helmRun(args...)` | Runs helm command with terminal I/O piping |
| `releaseNameFromChart(dir)` | Derives DNS-compatible release name from directory |
| `helmReleaseExists(release, ns)` | Checks if a release exists in a namespace |

### 5. File Operations

- Local precompiled binaries go in `dist/` (gitignored).
- Version is set in `internal/version/version.go` and overridden at build time via:
  ```bash
  go build -ldflags "-X github.com/ihyamarsdev/kgen/internal/version.Version=vX.Y.Z" -o kgen main.go
  ```

### 9. Release & Distribution

- Releases are **fully automated** via `.github/workflows/release.yml`.
- Triggered on `release: [published]` — just create a GitHub release with tag `vX.Y.Z`.
- Builds binaries for 3 platforms and uploads them + SHA256 checksums as release assets:
  - `kgen-linux-amd64` + `.sha256`
  - `kgen-darwin-amd64` + `.sha256`
  - `kgen-darwin-arm64` + `.sha256`
- Uses `setup-go@v5` with Go module caching.
- Uses `gh release upload --clobber` to attach assets.
- `linux/arm64` is excluded from the build matrix (no install.sh support yet).
- Installer script `install.sh` auto-detects OS/arch, downloads the correct binary, and verifies the SHA256 checksum.

### 6. Interactive File Selection & Editing

- When launching external terminal subprocesses (editors via `exec.Command`), ensure Bubble Tea terminal state is fully released/restored.
- The editor loop runs outside the Bubble Tea program loop to prevent tty conflicts.
- Editor errors should be reported to stderr (not silently swallowed):
  ```go
  if err := cmdEdit.Run(); err != nil {
      fmt.Fprintf(os.Stderr, "Warning: editor exited with error: %v\n", err)
  }
  ```

### 7. Helm Integration (deploy/undeploy/status)

- `kgen deploy` auto-detects install vs upgrade based on release existence.
- `kgen undeploy` requires confirmation (skippable with `-y`).
- All three commands require `helm` binary in PATH — use `requireHelm()` to check.
- Release name is auto-derived from chart directory name via `releaseNameFromChart()`.

### 8. Self-Update (`kgen update`)

- Downloads from GitHub releases, detects OS/arch automatically.
- Uses atomic binary replacement: temp file in same directory → `os.Rename()` (falls back to `copyFile` on EXDEV).
- Falls back to `sudo` when install directory is not writable.
- HTTP client has a 15-second timeout.

---

## 🧪 Testing & Verification

- Always verify changes by running the test suite:
  ```bash
  go test ./...
  ```
- Rebuild the binary to ensure no compilation errors:
  ```bash
  go build -o kgen main.go
  ```
- **Current test coverage**: cmd (11%), generator (65%), tui (10%), validator (83%).
- Test files: `cmd/common_test.go`, `cmd/e2e_test.go`, `internal/generator/generator_test.go`, `internal/tui/listmodel_test.go`, `internal/validator/validator_test.go`.
- New commands should have tests using `cobra.Command.ExecuteC()` for CLI-level testing.

---

## 🚫 Anti-Patterns (Things to Avoid)

1. **Never use `% 3` on a 2-element array** — the wizard had this bug in storage navigation.
2. **Never leave `if err == nil` after `if err != nil { return ... }`** — dead code (seen in scanAllChartFiles).
3. **Never swallow `os.File.Close()` errors** — use named returns with defer for `copyFile()`.
4. **Never use `http.DefaultClient` without timeout** — always use a client with `Timeout` set.
5. **Never duplicate `homeDir()`, `findEditor()`, or chart resolution logic** — use shared helpers.
6. **Never add repetitive `if` blocks for template writing** — use the table-driven approach.
7. **Never hardcode regex compilation inside a hot path** — hoist to package-level `var`.
8. **Never use `fmt.Println("Error: ...")` for errors** — use `printErr()` (goes to stderr).
9. **Never leave `_ = cmdEdit.Run()` without error handling** — report editor failures.
10. **Never use bubble sort** — use `sort.Strings()`.

---

## 📋 Command Reference

| Command | Description | Key Flags |
|---------|-------------|-----------|
| `kgen create` | Interactive chart generation wizard | `-p/--profile`, `-o/--output`, `-f/--force` |
| `kgen edit` | Interactive file selector + editor | `[chart-directory]` |
| `kgen validate` | Best practices validator | `-s/--strict` (exit 1 on warnings) |
| `kgen explain` | Kubernetes resource descriptions | — |
| `kgen diff` | Compare two chart directories | `[chart-a] [chart-b]` (exit 1 if different) |
| `kgen preview` | Render chart templates in terminal | `[chart-directory]` |
| `kgen deploy` | Install/upgrade chart to Kubernetes | `-r/--release`, `-n/--namespace`, `-f/--values`, `--set`, `-d/--dry-run`, `-t/--timeout`, `-w/--wait` |
| `kgen undeploy` | Uninstall release from Kubernetes | `-r/--release`, `-n/--namespace`, `-y/--yes`, `-d/--dry-run` |
| `kgen status` | Show release status | `-r/--release`, `-n/--namespace` |
| `kgen update` | Self-update from GitHub releases | `-y/--yes` |
| `kgen uninstall` | Remove kgen binary and data | `-y/--yes` |
| `kgen --version` | Print current version | `-V` |
