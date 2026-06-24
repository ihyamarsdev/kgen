package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

// confirm prompts the user with a yes/no question on stdin.
//
// It returns true when the user answers "y" or "yes" (case-insensitive) and
// false otherwise. An empty answer (just pressing Enter) is treated as "no",
// matching the conventional "y/N" semantics used throughout the CLI.
//
// When stdin is not a terminal (e.g. piped input, EOF immediately) the answer
// is treated as "no" so that destructive operations never run unattended.
func confirm(prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes"
}

// printErr writes a styled error message to stderr.
func printErr(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// findEditor returns the user's preferred terminal editor.
//
// It checks $EDITOR first, then falls back to nano, vim, vi (in that order)
// based on what is available in PATH. Returns an empty string if none found.
func findEditor() string {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	for _, e := range []string{"nano", "vim", "vi"} {
		if _, err := exec.LookPath(e); err == nil {
			return e
		}
	}
	return ""
}

// homeDir returns the current user's home directory, or an empty string on error.
//
// It tries os/user.Current first (more reliable in some environments) and falls
// back to os.UserHomeDir.
func homeDir() string {
	if usr, err := user.Current(); err == nil && usr.HomeDir != "" {
		return usr.HomeDir
	}
	if dir, err := os.UserHomeDir(); err == nil {
		return dir
	}
	return ""
}

// findHelm checks if the helm CLI binary is available in PATH.
//
// Returns the full path to the helm binary or an empty string if not found.
func findHelm() string {
	path, err := exec.LookPath("helm")
	if err != nil {
		return ""
	}
	return path
}

// helmOutput runs a helm command and returns stdout+stderr combined.
func helmOutput(args ...string) (string, error) {
	cmd := exec.Command("helm", args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return buf.String(), err
}

// helmRun runs a helm command, streaming output directly to the terminal.
//
// Returns true if the command succeeded, false otherwise.
func helmRun(args ...string) error {
	cmd := exec.Command("helm", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// releaseNameFromChart derives a sensible Helm release name from a chart
// directory. It uses the directory basename, sanitised to be DNS-compatible.
func releaseNameFromChart(chartDir string) string {
	name := strings.ToLower(filepath.Base(chartDir))
	// Replace non-alphanumeric characters with dashes.
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			result.WriteRune(r)
		} else {
			if result.Len() > 0 && result.String()[result.Len()-1] != '-' {
				result.WriteRune('-')
			}
		}
	}
	name = result.String()
	name = strings.Trim(name, "-")
	if name == "" {
		name = "kgen"
	}
	return name
}

// helmReleaseExists checks if a Helm release is already installed in a namespace.
func helmReleaseExists(release, namespace string) bool {
	out, err := helmOutput("list", "--namespace", namespace, "--short")
	if err != nil {
		return false
	}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == release {
			return true
		}
	}
	return false
}

// readChartNamespace reads the namespace from a chart's values.yaml.
// It looks for a top-level "namespace:" key. Returns "default" if not found.
func readChartNamespace(chartDir string) string {
	valuesPath := filepath.Join(chartDir, "values.yaml")
	data, err := os.ReadFile(valuesPath)
	if err != nil {
		return "default"
	}
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "namespace:") {
			val := strings.TrimSpace(strings.TrimPrefix(trimmed, "namespace:"))
			val = strings.Trim(val, "\"'")
			if val != "" {
				return val
			}
		}
	}
	return "default"
}
