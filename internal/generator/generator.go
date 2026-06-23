package generator

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates/*
var templateFS embed.FS

type Config struct {
	AppName         string
	Namespace       string
	ImageRepository string
	ImageTag        string
	Port            int
	ReplicaCount    int
	IngressEnabled  bool
	HPAEnabled      bool
	HPAMinReplicas  int
	HPAMaxReplicas  int
	ProdProfile     bool

	// Template Quality Level: basic, production, enterprise
	TemplateQuality string

	// Secret Backend for ExternalSecret: vault, aws, gcp, azure
	SecretBackend string

	// Smart Wizard Inputs
	StorageClass        string
	StorageSize         string
	StorageAccessMode   string
	ServiceAccountName  string
	RbacLevel           string
	RbacCustomResources []string
	IngressTlsEnabled   bool
	IngressTlsProvider  string
	NetworkPolicyPreset string

	// Resources to generate
	GenerateDeployment                bool
	GenerateService                   bool
	GenerateIngress                   bool
	GenerateGateway                   bool
	GenerateConfigMap                 bool
	GenerateSecret                    bool
	GenerateExternalSecret            bool
	GenerateSealedSecret              bool
	GenerateHPA                       bool
	GenerateServiceMonitor            bool
	GeneratePDB                       bool
	GenerateVPA                       bool
	GenerateKEDA                      bool
	GenerateStatefulSet               bool
	GenerateCronJob                   bool
	GenerateArgoCD                    bool
	GenerateIstio                     bool
	GeneratePVC                       bool
	GenerateNetworkPolicy             bool
	GenerateDaemonSet                 bool
	GenerateJob                       bool
	GenerateServiceAccount            bool
	GenerateRbac                      bool // Role and RoleBinding
	GenerateRole                      bool
	GenerateRoleBinding               bool
	GenerateClusterRole               bool
	GenerateClusterRoleBinding        bool
	GeneratePriorityClass             bool
	GeneratePodMonitor                bool
	GeneratePrometheusRule            bool
	GenerateGrafanaDashboard          bool
	GenerateArgoCDSet                 bool
	GenerateFlux                      bool
	GeneratePodAntiAffinity           bool
	GenerateTopologySpreadConstraints bool
}

// tmplEntry maps a boolean condition to a template filename.
type tmplEntry struct {
	enabled    bool
	embedPath  string
	outputName string
}

func readTemplate(path string) (string, error) {
	data, err := templateFS.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("embedded template %s not found: %w", path, err)
	}
	return string(data), nil
}

func Generate(cfg Config, outputDir string) error {
	// Create directories: outputDir/templates
	templatesDir := filepath.Join(outputDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	// renderAndWrite reads a template from embed.FS, renders it with Config, and writes to disk.
	renderAndWrite := func(embedPath, filePath string) error {
		tmplStr, err := readTemplate(embedPath)
		if err != nil {
			return err
		}
		tmpl, err := template.New(filepath.Base(filePath)).Funcs(template.FuncMap{
			"quote": func(s string) string {
				return fmt.Sprintf("%q", s)
			},
		}).Parse(tmplStr)
		if err != nil {
			return fmt.Errorf("failed to parse template for %s: %w", filePath, err)
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, cfg); err != nil {
			return fmt.Errorf("failed to execute template for %s: %w", filePath, err)
		}
		if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", filePath, err)
		}
		return nil
	}

	// writeStatic reads a template from embed.FS and writes it to disk without rendering.
	writeStatic := func(embedPath, filePath string) error {
		data, err := templateFS.ReadFile(embedPath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", embedPath, err)
		}
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filePath, err)
		}
		return nil
	}

	// Write Chart.yaml and values.yaml (rendered via text/template).
	if err := renderAndWrite("templates/chart.yaml", filepath.Join(outputDir, "Chart.yaml")); err != nil {
		return err
	}
	if err := renderAndWrite("templates/values.yaml", filepath.Join(outputDir, "values.yaml")); err != nil {
		return err
	}

	// Write _helpers.tpl (static).
	if err := writeStatic("templates/_helpers.tpl", filepath.Join(templatesDir, "_helpers.tpl")); err != nil {
		return err
	}

	// Table-driven conditional template writing — replaces ~200 lines of repetitive if blocks.
	templates := []tmplEntry{
		{cfg.GenerateDeployment, "templates/deployment.yaml", "deployment.yaml"},
		{cfg.GenerateService, "templates/service.yaml", "service.yaml"},
		{cfg.GenerateIngress, "templates/ingress.yaml", "ingress.yaml"},
		{cfg.GenerateGateway, "templates/gateway.yaml", "gateway.yaml"},
		{cfg.GenerateGateway, "templates/httproute.yaml", "httproute.yaml"},
		{cfg.GenerateConfigMap, "templates/configmap.yaml", "configmap.yaml"},
		{cfg.GenerateSecret, "templates/secret.yaml", "secret.yaml"},
		{cfg.GenerateExternalSecret, "templates/externalsecret.yaml", "externalsecret.yaml"},
		{cfg.GenerateSealedSecret, "templates/sealedsecret.yaml", "sealedsecret.yaml"},
		{cfg.GenerateHPA, "templates/hpa.yaml", "hpa.yaml"},
		{cfg.GenerateVPA, "templates/vpa.yaml", "vpa.yaml"},
		{cfg.GenerateKEDA, "templates/scaledobject.yaml", "scaledobject.yaml"},
		{cfg.GenerateStatefulSet, "templates/statefulset.yaml", "statefulset.yaml"},
		{cfg.GenerateCronJob, "templates/cronjob.yaml", "cronjob.yaml"},
		{cfg.GenerateDaemonSet, "templates/daemonset.yaml", "daemonset.yaml"},
		{cfg.GenerateJob, "templates/job.yaml", "job.yaml"},
		{cfg.GenerateArgoCD, "templates/application.yaml", "application.yaml"},
		{cfg.GenerateArgoCDSet, "templates/applicationset.yaml", "applicationset.yaml"},
		{cfg.GenerateIstio, "templates/virtualservice.yaml", "virtualservice.yaml"},
		{cfg.GeneratePVC, "templates/pvc.yaml", "pvc.yaml"},
		{cfg.GenerateNetworkPolicy, "templates/networkpolicy.yaml", "networkpolicy.yaml"},
		{cfg.GenerateServiceMonitor, "templates/servicemonitor.yaml", "servicemonitor.yaml"},
		{cfg.GeneratePodMonitor, "templates/podmonitor.yaml", "podmonitor.yaml"},
		{cfg.GeneratePrometheusRule, "templates/prometheusrule.yaml", "prometheusrule.yaml"},
		{cfg.GenerateGrafanaDashboard, "templates/grafanadashboard.yaml", "grafanadashboard.yaml"},
		{cfg.GeneratePDB, "templates/pdb.yaml", "pdb.yaml"},
		{cfg.GeneratePriorityClass, "templates/priorityclass.yaml", "priorityclass.yaml"},
		{cfg.GenerateServiceAccount, "templates/serviceaccount.yaml", "serviceaccount.yaml"},
		{cfg.GenerateRole, "templates/role.yaml", "role.yaml"},
		{cfg.GenerateRoleBinding, "templates/rolebinding.yaml", "rolebinding.yaml"},
		{cfg.GenerateClusterRole, "templates/clusterrole.yaml", "clusterrole.yaml"},
		{cfg.GenerateClusterRoleBinding, "templates/clusterrolebinding.yaml", "clusterrolebinding.yaml"},
		{cfg.GenerateFlux, "templates/helmrelease.yaml", "helmrelease.yaml"},
		{cfg.GenerateFlux, "templates/fluxkustomization.yaml", "fluxkustomization.yaml"},
	}

	for _, t := range templates {
		if t.enabled {
			if err := writeStatic(t.embedPath, filepath.Join(templatesDir, t.outputName)); err != nil {
				return err
			}
		}
	}

	return nil
}
