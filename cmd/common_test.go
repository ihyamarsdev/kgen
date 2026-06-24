package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsHelmChart(t *testing.T) {
	dir := t.TempDir()

	// Not a chart — missing both files.
	if isHelmChart(dir) {
		t.Error("expected false for empty directory")
	}

	// Not a chart — only Chart.yaml.
	writeFile(t, filepath.Join(dir, "Chart.yaml"), "apiVersion: v2\n")
	if isHelmChart(dir) {
		t.Error("expected false with only Chart.yaml")
	}

	// Chart — both files present.
	writeFile(t, filepath.Join(dir, "values.yaml"), "replicaCount: 1\n")
	if !isHelmChart(dir) {
		t.Error("expected true with both Chart.yaml and values.yaml")
	}
}

func TestIsHidden(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"Chart.yaml", false},
		{"templates/deployment.yaml", false},
		{".gitignore", true},
		{".hidden/file.yaml", true},
		{"templates/.hidden/file.yaml", true},
		{"my-app.deployment.yaml", false}, // should NOT be hidden — only dir components matter
		{"config/my-app.yaml", false},
	}
	for _, tt := range tests {
		got := isHidden(tt.path)
		if got != tt.expected {
			t.Errorf("isHidden(%q) = %v, want %v", tt.path, got, tt.expected)
		}
	}
}

func TestFindEditor(t *testing.T) {
	// Save original EDITOR.
	orig := os.Getenv("EDITOR")
	defer os.Setenv("EDITOR", orig)

	// When EDITOR is set, it should be returned.
	os.Setenv("EDITOR", "myeditor")
	if got := findEditor(); got != "myeditor" {
		t.Errorf("findEditor() = %q, want %q", got, "myeditor")
	}

	// When EDITOR is empty, it should fall back to system editors.
	os.Setenv("EDITOR", "")
	got := findEditor()
	// Should return nano, vim, vi, or empty depending on what's installed.
	if got != "" && got != "nano" && got != "vim" && got != "vi" {
		t.Errorf("findEditor() = %q, want one of: nano, vim, vi, or empty", got)
	}
}

func TestHomeDir(t *testing.T) {
	got := homeDir()
	if got == "" {
		t.Error("homeDir() returned empty string")
	}
}

func TestScanAllChartFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "Chart.yaml"), "apiVersion: v2\n")
	writeFile(t, filepath.Join(dir, "values.yaml"), "replicaCount: 1\n")
	os.MkdirAll(filepath.Join(dir, "templates"), 0755)
	writeFile(t, filepath.Join(dir, "templates", "deployment.yaml"), "apiVersion: apps/v1\n")
	// Hidden file — should be excluded.
	writeFile(t, filepath.Join(dir, "templates", ".hidden.yaml"), "hidden\n")

	files, err := scanAllChartFiles(dir)
	if err != nil {
		t.Fatalf("scanAllChartFiles() error: %v", err)
	}

	if len(files) != 3 {
		t.Errorf("scanAllChartFiles() returned %d files, want 3", len(files))
	}
	if _, ok := files["Chart.yaml"]; !ok {
		t.Error("expected Chart.yaml in results")
	}
	if _, ok := files["values.yaml"]; !ok {
		t.Error("expected values.yaml in results")
	}
	if _, ok := files["templates/deployment.yaml"]; !ok {
		t.Error("expected templates/deployment.yaml in results")
	}
	if _, ok := files["templates/.hidden.yaml"]; ok {
		t.Error("hidden file should not be in results")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}

func TestReadChartNamespace(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{"with namespace", "replicaCount: 1\nnamespace: production\n", "production"},
		{"without namespace", "replicaCount: 1\n", "default"},
		{"quoted namespace", "namespace: \"staging\"\n", "staging"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, _ := os.MkdirTemp("", "kgen-ns-test-*")
			defer os.RemoveAll(tmpDir)

			os.WriteFile(filepath.Join(tmpDir, "values.yaml"), []byte(tt.content), 0644)
			got := readChartNamespace(tmpDir)
			if got != tt.want {
				t.Errorf("readChartNamespace() = %q, want %q", got, tt.want)
			}
		})
	}
}
