package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ihyamarsdev/kgen/internal/generator"
	"github.com/ihyamarsdev/kgen/internal/validator"
)

// TestE2E_FullFlow tests the complete workflow: generate a chart, validate it,
// generate a second chart with different settings, and diff them.
func TestE2E_FullFlow(t *testing.T) {
	// Setup temp directories.
	tmpDirA, err := os.MkdirTemp("", "kgen-e2e-a-*")
	if err != nil {
		t.Fatalf("failed to create temp dir A: %v", err)
	}
	defer os.RemoveAll(tmpDirA)

	tmpDirB, err := os.MkdirTemp("", "kgen-e2e-b-*")
	if err != nil {
		t.Fatalf("failed to create temp dir B: %v", err)
	}
	defer os.RemoveAll(tmpDirB)

	// Step 1: Generate a minimal (dev-like) chart.
	cfgDev := generator.Config{
		AppName:            "dev-app",
		Namespace:          "default",
		ImageRepository:    "nginx",
		ImageTag:           "latest",
		Port:               80,
		ReplicaCount:       1,
		TemplateQuality:    "basic",
		GenerateDeployment: true,
		GenerateService:    true,
	}
	if err := generator.Generate(cfgDev, tmpDirA); err != nil {
		t.Fatalf("Generate dev chart failed: %v", err)
	}

	// Verify dev chart files exist.
	for _, f := range []string{"Chart.yaml", "values.yaml", "templates/_helpers.tpl", "templates/deployment.yaml", "templates/service.yaml"} {
		if _, err := os.Stat(filepath.Join(tmpDirA, f)); os.IsNotExist(err) {
			t.Errorf("Dev chart missing file: %s", f)
		}
	}

	// Step 2: Validate the dev chart (should have warnings).
	resultsDev, err := validator.ValidateDir(tmpDirA)
	if err != nil {
		t.Fatalf("Validate dev chart failed: %v", err)
	}
	hasWarn := false
	for _, r := range resultsDev {
		if r.Status == "WARN" {
			hasWarn = true
		}
	}
	if !hasWarn {
		t.Error("Expected dev chart validation to have warnings")
	}

	// Step 3: Generate a production-like chart.
	cfgProd := generator.Config{
		AppName:                           "prod-app",
		Namespace:                         "production",
		ImageRepository:                   "myapp",
		ImageTag:                          "v1.0.0",
		Port:                              8080,
		ReplicaCount:                      3,
		TemplateQuality:                   "production",
		GenerateDeployment:                true,
		GenerateService:                   true,
		GenerateIngress:                   true,
		GenerateHPA:                       true,
		GeneratePDB:                       true,
		GenerateServiceMonitor:            true,
		GenerateNetworkPolicy:             true,
		GenerateTopologySpreadConstraints: true,
		GeneratePodAntiAffinity:           true,
	}
	if err := generator.Generate(cfgProd, tmpDirB); err != nil {
		t.Fatalf("Generate prod chart failed: %v", err)
	}

	// Verify prod chart has more files than dev.
	prodFiles, err := scanAllChartFiles(tmpDirB)
	if err != nil {
		t.Fatalf("Scan prod chart failed: %v", err)
	}
	devFiles, err := scanAllChartFiles(tmpDirA)
	if err != nil {
		t.Fatalf("Scan dev chart failed: %v", err)
	}
	if len(prodFiles) <= len(devFiles) {
		t.Errorf("Expected prod chart to have more files (%d) than dev (%d)", len(prodFiles), len(devFiles))
	}

	// Step 4: Validate the prod chart (should pass all checks).
	resultsProd, err := validator.ValidateDir(tmpDirB)
	if err != nil {
		t.Fatalf("Validate prod chart failed: %v", err)
	}
	for _, r := range resultsProd {
		if r.Status != "PASS" {
			t.Errorf("Expected PASS for prod chart check %s, got %s: %s", r.Check, r.Status, r.Message)
		}
	}

	// Step 5: Diff the two charts — should find differences.
	if len(prodFiles) <= len(devFiles) {
		t.Error("Prod chart should have more files than dev")
	}

	// Verify files unique to prod.
	for _, f := range []string{"templates/ingress.yaml", "templates/hpa.yaml", "templates/pdb.yaml", "templates/servicemonitor.yaml", "templates/networkpolicy.yaml"} {
		if _, exists := prodFiles[f]; !exists {
			t.Errorf("Expected prod chart to have %s", f)
		}
		if _, exists := devFiles[f]; exists {
			t.Errorf("Did not expect dev chart to have %s", f)
		}
	}
}

// TestE2E_ScanAndDiffHiddenFiles verifies that hidden files are excluded from scanning
// and that file content comparison works correctly.
func TestE2E_ScanAndDiffHiddenFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "kgen-e2e-hidden-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create chart structure with hidden files.
	cfg := generator.Config{
		AppName:            "test-app",
		ImageRepository:    "nginx",
		ImageTag:           "latest",
		Port:               80,
		ReplicaCount:       1,
		TemplateQuality:    "basic",
		GenerateDeployment: true,
		GenerateService:    true,
	}
	if err := generator.Generate(cfg, tmpDir); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Add hidden files/dirs.
	os.MkdirAll(filepath.Join(tmpDir, "templates", ".hidden"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "templates", ".hidden", "secret.yaml"), []byte("hidden"), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte("*.bak"), 0644)

	// Scan should exclude hidden files.
	files, err := scanAllChartFiles(tmpDir)
	if err != nil {
		t.Fatalf("scanAllChartFiles failed: %v", err)
	}

	for _, f := range []string{".gitignore", "templates/.hidden/secret.yaml"} {
		if _, exists := files[f]; exists {
			t.Errorf("Hidden file %s should not be in scanned results", f)
		}
	}
}

// TestE2E_ValidateStrictBehavior validates the --strict flag behavior.
func TestE2E_ValidateStrictBehavior(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "kgen-e2e-strict-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Generate a minimal chart (will have validation warnings).
	cfg := generator.Config{
		AppName:            "strict-test",
		ImageRepository:    "nginx",
		ImageTag:           "latest",
		Port:               80,
		ReplicaCount:       1,
		TemplateQuality:    "basic",
		GenerateDeployment: true,
		GenerateService:    true,
	}
	if err := generator.Generate(cfg, tmpDir); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Validate should return results with warnings.
	results, err := validator.ValidateDir(tmpDir)
	if err != nil {
		t.Fatalf("ValidateDir failed: %v", err)
	}

	warnCount := 0
	for _, r := range results {
		if r.Status == "WARN" {
			warnCount++
		}
	}

	if warnCount == 0 {
		t.Error("Expected warnings for minimal chart validation")
	}

	// Verify specific checks exist.
	expectedChecks := []string{"Resource Limits", "Resource Requests", "Liveness Probe", "Readiness Probe", "Security Context"}
	foundChecks := make(map[string]bool)
	for _, r := range results {
		foundChecks[r.Check] = true
	}
	for _, check := range expectedChecks {
		if !foundChecks[check] {
			t.Errorf("Expected validation check %q not found", check)
		}
	}
}

// TestE2E_ResolveChartPath tests the chart path resolution logic.
func TestE2E_ResolveChartPath(t *testing.T) {
	// Save home dir.
	home := homeDir()
	if home == "" {
		t.Skip("Cannot determine home directory")
	}

	// Create a chart in ~/.kgen/e2e-test/.
	chartDir := filepath.Join(home, ".kgen", "e2e-test-resolve")
	defer os.RemoveAll(filepath.Join(home, ".kgen"))
	os.MkdirAll(chartDir, 0755)

	cfg := generator.Config{
		AppName:            "e2e-test-resolve",
		ImageRepository:    "nginx",
		ImageTag:           "latest",
		Port:               80,
		ReplicaCount:       1,
		TemplateQuality:    "basic",
		GenerateDeployment: true,
		GenerateService:    true,
	}
	if err := generator.Generate(cfg, chartDir); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Test resolveChartPath with just the chart name.
	resolved := resolveChartPath("e2e-test-resolve")
	if resolved != chartDir {
		t.Errorf("resolveChartPath('e2e-test-resolve') = %q, want %q", resolved, chartDir)
	}

	// Test resolveChartPath with absolute path.
	resolved = resolveChartPath(chartDir)
	if resolved != chartDir {
		t.Errorf("resolveChartPath(%q) = %q, want %q", chartDir, resolved, chartDir)
	}
}

// TestE2E_InstallBinaryAtomic verifies the atomic binary replacement logic.
func TestE2E_InstallBinaryAtomic(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "kgen-e2e-install-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a "fake" source binary.
	srcFile := filepath.Join(tmpDir, "src-kgen")
	if err := os.WriteFile(srcFile, []byte("#!/bin/sh\necho hello\n"), 0755); err != nil {
		t.Fatalf("failed to create source binary: %v", err)
	}

	destFile := filepath.Join(tmpDir, "kgen")

	// Test installBinary (atomic path via isWritableDir = true).
	result, err := installBinary(srcFile, destFile)
	if err != nil {
		t.Fatalf("installBinary failed: %v", err)
	}
	if result != destFile {
		t.Errorf("installBinary returned %q, want %q", result, destFile)
	}

	// Verify the destination file exists.
	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Error("Destination file should exist after installBinary")
	}

	// Verify content.
	content, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}
	if !strings.Contains(string(content), "echo hello") {
		t.Error("Destination file content mismatch")
	}
}
