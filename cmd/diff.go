package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ihyamarsdev/kgen/internal/tui"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff [chart-a] [chart-b]",
	Short: "Compare two Helm chart directories",
	Long: `Compare two generated Helm chart directories and display a unified diff
of their file contents. Hidden files (dotfiles) and empty directories
are skipped.

If either path is omitted, kgen will attempt to auto-select from ~/.kgen/.`,
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

	// Guard: prevent diffing the same chart against itself.
	if filepath.Clean(dirA) == filepath.Clean(dirB) {
		printErr("Error: Both paths point to the same chart (%s).", dirA)
		os.Exit(1)
	}

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

		// Both exist — compare content with proper LCS-based diff.
		if !bytes.Equal([]byte(contentA), []byte(contentB)) {
			fmt.Println(tui.HeaderStyle.Render("File changed: " + rel))
			printDiffHeader(rel, dirA, dirB)
			printUnifiedDiff(contentA, contentB)
			hasDiff = true
			fmt.Println()
		}
	}

	if !hasDiff {
		fmt.Println(tui.SuccessStyle.Render("No differences found between the two charts."))
	} else {
		os.Exit(1)
	}
}

// resolveDiffPaths resolves dirA and dirB. If omitted, it auto-selects from ~/.kgen/.
func resolveDiffPaths(args []string) (string, string) {
	var dirA, dirB string

	if len(args) >= 1 {
		dirA = args[0]
	}
	if len(args) >= 2 {
		dirB = args[1]
	}

	// Resolve missing paths from ~/.kgen/.
	if dirA == "" || dirB == "" {
		charts := listAvailableCharts()
		if len(charts) < 2 {
			printErr("Error: Not enough charts in ~/.kgen/ to diff (found %d, need 2).", len(charts))
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

// printUnifiedDiff generates and prints a proper unified diff using
// the go-diff library (LCS-based algorithm) instead of naive line-by-line comparison.
func printUnifiedDiff(contentA, contentB string) {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(contentA, contentB, false)
	diffs = dmp.DiffCleanupSemantic(diffs)

	// Build unified diff hunks using go-diff's output.
	var result []string
	oldLine, newLine := 1, 1
	var hunkLines []string
	hunkOldStart, hunkNewStart := 0, 0
	inHunk := false

	flushHunk := func() {
		if len(hunkLines) == 0 {
			return
		}
		if hunkOldStart == 0 {
			hunkOldStart = oldLine
		}
		if hunkNewStart == 0 {
			hunkNewStart = newLine
		}
		result = append(result, fmt.Sprintf("@@ -%d +%d @@ ", hunkOldStart, hunkNewStart))
		result = append(result, hunkLines...)
		hunkLines = nil
		hunkOldStart = 0
		hunkNewStart = 0
		inHunk = false
	}

	for _, diff := range diffs {
		switch diff.Type {
		case diffmatchpatch.DiffEqual:
			equalLines := strings.Split(diff.Text, "\n")
			// Remove trailing empty element from split.
			if len(equalLines) > 0 && equalLines[len(equalLines)-1] == "" {
				equalLines = equalLines[:len(equalLines)-1]
			}
			for _, line := range equalLines {
				// Context line: show only if we're in or near a hunk.
				if inHunk {
					hunkLines = append(hunkLines, " "+line)
				}
				oldLine++
				newLine++
			}
		case diffmatchpatch.DiffDelete:
			if !inHunk {
				hunkOldStart = oldLine
				hunkNewStart = newLine
				inHunk = true
			}
			deleteLines := strings.Split(diff.Text, "\n")
			if len(deleteLines) > 0 && deleteLines[len(deleteLines)-1] == "" {
				deleteLines = deleteLines[:len(deleteLines)-1]
			}
			for _, line := range deleteLines {
				hunkLines = append(hunkLines, tui.ErrorStyle.Render("-")+line)
				oldLine++
			}
		case diffmatchpatch.DiffInsert:
			if !inHunk {
				hunkOldStart = oldLine
				hunkNewStart = newLine
				inHunk = true
			}
			insertLines := strings.Split(diff.Text, "\n")
			if len(insertLines) > 0 && insertLines[len(insertLines)-1] == "" {
				insertLines = insertLines[:len(insertLines)-1]
			}
			for _, line := range insertLines {
				hunkLines = append(hunkLines, tui.SuccessStyle.Render("+")+line)
				newLine++
			}
		}
	}
	flushHunk()

	for _, line := range result {
		fmt.Println(line)
	}
}
