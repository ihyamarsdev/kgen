package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// listAvailableCharts returns chart directory names found in ~/kgen/.
func listAvailableCharts() []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	kgenDir := filepath.Join(homeDir, "kgen")
	entries, err := os.ReadDir(kgenDir)
	if err != nil {
		return nil
	}

	var charts []string
	for _, entry := range entries {
		if entry.IsDir() && isHelmChart(filepath.Join(kgenDir, entry.Name())) {
			charts = append(charts, entry.Name())
		}
	}
	return charts
}

// promptChartChoice asks the user to pick a chart from the list by number.
// Returns the resolved full path.
func promptChartChoice(charts []string) string {
	var choice int
	_, err := fmt.Scanf("%d", &choice)
	if err != nil || choice < 1 || choice > len(charts) {
		printErr("Invalid selection. Please enter a number between 1 and %d.", len(charts))
		os.Exit(1)
	}
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, "kgen", charts[choice-1])
}

// resolveChartPath resolves a chart path:
//   - If it's an absolute path, return as-is.
//   - If it's relative but exists, resolve to absolute.
//   - If it matches a chart name in ~/kgen/, resolve to ~/kgen/<name>.
func resolveChartPath(path string) string {
	if filepath.IsAbs(path) {
		if isHelmChart(path) {
			return path
		}
		printErr("Error: '%s' is not a valid Helm chart directory.", path)
		os.Exit(1)
	}

	// Try as-is (relative path).
	if abs, err := filepath.Abs(path); err == nil && isHelmChart(abs) {
		return abs
	}

	// Try as a chart name in ~/kgen/.
	homeDir, err := os.UserHomeDir()
	if err == nil {
		candidate := filepath.Join(homeDir, "kgen", path)
		if isHelmChart(candidate) {
			return candidate
		}
	}

	printErr("Error: '%s' is not a valid Helm chart directory.", path)
	os.Exit(1)
	return ""
}

// scanAllChartFiles scans a Helm chart directory for all regular files,
// excluding hidden files, and returns a map of relPath -> content.
func scanAllChartFiles(dir string) (map[string]string, error) {
	files := make(map[string]string)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			rel, err := filepath.Rel(dir, path)
			if err == nil && !isHidden(rel) {
				content, err := os.ReadFile(path)
				if err == nil {
					files[rel] = string(content)
				}
			}
		}
		return nil
	})
	return files, err
}

// isHidden reports whether any directory component of the path starts with a dot.
// The filename itself is NOT checked — only parent directories matter.
// This ensures files like "my-app.deployment.yaml" are not incorrectly excluded.
func isHidden(rel string) bool {
	dir := filepath.Dir(rel)
	if dir == "." {
		return false
	}
	parts := strings.Split(dir, string(filepath.Separator))
	for _, p := range parts {
		if strings.HasPrefix(p, ".") {
			return true
		}
	}
	return false
}
