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

	// Resources to generate
	GenerateDeployment     bool
	GenerateService        bool
	GenerateIngress        bool
	GenerateGateway        bool
	GenerateConfigMap      bool
	GenerateExternalSecret bool
	GenerateHPA            bool
	GenerateServiceMonitor bool
	GeneratePDB            bool
	GenerateVPA            bool
	GenerateKEDA           bool
	GenerateStatefulSet    bool
	GenerateCronJob        bool
	GenerateArgoCD         bool
	GenerateIstio          bool
	GeneratePVC            bool
	GenerateNetworkPolicy  bool
}

func Generate(cfg Config, outputDir string) error {
	// Create directories:
	// outputDir/templates
	templatesDir := filepath.Join(outputDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	// Helper function to render a template and write to file
	renderAndWrite := func(tmplStr, filePath string) error {
		tmpl, err := template.New(filepath.Base(filePath)).Parse(tmplStr)
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

	// Write Chart.yaml
	if err := renderAndWrite(ChartTemplate, filepath.Join(outputDir, "Chart.yaml")); err != nil {
		return err
	}

	// Write values.yaml
	if err := renderAndWrite(ValuesTemplate, filepath.Join(outputDir, "values.yaml")); err != nil {
		return err
	}

	// Write _helpers.tpl
	if err := os.WriteFile(filepath.Join(templatesDir, "_helpers.tpl"), []byte(HelpersTemplate), 0644); err != nil {
		return fmt.Errorf("failed to write _helpers.tpl: %w", err)
	}

	// Write conditional template files
	if cfg.GenerateDeployment {
		if err := os.WriteFile(filepath.Join(templatesDir, "deployment.yaml"), []byte(DeploymentTemplate), 0644); err != nil {
			return fmt.Errorf("failed to write deployment.yaml: %w", err)
		}
	}

	if cfg.GenerateService {
		if err := os.WriteFile(filepath.Join(templatesDir, "service.yaml"), []byte(ServiceTemplate), 0644); err != nil {
			return fmt.Errorf("failed to write service.yaml: %w", err)
		}
	}

	if cfg.GenerateIngress {
		if err := os.WriteFile(filepath.Join(templatesDir, "ingress.yaml"), []byte(IngressTemplate), 0644); err != nil {
			return fmt.Errorf("failed to write ingress.yaml: %w", err)
		}
	}

	if cfg.GenerateGateway {
		if err := os.WriteFile(filepath.Join(templatesDir, "gateway.yaml"), []byte(GatewayTemplate), 0644); err != nil {
			return fmt.Errorf("failed to write gateway.yaml: %w", err)
		}
		if err := os.WriteFile(filepath.Join(templatesDir, "httproute.yaml"), []byte(HTTPRouteTemplate), 0644); err != nil {
			return fmt.Errorf("failed to write httproute.yaml: %w", err)
		}
	}

	if cfg.GenerateConfigMap {
		if err := os.WriteFile(filepath.Join(templatesDir, "configmap.yaml"), []byte(ConfigMapTemplate), 0644); err != nil {
			return fmt.Errorf("failed to write configmap.yaml: %w", err)
		}
	}

	if cfg.GenerateExternalSecret {
		if err := os.WriteFile(filepath.Join(templatesDir, "externalsecret.yaml"), []byte(ExternalSecretTemplate), 0644); err != nil {
			return fmt.Errorf("failed to write externalsecret.yaml: %w", err)
		}
	}

	if cfg.GenerateHPA {
		if err := os.WriteFile(filepath.Join(templatesDir, "hpa.yaml"), []byte(HPATemplate), 0644); err != nil {
			return fmt.Errorf("failed to write hpa.yaml: %w", err)
		}
	}

	if cfg.GenerateVPA {
		if err := os.WriteFile(filepath.Join(templatesDir, "vpa.yaml"), []byte(VPATemplate), 0644); err != nil {
			return fmt.Errorf("failed to write vpa.yaml: %w", err)
		}
	}

	if cfg.GenerateKEDA {
		if err := os.WriteFile(filepath.Join(templatesDir, "scaledobject.yaml"), []byte(ScaledObjectTemplate), 0644); err != nil {
			return fmt.Errorf("failed to write scaledobject.yaml: %w", err)
		}
	}

	if cfg.GenerateStatefulSet {
		if err := os.WriteFile(filepath.Join(templatesDir, "statefulset.yaml"), []byte(StatefulSetTemplate), 0644); err != nil {
			return fmt.Errorf("failed to write statefulset.yaml: %w", err)
		}
	}

	if cfg.GenerateCronJob {
		if err := os.WriteFile(filepath.Join(templatesDir, "cronjob.yaml"), []byte(CronJobTemplate), 0644); err != nil {
			return fmt.Errorf("failed to write cronjob.yaml: %w", err)
		}
	}

	if cfg.GenerateArgoCD {
		if err := os.WriteFile(filepath.Join(templatesDir, "application.yaml"), []byte(ArgoApplicationTemplate), 0644); err != nil {
			return fmt.Errorf("failed to write application.yaml: %w", err)
		}
	}

	if cfg.GenerateIstio {
		if err := os.WriteFile(filepath.Join(templatesDir, "virtualservice.yaml"), []byte(IstioVirtualServiceTemplate), 0644); err != nil {
			return fmt.Errorf("failed to write virtualservice.yaml: %w", err)
		}
	}

	if cfg.GeneratePVC {
		if err := os.WriteFile(filepath.Join(templatesDir, "pvc.yaml"), []byte(PVCTemplate), 0644); err != nil {
			return fmt.Errorf("failed to write pvc.yaml: %w", err)
		}
	}

	if cfg.GenerateNetworkPolicy {
		if err := os.WriteFile(filepath.Join(templatesDir, "networkpolicy.yaml"), []byte(NetworkPolicyTemplate), 0644); err != nil {
			return fmt.Errorf("failed to write networkpolicy.yaml: %w", err)
		}
	}

	if cfg.GenerateServiceMonitor {
		if err := os.WriteFile(filepath.Join(templatesDir, "servicemonitor.yaml"), []byte(ServiceMonitorTemplate), 0644); err != nil {
			return fmt.Errorf("failed to write servicemonitor.yaml: %w", err)
		}
	}

	if cfg.GeneratePDB {
		if err := os.WriteFile(filepath.Join(templatesDir, "pdb.yaml"), []byte(PdbTemplate), 0644); err != nil {
			return fmt.Errorf("failed to write pdb.yaml: %w", err)
		}
	}

	return nil
}
