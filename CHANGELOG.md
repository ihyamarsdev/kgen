# Changelog

All notable changes to KGen will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.1.9] - 2026-06-23

### Added
- **Input validation**: Port (1-65535), image tag (auto-append `:latest`), storage size (pattern validation)

## [v0.1.8] - 2026-06-23

### Added
- **13 unit tests** across cmd, generator, and validator packages
- **Atomic binary replacement** in `kgen update` via `os.Rename()`
- **`isHidden()` fix**: Correctly handle top-level dotfiles, hidden dirs, and non-dot filenames

### Changed
- **Extract ListModel**: Reusable cursor-based list model eliminating 90% duplicated code between SelectorModel and ChartListModel

## [v0.1.7] - 2026-06-23

### Added
- **`kgen create --force`**: Overwrite existing output directories
- **`kgen validate --strict`**: Exit code 1 on warnings for CI/CD quality gates
- **Styled help templates** for all subcommands (not just root)
- **SHA256 checksum verification** in `install.sh`
- **Auto-add `~/.local/bin` to PATH** when installed locally without sudo
- **35 English resource descriptions** in `kgen explain` (was 9 Indonesian)

### Changed
- **Table-driven generator**: Replaced ~200 lines of repetitive `if` blocks
- **Table-driven validator**: Replaced 5 repetitive PASS/WARN blocks
- **Flux API update**: `v2beta1→v2`, `v1beta2→v1`

### Fixed
- Replace O(n²) bubble sort with `sort.Strings()` in preview
- Tree character: use `└──` for last item instead of `├──`
- Missing DaemonSet/Job in tree output
- `promptChartChoice` error message now shows valid range
- Dead code removal (`_ = successStyle`)

## [v0.1.6] - 2026-06-23

### Added
- **`kgen diff`**: Compare two Helm chart directories with color-coded unified diff
- **`kgen preview`**: Render and display Helm chart templates in terminal
- **`--strict` flag** for `kgen validate` (CI/CD quality gate)

### Changed
- **Module path**: `kgen` → `github.com/ihyamarsdev/kgen` (enables `go install`)
- Extract `findEditor()` and `homeDir()` to shared `common.go`
- Refactor `kgen edit` to use shared chart helpers

### Fixed
- Diff algorithm: Replace naive line-by-line with proper LCS-based `go-diff`
- HTTP timeout (15s) for all GitHub API and download calls
- `isHidden()` false positive on files with dots in filename
- Guard `kgen diff` from comparing same chart against itself

## [v0.1.5] - 2026-06-23

### Added
- **`kgen diff`**: Compare two Helm chart directories
- **`kgen preview`**: Render and display chart templates in terminal

## [v0.1.4] - 2026-06-23

### Changed
- Replace text-based chart picker in `kgen edit` with interactive TUI list (ChartListModel)

## [v0.1.3] - 2026-06-23

### Added
- **`kgen update`**: Check for latest release and replace binary in-place
- **`kgen uninstall`**: Remove binary, generated charts, and configuration
- **`kgen --version` / `-V`**: Print current version
- `internal/version` package with build-time overrideable version

## [v0.1.2] - 2026-06-22

### Fixed
- Fixes type assertion and pointer receiver in File Selector TUI (edit menu was exiting immediately)

## [v0.1.1] - 2026-06-22

### Added
- Interactive file selector and editor launcher post-generation
- `kgen edit` command

## [v0.1.0] - 2026-06-22

### Added
- Revamped Custom Mode with category-based TUI navigation
- Smart dependency engine
- Production Readiness Score
- AGENTS.md workspace rules
- English README with install.sh

## [v0.0.1] - 2026-06-22

### Added
- Initial KGen CLI with interactive TUI wizard
- Simple, Advanced, and Custom deployment modes
- Profile-based configuration (`--profile dev/prod`)
- Helm chart generation with 30+ Kubernetes resource templates
- Best practices validator (`kgen validate`)
- Resource explainer (`kgen explain`)

[Unreleased]: https://github.com/ihyamarsdev/kgen/compare/v0.1.9...HEAD
[v0.1.9]: https://github.com/ihyamarsdev/kgen/compare/v0.1.8...v0.1.9
[v0.1.8]: https://github.com/ihyamarsdev/kgen/compare/v0.1.7...v0.1.8
[v0.1.7]: https://github.com/ihyamarsdev/kgen/compare/v0.1.6...v0.1.7
[v0.1.6]: https://github.com/ihyamarsdev/kgen/compare/v0.1.5...v0.1.6
[v0.1.5]: https://github.com/ihyamarsdev/kgen/compare/v0.1.4...v0.1.5
[v0.1.4]: https://github.com/ihyamarsdev/kgen/compare/v0.1.3...v0.1.4
[v0.1.3]: https://github.com/ihyamarsdev/kgen/compare/v0.1.2...v0.1.3
[v0.1.2]: https://github.com/ihyamarsdev/kgen/compare/v0.1.1...v0.1.2
[v0.1.1]: https://github.com/ihyamarsdev/kgen/compare/v0.1.0...v0.1.1
[v0.1.0]: https://github.com/ihyamarsdev/kgen/compare/v0.0.1...v0.1.0
[v0.0.1]: https://github.com/ihyamarsdev/kgen/releases/tag/v0.0.1
