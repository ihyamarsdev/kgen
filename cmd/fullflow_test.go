package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ihyamarsdev/kgen/internal/generator"
)

func TestE2E_CreateToDeploy(t *testing.T) {
	// Step 1: Generate chart using the generator
	tmpDir, err := os.MkdirTemp("", "kgen-flow-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := generator.Config{
		AppName:            "flow-test",
		Namespace:          "flow-ns",
		ImageRepository:    "nginx",
		ImageTag:           "1.25",
		Port:               8080,
		ReplicaCount:       2,
		TemplateQuality:    "production",
		ServiceType:        "ClusterIP",
		HPAEnabled:         true,
		HPAMinReplicas:     2,
		HPAMaxReplicas:     6,
		GenerateDeployment: true,
		GenerateService:    true,
		GenerateHPA:        true,
		GeneratePDB:        true,
		IngressEnabled:     false,
	}

	if err := generator.Generate(cfg, tmpDir); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Step 2: Verify chart files exist
	expectedFiles := []string{
		"Chart.yaml",
		"values.yaml",
		filepath.Join("templates", "_helpers.tpl"),
		filepath.Join("templates", "deployment.yaml"),
		filepath.Join("templates", "service.yaml"),
		filepath.Join("templates", "hpa.yaml"),
		filepath.Join("templates", "pdb.yaml"),
	}

	for _, f := range expectedFiles {
		p := filepath.Join(tmpDir, f)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", f)
		}
	}

	// Step 3: Verify namespace is in values.yaml
	values, err := os.ReadFile(filepath.Join(tmpDir, "values.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !containsStr(string(values), "namespace: flow-ns") {
		t.Error("values.yaml should contain 'namespace: flow-ns'")
	}

	// Step 4: Verify autoscaling block exists
	if !containsStr(string(values), "autoscaling:") {
		t.Error("values.yaml should contain 'autoscaling:' block")
	}
	if !containsStr(string(values), "enabled: true") {
		t.Error("values.yaml should contain 'enabled: true' for autoscaling")
	}

	// Step 5: Verify service type
	if !containsStr(string(values), "type: ClusterIP") {
		t.Error("values.yaml should contain 'type: ClusterIP'")
	}

	// Step 6: Verify replicaCount
	if !containsStr(string(values), "replicaCount: 2") {
		t.Error("values.yaml should contain 'replicaCount: 2'")
	}

	// Step 7: Verify HPA config
	if !containsStr(string(values), "minReplicas: 2") {
		t.Error("values.yaml should contain 'minReplicas: 2'")
	}
	if !containsStr(string(values), "maxReplicas: 6") {
		t.Error("values.yaml should contain 'maxReplicas: 6'")
	}

	t.Logf("Chart generated successfully at %s", tmpDir)
	t.Logf("All %d files verified", len(expectedFiles))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
