package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "kgen-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := Config{
		AppName:            "test-app",
		Namespace:          "test-ns",
		ImageRepository:    "nginx",
		ImageTag:           "alpine",
		Port:               8080,
		ReplicaCount:       2,
		IngressEnabled:     true,
		HPAEnabled:         true,
		HPAMinReplicas:     2,
		HPAMaxReplicas:     6,
		ProdProfile:        true,
		GenerateDeployment: true,
		GenerateService:    true,
		GenerateIngress:    true,
		GenerateHPA:        true,
	}

	err = Generate(cfg, tmpDir)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	files := []string{
		"Chart.yaml",
		"values.yaml",
		filepath.Join("templates", "_helpers.tpl"),
		filepath.Join("templates", "deployment.yaml"),
		filepath.Join("templates", "service.yaml"),
		filepath.Join("templates", "ingress.yaml"),
		filepath.Join("templates", "hpa.yaml"),
	}

	for _, f := range files {
		p := filepath.Join(tmpDir, f)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", f)
		}
	}
}
