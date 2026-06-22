package validator

import (
	"os"
	"testing"

	"kgen/internal/generator"
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
