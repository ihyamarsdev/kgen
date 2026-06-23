package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ihyamarsdev/kgen/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit [chart-directory]",
	Short: "Interactive selector to view or edit generated Helm chart files",
	Long:  `Scan a Helm chart directory and launch an interactive menu to view or edit its files in your terminal editor.`,
	Run: func(cmd *cobra.Command, args []string) {
		var targetDir string

		if len(args) > 0 {
			targetDir = args[0]
		} else {
			// Check current directory first
			if isHelmChart(".") {
				targetDir = "."
			} else {
				// Scan ~/kgen/ for chart directories
				charts := listAvailableCharts()
				switch len(charts) {
				case 0:
					fmt.Println("Error: No generated Helm charts found.")
					fmt.Println("Run 'kgen create' first to generate a chart, or specify a chart directory:")
					fmt.Println("  kgen edit [chart-directory]")
					os.Exit(1)

				case 1:
					// Only one chart — auto-select it
					targetDir = filepath.Join(homeDir(), "kgen", charts[0])

				default:
					// Multiple charts — let user pick via interactive TUI
					listModel := tui.InitialChartListModel(charts)
					listProg := tea.NewProgram(&listModel)
					mRes, err := listProg.Run()
					if err != nil {
						printErr("Error running chart selector: %v", err)
						os.Exit(1)
					}
					resModel, ok := mRes.(*tui.ChartListModel)
					if !ok || resModel.Quitted || resModel.SelectedChart == "" {
						// User cancelled
						fmt.Println("Edit cancelled.")
						return
					}
					targetDir = filepath.Join(homeDir(), "kgen", resModel.SelectedChart)
				}
			}
		}

		// Verify targetDir is a Helm chart
		if !isHelmChart(targetDir) {
			fmt.Fprintf(os.Stderr, "Error: '%s' is not a valid Helm chart directory (missing Chart.yaml or values.yaml)\n", targetDir)
			os.Exit(1)
		}

		// Scan for all files recursively
		files, err := scanAllChartFiles(targetDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning files in '%s': %v\n", targetDir, err)
			os.Exit(1)
		}
		if len(files) == 0 {
			fmt.Fprintf(os.Stderr, "Error: No editable files found in '%s'\n", targetDir)
			os.Exit(1)
		}

		// Build sorted file list
		var fileList []string
		for f := range files {
			if !isHidden(f) {
				fileList = append(fileList, f)
			}
		}
		if len(fileList) == 0 {
			fmt.Fprintf(os.Stderr, "Error: No editable files found in '%s'\n", targetDir)
			os.Exit(1)
		}

		// Launcher loop
		editor := findEditor()
		if editor == "" {
			printErr("Error: No terminal editor ($EDITOR, nano, vim, vi) found in path.")
			os.Exit(1)
		}

		for {
			selModel := tui.InitialSelectorModel(fileList)
			selProg := tea.NewProgram(&selModel)
			mRes, err := selProg.Run()
			if err != nil {
				break
			}
			resModel, ok := mRes.(*tui.SelectorModel)
			if !ok || resModel.Quitted || resModel.SelectedFile == "" {
				break
			}

			// Launch Editor
			filePath := filepath.Join(targetDir, resModel.SelectedFile)
			cmdEdit := exec.Command(editor, filePath)
			cmdEdit.Stdin = os.Stdin
			cmdEdit.Stdout = os.Stdout
			cmdEdit.Stderr = os.Stderr
			_ = cmdEdit.Run()
		}
	},
}

func isHelmChart(dir string) bool {
	chartYaml := filepath.Join(dir, "Chart.yaml")
	valuesYaml := filepath.Join(dir, "values.yaml")
	if _, err := os.Stat(chartYaml); err == nil {
		if _, err := os.Stat(valuesYaml); err == nil {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(editCmd)
}
