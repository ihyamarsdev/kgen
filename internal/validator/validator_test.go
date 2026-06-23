package validator

import (
	"os"
	"testing"

	"github.com/ihyamarsdev/kgen/internal/generator"
)

func TestValidateDir(t *testing.T) {
	// 1. Test Dev Profile (should have warnings)
	tmpDirDev, err := os.MkdirTemp("", "kgen-val-dev-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDirDev)

	cfgDev := generator.Config{
		AppName:            "dev-app",
		Namespace:          "default",
		ImageRepository:    "nginx",
		ImageTag:           "latest",
		Port:               80,
		ReplicaCount:       1,
		IngressEnabled:     false,
		HPAEnabled:         false,
		ProdProfile:        false,
		TemplateQuality:    "basic",
		GenerateDeployment: true,
		GenerateService:    true,
	}

	err = generator.Generate(cfgDev, tmpDirDev)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	resultsDev, err := ValidateDir(tmpDirDev)
	if err != nil {
		t.Fatalf("ValidateDir failed: %v", err)
	}

	hasWarn := false
	for _, res := range resultsDev {
		if res.Status == "WARN" {
			hasWarn = true
		}
	}
	if !hasWarn {
		t.Errorf("Expected dev profile to have warnings, but it had none")
	}

	// 2. Test Prod Profile (should pass)
	tmpDirProd, err := os.MkdirTemp("", "kgen-val-prod-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDirProd)

	cfgProd := generator.Config{
		AppName:            "prod-app",
		Namespace:          "default",
		ImageRepository:    "nginx",
		ImageTag:           "latest",
		Port:               80,
		ReplicaCount:       3,
		IngressEnabled:     true,
		HPAEnabled:         true,
		ProdProfile:        true,
		TemplateQuality:    "production",
		GenerateDeployment: true,
		GenerateService:    true,
		GenerateIngress:    true,
		GenerateHPA:        true,
	}

	err = generator.Generate(cfgProd, tmpDirProd)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	resultsProd, err := ValidateDir(tmpDirProd)
	if err != nil {
		t.Fatalf("ValidateDir failed: %v", err)
	}

	for _, res := range resultsProd {
		if res.Status != "PASS" {
			t.Errorf("Expected PASS for check %s, got %s: %s", res.Check, res.Status, res.Message)
		}
	}
}

func TestValidateDir_FallbackPath(t *testing.T) {
	// Test validation without values.yaml — the string-based fallback should be used.
	tmpDir, err := os.MkdirTemp("", "kgen-val-fallback-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write a raw deployment YAML with probes and limits.
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
spec:
  template:
    spec:
      containers:
      - name: app
        image: nginx:latest
        resources:
          limits:
            cpu: "1"
            memory: "128Mi"
          requests:
            cpu: "100m"
            memory: "64Mi"
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
        securityContext:
          runAsNonRoot: true
`
	if err := os.WriteFile(tmpDir+"/deployment.yaml", []byte(deployment), 0644); err != nil {
		t.Fatalf("failed to write deployment: %v", err)
	}

	results, err := ValidateDir(tmpDir)
	if err != nil {
		t.Fatalf("ValidateDir failed: %v", err)
	}

	// All checks should PASS via the fallback string-based scanning.
	for _, res := range results {
		if res.Status != "PASS" {
			t.Errorf("Expected PASS for check %s (fallback path), got %s: %s", res.Check, res.Status, res.Message)
		}
	}
}

func TestValidateDir_NonExistentDir(t *testing.T) {
	_, err := ValidateDir("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}

func TestHasKeyPath(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]any
		path     []string
		expected bool
	}{
		{"direct key", map[string]any{"foo": "bar"}, []string{"foo"}, true},
		{"missing key", map[string]any{"foo": "bar"}, []string{"baz"}, false},
		{"nested key", map[string]any{"a": map[string]any{"b": "c"}}, []string{"a", "b"}, true},
		{"missing nested key", map[string]any{"a": map[string]any{"b": "c"}}, []string{"a", "x"}, false},
		{"nil value", map[string]any{"foo": nil}, []string{"foo"}, false},
		{"empty path", map[string]any{"foo": "bar"}, []string{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasKeyPath(tt.data, tt.path...)
			if got != tt.expected {
				t.Errorf("hasKeyPath(%v, %v) = %v, want %v", tt.data, tt.path, got, tt.expected)
			}
		})
	}
}
