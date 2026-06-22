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
	// Resources to generate
	GenerateDeployment bool
	GenerateService    bool
	GenerateIngress    bool
	GenerateHPA        bool
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

	if cfg.GenerateHPA {
		if err := os.WriteFile(filepath.Join(templatesDir, "hpa.yaml"), []byte(HPATemplate), 0644); err != nil {
			return fmt.Errorf("failed to write hpa.yaml: %w", err)
		}
	}

	return nil
}
