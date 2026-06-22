# KGen Workspace Rules & Guidelines

Welcome to the **KGen** codebase. Below are the rules, guidelines, and context required for any autonomous agents developing or maintaining this project.

---

## 🛠 Tech Stack & Dependencies

- **Language**: Go 1.22+
- **CLI Framework**: [spf13/cobra](https://github.com/spf13/cobra)
- **Terminal UI Framework**: [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)
- **CLI Styling**: [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss)

---

## 📌 Core Architecture & Guidelines

### 1. Interactive Terminal UI (TUI) Models
- All Bubble Tea models (e.g., `WizardModel` in `wizard.go`, `SelectorModel` in `selector.go`) **MUST** use pointer receivers (`*Model`) for `Init()`, `Update()`, and `View()`.
- Returning pointer receivers ensures that `tea.Program.Run()` returns the correct concrete pointer type, preventing type-assertion failures (e.g., `mRes.(*SelectorModel)`) when capturing final UI states.

### 2. Generator & Templates
- Standard Helm templates are stored as string constants in `internal/generator/templates.go`.
- Only `Chart.yaml` and `values.yaml` are evaluated as Go templates via `text/template` by KGen.
- Any Go template helpers (like `quote`) **MUST** be explicitly mapped in the `FuncMap` inside the generator parser (`internal/generator/generator.go`) to prevent parsing/execution errors.

### 3. File Operations
- By default, generated charts are saved under `~/kgen/<app-name>` if the output directory flag `-o` is omitted.
- Local precompiled binaries should be placed in `dist/` and are ignored by git via `.gitignore`.

### 4. Interactive File Selection & Editing
- When adding or modifying interactive file selections, always ensure terminal states are fully released/restored by Bubble Tea before launching external terminal subprocesses (like editors via `exec.Command`).
- The editor selection loop should run outside the Bubble Tea program loop to prevent terminal raw mode/tty conflicts.

---

## 🧪 Testing & Verification

- Always verify changes by running the test suite:
  ```bash
  go test ./...
  ```
- Rebuild the binary to ensure there are no compilation errors:
  ```bash
  go build -o kgen main.go
  ```
