package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"kgen/internal/tui"

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
				var charts []string
				homeDir, err := os.UserHomeDir()
				if err == nil {
					kgenDir := filepath.Join(homeDir, "kgen")
					entries, err := os.ReadDir(kgenDir)
					if err == nil {
						for _, entry := range entries {
							if entry.IsDir() && isHelmChart(filepath.Join(kgenDir, entry.Name())) {
								charts = append(charts, entry.Name())
							}
						}
					}
				}

				switch len(charts) {
				case 0:
					fmt.Println("Error: No generated Helm charts found.")
					fmt.Println("Run 'kgen create' first to generate a chart, or specify a chart directory:")
					fmt.Println("  kgen edit [chart-directory]")
					os.Exit(1)

				case 1:
					// Only one chart — auto-select it
					homeDir, _ := os.UserHomeDir()
					targetDir = filepath.Join(homeDir, "kgen", charts[0])

				default:
					// Multiple charts — let user pick via interactive TUI
					homeDir, _ := os.UserHomeDir()
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
						return
					}
					targetDir = filepath.Join(homeDir, "kgen", resModel.SelectedChart)
				}
			}
		}

		// Verify targetDir is a Helm chart
		if !isHelmChart(targetDir) {
			fmt.Fprintf(os.Stderr, "Error: '%s' is not a valid Helm chart directory (missing Chart.yaml or values.yaml)\n", targetDir)
			os.Exit(1)
		}

		// Scan for all files recursively
		var files []string
		err := filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				rel, err := filepath.Rel(targetDir, path)
				if err == nil {
					// Exclude hidden files / dirs
					if !strings.HasPrefix(rel, ".") && !strings.Contains(rel, "/.") {
						files = append(files, rel)
					}
				}
			}
			return nil
		})

		if err != nil || len(files) == 0 {
			fmt.Fprintf(os.Stderr, "Error scanning files in '%s': %v\n", targetDir, err)
			os.Exit(1)
		}

		// Launcher loop
		editor := os.Getenv("EDITOR")
		if editor == "" {
			for _, e := range []string{"nano", "vim", "vi"} {
				if _, err := exec.LookPath(e); err == nil {
					editor = e
					break
				}
			}
		}

		if editor == "" {
			fmt.Println("Error: No terminal editor ($EDITOR, nano, vim, vi) found in path.")
			os.Exit(1)
		}

		for {
			selModel := tui.InitialSelectorModel(files)
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
