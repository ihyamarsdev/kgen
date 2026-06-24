package cmd

import (
	"testing"
)

func TestBuildDeployArgs_HasCreateNamespace(t *testing.T) {
	args := buildDeployArgs("my-release", "/some/chart", "my-ns")

	found := false
	for _, a := range args {
		if a == "--create-namespace" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected --create-namespace in args, got: %v", args)
	}

	// Verify the args order
	expectedOrder := []string{"install", "my-release", "/some/chart", "--namespace", "my-ns", "--create-namespace"}
	for i, exp := range expectedOrder {
		if args[i] != exp {
			t.Errorf("args[%d] = %q, want %q (full args: %v)", i, args[i], exp, args)
		}
	}
}

func TestBuildUpgradeArgs_NoCreateNamespace(t *testing.T) {
	// Upgrade doesn't need --create-namespace (namespace already exists)
	args := buildUpgradeArgs("my-release", "/some/chart", "my-ns")

	for _, a := range args {
		if a == "--create-namespace" {
			t.Errorf("upgrade should NOT have --create-namespace, got: %v", args)
		}
	}
}

func TestBuildDeployArgs_WithDryRun(t *testing.T) {
	deployDryRun = true
	defer func() { deployDryRun = false }()

	args := buildDeployArgs("rel", "/chart", "ns")
	found := false
	for _, a := range args {
		if a == "--dry-run" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected --dry-run in args, got: %v", args)
	}
}
