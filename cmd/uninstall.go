package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ihyamarsdev/kgen/internal/tui"

	"github.com/spf13/cobra"
)

// uninstallYesFlag (-y / --yes) skips the interactive confirmation prompt.
var uninstallYesFlag bool

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall kgen and remove generated data",
	Long: `Remove the kgen binary from your PATH and optionally delete any
generated charts (~/kgen/) and configuration files (~/.config/kgen/).
A confirmation prompt is shown before any deletions; use the --yes flag
to skip it (useful for automation and scripts).`,
	Run: func(cmd *cobra.Command, args []string) {
		runUninstall()
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
	uninstallCmd.Flags().BoolVarP(&uninstallYesFlag, "yes", "y", false, "skip the confirmation prompt")
}

func runUninstall() {
	var items []uninstallItem

	// 1. Locate the kgen binary to remove.
	if binPath := findKgenBinary(); binPath != "" {
		items = append(items, uninstallItem{
			label: "Binary",
			path:  binPath,
			kind:  itemBinary,
		})
	}

	home := homeDir()

	// 2. Generated charts directory (~/kgen).
	if home != "" {
		kgenDir := filepath.Join(home, "kgen")
		if info, err := os.Stat(kgenDir); err == nil && info.IsDir() {
			items = append(items, uninstallItem{
				label: "Generated charts directory",
				path:  kgenDir,
				kind:  itemCharts,
			})
		}
	}

	// 3. Configuration directory (~/.config/kgen).
	if home != "" {
		cfgDir := filepath.Join(home, ".config", "kgen")
		if info, err := os.Stat(cfgDir); err == nil && info.IsDir() {
			items = append(items, uninstallItem{
				label: "Configuration directory",
				path:  cfgDir,
				kind:  itemConfig,
			})
		}
	}

	if len(items) == 0 {
		fmt.Println(tui.GrayStyle.Render("Nothing to uninstall. kgen does not appear to be installed."))
		return
	}

	// Display summary.
	fmt.Println(tui.HeaderStyle.Render("KGen Uninstall"))
	fmt.Println("The following items will be removed:")
	fmt.Println()
	for _, item := range items {
		fmt.Printf("  • %s (%s)\n", item.label, item.path)
	}
	fmt.Println()

	if !uninstallYesFlag {
		if !confirm("Do you want to proceed?") {
			fmt.Println("Uninstall cancelled.")
			return
		}
	}

	// Perform removals.
	var failed bool
	for _, item := range items {
		if err := removeItem(item); err != nil {
			printErr("  ✗ Failed to remove %s: %v", item.label, err)
			failed = true
		} else {
			fmt.Printf("  ✓ Removed %s\n", tui.GrayStyle.Render(item.label))
		}
	}

	fmt.Println()
	if failed {
		fmt.Println(tui.ErrorStyle.Render("Uninstall completed with errors."))
		fmt.Println("Some items could not be removed — check the messages above.")
	} else {
		fmt.Println(tui.SuccessStyle.Render("KGen has been uninstalled successfully."))
		fmt.Println("Thank you for using KGen!")
	}
}

// uninstallItem represents a single file or directory to be removed.
type uninstallItem struct {
	label string
	path  string
	kind  itemKind
}

type itemKind int

const (
	itemBinary itemKind = iota
	itemCharts
	itemConfig
)

// findKgenBinary returns the absolute path of the kgen binary that would be
// executed from the current PATH. It returns an empty string when kgen is not
// found (e.g. the binary was built locally and never installed system-wide).
func findKgenBinary() string {
	path, err := exec.LookPath("kgen")
	if err != nil {
		return ""
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	// Resolve symlinks so we report the real location.
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return resolved
	}
	return abs
}

// removeItem removes the uninstall item. Binaries may need sudo.
func removeItem(item uninstallItem) error {
	if item.kind == itemBinary {
		return removeBinary(item.path)
	}
	return os.RemoveAll(item.path)
}

// removeBinary tries to remove a binary, falling back to sudo when necessary.
func removeBinary(path string) error {
	if err := os.Remove(path); err == nil {
		return nil
	}

	// Possibly a permission issue — try sudo.
	if _, err := exec.LookPath("sudo"); err != nil {
		return fmt.Errorf("sudo required to remove %s but was not found", path)
	}
	cmd := exec.Command("sudo", "rm", "-f", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
