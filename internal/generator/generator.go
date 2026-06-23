package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

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
	StorageClass         string
	StorageSize          string
	StorageAccessMode    string
	ServiceAccountName   string
	RbacLevel            string
	RbacCustomResources  []string
	IngressTlsEnabled    bool
	IngressTlsProvider   string
	NetworkPolicyPreset  string

	// Resources to generate
	GenerateDeployment        bool
	GenerateService           bool
	GenerateIngress           bool
	GenerateGateway           bool
	GenerateConfigMap         bool
	GenerateSecret            bool
	GenerateExternalSecret    bool
	GenerateSealedSecret      bool
	GenerateHPA               bool
	GenerateServiceMonitor    bool
	GeneratePDB               bool
	GenerateVPA               bool
	GenerateKEDA              bool
	GenerateStatefulSet       bool
	GenerateCronJob           bool
	GenerateArgoCD            bool
	GenerateIstio             bool
	GeneratePVC               bool
	GenerateNetworkPolicy     bool
	GenerateDaemonSet         bool
	GenerateJob               bool
	GenerateServiceAccount    bool
	GenerateRbac              bool // Role and RoleBinding
	GenerateRole              bool
	GenerateRoleBinding       bool
	GenerateClusterRole       bool
	GenerateClusterRoleBinding bool
	GeneratePriorityClass     bool
	GeneratePodMonitor        bool
	GeneratePrometheusRule    bool
	GenerateGrafanaDashboard  bool
	GenerateArgoCDSet         bool
	GenerateFlux              bool
	GeneratePodAntiAffinity   bool
	GenerateTopologySpreadConstraints bool
}

// tmplEntry maps a boolean condition to a template content string.
type tmplEntry struct {
	enabled  bool
	template string
}

func Generate(cfg Config, outputDir string) error {
	// Create directories: outputDir/templates
	templatesDir := filepath.Join(outputDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	// renderAndWrite renders a template string with the Config and writes it to a file.
	renderAndWrite := func(tmplStr, filePath string) error {
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

	// Write Chart.yaml and values.yaml (rendered via text/template).
	if err := renderAndWrite(ChartTemplate, filepath.Join(outputDir, "Chart.yaml")); err != nil {
		return err
	}
	if err := renderAndWrite(ValuesTemplate, filepath.Join(outputDir, "values.yaml")); err != nil {
		return err
	}

	// Write _helpers.tpl (static).
	if err := os.WriteFile(filepath.Join(templatesDir, "_helpers.tpl"), []byte(HelpersTemplate), 0644); err != nil {
		return fmt.Errorf("failed to write _helpers.tpl: %w", err)
	}

	// Table-driven conditional template writing — replaces ~200 lines of repetitive if blocks.
	templates := []struct {
		name    string
		enabled bool
		content string
	}{
		{"deployment.yaml", cfg.GenerateDeployment, DeploymentTemplate},
		{"service.yaml", cfg.GenerateService, ServiceTemplate},
		{"ingress.yaml", cfg.GenerateIngress, IngressTemplate},
		{"gateway.yaml", cfg.GenerateGateway, GatewayTemplate},
		{"httproute.yaml", cfg.GenerateGateway, HTTPRouteTemplate},
		{"configmap.yaml", cfg.GenerateConfigMap, ConfigMapTemplate},
		{"secret.yaml", cfg.GenerateSecret, SecretTemplate},
		{"externalsecret.yaml", cfg.GenerateExternalSecret, ExternalSecretTemplate},
		{"sealedsecret.yaml", cfg.GenerateSealedSecret, SealedSecretTemplate},
		{"hpa.yaml", cfg.GenerateHPA, HPATemplate},
		{"vpa.yaml", cfg.GenerateVPA, VPATemplate},
		{"scaledobject.yaml", cfg.GenerateKEDA, ScaledObjectTemplate},
		{"statefulset.yaml", cfg.GenerateStatefulSet, StatefulSetTemplate},
		{"cronjob.yaml", cfg.GenerateCronJob, CronJobTemplate},
		{"daemonset.yaml", cfg.GenerateDaemonSet, DaemonSetTemplate},
		{"job.yaml", cfg.GenerateJob, JobTemplate},
		{"application.yaml", cfg.GenerateArgoCD, ArgoApplicationTemplate},
		{"applicationset.yaml", cfg.GenerateArgoCDSet, ArgoApplicationSetTemplate},
		{"virtualservice.yaml", cfg.GenerateIstio, IstioVirtualServiceTemplate},
		{"pvc.yaml", cfg.GeneratePVC, PVCTemplate},
		{"networkpolicy.yaml", cfg.GenerateNetworkPolicy, NetworkPolicyTemplate},
		{"servicemonitor.yaml", cfg.GenerateServiceMonitor, ServiceMonitorTemplate},
		{"podmonitor.yaml", cfg.GeneratePodMonitor, PodMonitorTemplate},
		{"prometheusrule.yaml", cfg.GeneratePrometheusRule, PrometheusRuleTemplate},
		{"grafanadashboard.yaml", cfg.GenerateGrafanaDashboard, GrafanaDashboardTemplate},
		{"pdb.yaml", cfg.GeneratePDB, PdbTemplate},
		{"priorityclass.yaml", cfg.GeneratePriorityClass, PriorityClassTemplate},
		{"serviceaccount.yaml", cfg.GenerateServiceAccount, ServiceAccountTemplate},
		{"role.yaml", cfg.GenerateRole, RoleTemplate},
		{"rolebinding.yaml", cfg.GenerateRoleBinding, RoleBindingTemplate},
		{"clusterrole.yaml", cfg.GenerateClusterRole, ClusterRoleTemplate},
		{"clusterrolebinding.yaml", cfg.GenerateClusterRoleBinding, ClusterRoleBindingTemplate},
		{"helmrelease.yaml", cfg.GenerateFlux, FluxHelmReleaseTemplate},
		{"fluxkustomization.yaml", cfg.GenerateFlux, FluxKustomizationTemplate},
	}

	for _, t := range templates {
		if t.enabled {
			if err := os.WriteFile(filepath.Join(templatesDir, t.name), []byte(t.content), 0644); err != nil {
				return fmt.Errorf("failed to write %s: %w", t.name, err)
			}
		}
	}

	return nil
}
