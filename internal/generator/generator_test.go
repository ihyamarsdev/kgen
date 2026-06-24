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
		TemplateQuality:    "production",
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

func TestGenerate_AllResources(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "kgen-all-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := Config{
		AppName:                    "all-app",
		Namespace:                  "default",
		ImageRepository:            "myapp",
		ImageTag:                   "v1",
		Port:                       80,
		ReplicaCount:               1,
		TemplateQuality:            "basic",
		GenerateDeployment:         true,
		GenerateService:            true,
		GenerateIngress:            true,
		GenerateGateway:            true,
		GenerateConfigMap:          true,
		GenerateSecret:             true,
		GenerateExternalSecret:     true,
		GenerateSealedSecret:       true,
		GenerateHPA:                true,
		GenerateVPA:                true,
		GenerateKEDA:               true,
		GenerateStatefulSet:        true,
		GenerateCronJob:            true,
		GenerateDaemonSet:          true,
		GenerateJob:                true,
		GenerateArgoCD:             true,
		GenerateArgoCDSet:          true,
		GenerateIstio:              true,
		GeneratePVC:                true,
		GenerateNetworkPolicy:      true,
		GenerateServiceMonitor:     true,
		GeneratePodMonitor:         true,
		GeneratePrometheusRule:     true,
		GenerateGrafanaDashboard:   true,
		GeneratePDB:                true,
		GeneratePriorityClass:      true,
		GenerateServiceAccount:     true,
		GenerateRole:               true,
		GenerateRoleBinding:        true,
		GenerateClusterRole:        true,
		GenerateClusterRoleBinding: true,
		GenerateFlux:               true,
	}

	if err := Generate(cfg, tmpDir); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	expectedFiles := []string{
		"Chart.yaml", "values.yaml",
		filepath.Join("templates", "_helpers.tpl"),
		filepath.Join("templates", "deployment.yaml"),
		filepath.Join("templates", "service.yaml"),
		filepath.Join("templates", "ingress.yaml"),
		filepath.Join("templates", "gateway.yaml"),
		filepath.Join("templates", "httproute.yaml"),
		filepath.Join("templates", "configmap.yaml"),
		filepath.Join("templates", "secret.yaml"),
		filepath.Join("templates", "externalsecret.yaml"),
		filepath.Join("templates", "sealedsecret.yaml"),
		filepath.Join("templates", "hpa.yaml"),
		filepath.Join("templates", "vpa.yaml"),
		filepath.Join("templates", "scaledobject.yaml"),
		filepath.Join("templates", "statefulset.yaml"),
		filepath.Join("templates", "cronjob.yaml"),
		filepath.Join("templates", "daemonset.yaml"),
		filepath.Join("templates", "job.yaml"),
		filepath.Join("templates", "application.yaml"),
		filepath.Join("templates", "applicationset.yaml"),
		filepath.Join("templates", "virtualservice.yaml"),
		filepath.Join("templates", "pvc.yaml"),
		filepath.Join("templates", "networkpolicy.yaml"),
		filepath.Join("templates", "servicemonitor.yaml"),
		filepath.Join("templates", "podmonitor.yaml"),
		filepath.Join("templates", "prometheusrule.yaml"),
		filepath.Join("templates", "grafanadashboard.yaml"),
		filepath.Join("templates", "pdb.yaml"),
		filepath.Join("templates", "priorityclass.yaml"),
		filepath.Join("templates", "serviceaccount.yaml"),
		filepath.Join("templates", "role.yaml"),
		filepath.Join("templates", "rolebinding.yaml"),
		filepath.Join("templates", "clusterrole.yaml"),
		filepath.Join("templates", "clusterrolebinding.yaml"),
		filepath.Join("templates", "helmrelease.yaml"),
		filepath.Join("templates", "fluxkustomization.yaml"),
	}

	for _, f := range expectedFiles {
		p := filepath.Join(tmpDir, f)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", f)
		}
	}
}

func TestGenerate_MinimalConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "kgen-minimal-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Only core templates enabled.
	cfg := Config{
		AppName:            "minimal",
		ImageRepository:    "nginx",
		ImageTag:           "latest",
		Port:               80,
		ReplicaCount:       1,
		TemplateQuality:    "basic",
		GenerateDeployment: true,
		GenerateService:    true,
	}

	if err := Generate(cfg, tmpDir); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Core files should exist.
	for _, f := range []string{
		"Chart.yaml", "values.yaml",
		filepath.Join("templates", "_helpers.tpl"),
		filepath.Join("templates", "deployment.yaml"),
		filepath.Join("templates", "service.yaml"),
	} {
		p := filepath.Join(tmpDir, f)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", f)
		}
	}

	// Optional files should NOT exist.
	for _, f := range []string{
		filepath.Join("templates", "ingress.yaml"),
		filepath.Join("templates", "hpa.yaml"),
		filepath.Join("templates", "pdb.yaml"),
	} {
		p := filepath.Join(tmpDir, f)
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("Expected file %s to NOT exist", f)
		}
	}
}

// TestGenerate_ValuesAlwaysContainsAutoscaling verifies that values.yaml
// always contains the autoscaling block — deployment.yaml references
// .Values.autoscaling.enabled and nil causes a helm install failure.
func TestGenerate_ValuesAlwaysContainsAutoscaling(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "kgen-nohpa-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// HPA disabled — the exact scenario that caused the bug.
	cfg := Config{
		AppName:            "nohpa",
		Namespace:          "default",
		ImageRepository:    "nginx",
		ImageTag:           "latest",
		Port:               80,
		ReplicaCount:       1,
		HPAEnabled:         false,
		TemplateQuality:    "basic",
		GenerateDeployment: true,
		GenerateService:    true,
		GenerateHPA:        false,
	}

	if err := Generate(cfg, tmpDir); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, "values.yaml"))
	if err != nil {
		t.Fatalf("failed to read values.yaml: %v", err)
	}

	content := string(data)
	if !containsString(content, "autoscaling:") {
		t.Fatal("values.yaml is missing 'autoscaling:' block — deployment.yaml will fail with nil pointer")
	}
	if !containsString(content, "enabled: false") {
		t.Fatal("values.yaml autoscaling.enabled should be 'false' when HPA is disabled")
	}
}

func containsString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
