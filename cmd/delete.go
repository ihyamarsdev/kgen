package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ihyamarsdev/kgen/internal/tui"
	"github.com/spf13/cobra"
)

var deleteYes bool

var deleteCmd = &cobra.Command{
	Use:   "delete [chart-directory]",
	Short: "Delete a generated Helm chart from disk",
	Long: `Delete a generated Helm chart directory and all its files from ~/kgen/.

This removes the local chart files only — it does NOT uninstall
the Helm release from a Kubernetes cluster (use 'kgen undeploy' for that).

If no chart directory is provided, kgen will auto-select from ~/kgen/.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runDelete(args)
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteYes, "yes", "y", false, "Skip confirmation prompt")
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(args []string) {
	chartDir := resolveDeleteChart(args)

	// List files that will be deleted.
	files, err := os.ReadDir(chartDir)
	if err != nil {
		printErr("Error reading chart directory: %v", err)
		os.Exit(1)
	}

	fmt.Println(tui.HeaderStyle.Render("Deleting Helm Chart"))
	fmt.Printf("  Chart: %s\n\n", chartDir)

	totalFiles := 0
	for _, f := range files {
		if !f.IsDir() && f.Name()[0] != '.' {
			fmt.Printf("  %s\n", f.Name())
			totalFiles++
		}
	}
	tplDir := filepath.Join(chartDir, "templates")
	if tplInfo, err := os.Stat(tplDir); err == nil && tplInfo.IsDir() {
		templates, _ := os.ReadDir(tplDir)
		for _, t := range templates {
			if !t.IsDir() && t.Name()[0] != '.' {
				fmt.Printf("  templates/%s\n", t.Name())
				totalFiles++
			}
		}
	}
	fmt.Printf("\n  Total: %d files\n\n", totalFiles)

	if !deleteYes {
		if !confirm(fmt.Sprintf("Delete chart '%s' and all its files?", filepath.Base(chartDir))) {
			fmt.Println("Delete cancelled.")
			return
		}
	}

	if err := os.RemoveAll(chartDir); err != nil {
		printErr("Error deleting chart: %v", err)
		os.Exit(1)
	}

	fmt.Println(tui.SuccessStyle.Render(fmt.Sprintf("Chart '%s' deleted successfully.", filepath.Base(chartDir))))
}

// resolveDeleteChart resolves the chart directory for delete command.
// Same as resolveDeployChart but separate to keep delete independent.
func resolveDeleteChart(args []string) string {
	if len(args) > 0 {
		return resolveChartPath(args[0])
	}

	charts := listAvailableCharts()
	if len(charts) == 0 {
		printErr("Error: No generated Helm charts found in ~/kgen/.")
		fmt.Println("Nothing to delete.")
		os.Exit(1)
	}

	if len(charts) == 1 {
		return filepath.Join(homeDir(), "kgen", charts[0])
	}

	fmt.Println(tui.HeaderStyle.Render("Select a chart to delete:"))
	for i, c := range charts {
		fmt.Printf("  [%d] %s\n", i+1, c)
	}
	return promptChartChoice(charts)
}
