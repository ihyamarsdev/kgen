package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"kgen/internal/tui"

	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff [chart-a] [chart-b]",
	Short: "Compare two Helm chart directories",
	Long: `Compare two generated Helm chart directories and display a unified diff
of their file contents. Hidden files (dotfiles) and empty directories
are skipped.

If either path is omitted, kgen will attempt to auto-select from ~/kgen/.`,
	Args: cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		runDiff(args)
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
}

func runDiff(args []string) {
	dirA, dirB := resolveDiffPaths(args)

	filesA, err := scanAllChartFiles(dirA)
	if err != nil {
		printErr("Error scanning '%s': %v", dirA, err)
		os.Exit(1)
	}

	filesB, err := scanAllChartFiles(dirB)
	if err != nil {
		printErr("Error scanning '%s': %v", dirB, err)
		os.Exit(1)
	}

	// Build a sorted unique file list.
	allFiles := make(map[string]struct{})
	for f := range filesA {
		allFiles[f] = struct{}{}
	}
	for f := range filesB {
		allFiles[f] = struct{}{}
	}

	var unique []string
	for f := range allFiles {
		unique = append(unique, f)
	}
	sort.Strings(unique)

	var hasDiff bool
	for _, rel := range unique {
		contentA, okA := filesA[rel]
		contentB, okB := filesB[rel]

		// File exists only in dirA.
		if okA && !okB {
			fmt.Println(tui.HeaderStyle.Render("File removed in B: " + rel))
			printDiffHeader(rel, dirA, dirB)
			for _, line := range strings.Split(contentA, "\n") {
				fmt.Println(tui.ErrorStyle.Render("- ") + line)
			}
			hasDiff = true
			fmt.Println()
			continue
		}

		// File exists only in dirB.
		if !okA && okB {
			fmt.Println(tui.HeaderStyle.Render("File added in B: " + rel))
			printDiffHeader(rel, dirA, dirB)
			for _, line := range strings.Split(contentB, "\n") {
				fmt.Println(tui.SuccessStyle.Render("+ ") + line)
			}
			hasDiff = true
			fmt.Println()
			continue
		}

		// Both exist — compare content.
		if !bytes.Equal([]byte(contentA), []byte(contentB)) {
			fmt.Println(tui.HeaderStyle.Render("File changed: " + rel))
			printDiffHeader(rel, dirA, dirB)
			unifiedDiff(string(contentA), string(contentB))
			hasDiff = true
			fmt.Println()
		}
	}

	if !hasDiff {
		fmt.Println(tui.SuccessStyle.Render("No differences found between the two charts."))
	}
}

// resolveDiffPaths resolves dirA and dirB. If omitted, it auto-selects from ~/kgen/.
func resolveDiffPaths(args []string) (string, string) {
	var dirA, dirB string

	if len(args) >= 1 {
		dirA = args[0]
	}
	if len(args) >= 2 {
		dirB = args[1]
	}

	// Resolve missing paths from ~/kgen/.
	if dirA == "" || dirB == "" {
		charts := listAvailableCharts()
		if len(charts) < 2 {
			printErr("Error: Not enough charts in ~/kgen/ to diff (found %d, need 2).", len(charts))
			fmt.Println("Usage: kgen diff [chart-a] [chart-b]")
			os.Exit(1)
		}

		if dirA == "" {
			fmt.Println(tui.HeaderStyle.Render("Select the first chart to compare (A):"))
			for i, c := range charts {
				fmt.Printf("  [%d] %s\n", i+1, c)
			}
			dirA = promptChartChoice(charts)
		}
		if dirB == "" {
			fmt.Println(tui.HeaderStyle.Render("Select the second chart to compare (B):"))
			for i, c := range charts {
				fmt.Printf("  [%d] %s\n", i+1, c)
			}
			dirB = promptChartChoice(charts)
		}
	}

	// Resolve relative to absolute.
	dirA = resolveChartPath(dirA)
	dirB = resolveChartPath(dirB)

	return dirA, dirB
}

// printDiffHeader prints the diff legend for a given file.
func printDiffHeader(rel, dirA, dirB string) {
	fmt.Println()
	fmt.Printf("--- %s/%s\n", filepath.Base(dirA), rel)
	fmt.Printf("+++ %s/%s\n", filepath.Base(dirB), rel)
}

// unifiedDiff prints a simple line-by-line unified diff with context.
func unifiedDiff(contentA, contentB string) {
	linesA := strings.Split(contentA, "\n")
	linesB := strings.Split(contentB, "\n")

	maxLen := len(linesA)
	if len(linesB) > maxLen {
		maxLen = len(linesB)
	}

	ctxBefore, ctxAfter := 3, 3
	// Find hunks — ranges of lines where content differs.
	var hunks []hunk
	i := 0
	for i < maxLen {
		a := getLine(linesA, i)
		b := getLine(linesB, i)
		if a == b {
			i++
			continue
		}
		// Found a difference — find the extent.
		start := i - ctxBefore
		if start < 0 {
			start = 0
		}
		end := i + 1
		for end < maxLen {
			if getLine(linesA, end) != getLine(linesB, end) {
				end++
			} else {
				break
			}
		}
		end += ctxAfter
		if end > maxLen {
			end = maxLen
		}
		hunks = append(hunks, hunk{start: start, end: end})
		i = end
	}

	for _, h := range hunks {
		fmt.Printf("@@ -%d +%d @@\n", h.start+1, h.start+1)
		for i := h.start; i < h.end; i++ {
			a := getLine(linesA, i)
			b := getLine(linesB, i)
			if a == b {
				fmt.Printf("  %s\n", a)
			} else {
				if i < len(linesA) {
					fmt.Println(tui.ErrorStyle.Render("- ") + linesA[i])
				}
				if i < len(linesB) {
					fmt.Println(tui.SuccessStyle.Render("+ ") + linesB[i])
				}
			}
		}
	}
}

type hunk struct {
	start, end int
}

func getLine(lines []string, i int) string {
	if i < len(lines) {
		return lines[i]
	}
	return ""
}
