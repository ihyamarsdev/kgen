package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"text/template"

	"github.com/ihyamarsdev/kgen/internal/tui"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var previewCmd = &cobra.Command{
	Use:   "preview [chart-directory]",
	Short: "Render and display Helm chart templates in the terminal",
	Long: `Render all Go-template files (Chart.yaml, values.yaml) in a Helm chart
directory and display the result in the terminal — similar to 'helm template'.
Static template files (templates/*.yaml) are printed as-is since they contain
Helm template syntax to be rendered at install time.

If no directory is specified, kgen will attempt to auto-select from ~/.kgen/.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runPreview(args)
	},
}

func init() {
	rootCmd.AddCommand(previewCmd)
}

func runPreview(args []string) {
	var chartDir string

	if len(args) > 0 {
		chartDir = resolveChartPath(args[0])
	} else {
		charts := listAvailableCharts()
		if len(charts) == 0 {
			printErr("Error: No generated Helm charts found in ~/.kgen/.")
			fmt.Println("Run 'kgen create' first to generate a chart.")
			os.Exit(1)
		}
		if len(charts) == 1 {
			chartDir = filepath.Join(chartsDir(), charts[0])
		} else {
			fmt.Println(tui.HeaderStyle.Render("Select a chart to preview:"))
			for i, c := range charts {
				fmt.Printf("  [%d] %s\n", i+1, c)
			}
			chartDir = promptChartChoice(charts)
		}
	}

	fmt.Println(tui.HeaderStyle.Render("Previewing chart: " + filepath.Base(chartDir)))
	fmt.Println()

	// Read all chart files.
	files, err := scanAllChartFiles(chartDir)
	if err != nil {
		printErr("Error scanning chart directory: %v", err)
		os.Exit(1)
	}

	// Sort files for deterministic output.
	var relPaths []string
	for p := range files {
		relPaths = append(relPaths, p)
	}
	sort.Strings(relPaths)

	// Try to parse values.yaml for more accurate rendering.
	renderValues := buildRenderValues(chartDir, files["values.yaml"])

	// Render and display each file.
	first := true
	for _, rel := range relPaths {
		content := files[rel]

		if !first {
			fmt.Println()
			fmt.Println(tui.GrayStyle.Render("──────────────────────────────────────────────────────"))
			fmt.Println()
		}
		first = false

		// Try to render as Go template for Chart.yaml and values.yaml.
		// For templates/* files, print as-is (they contain Helm syntax).
		rendered, err := renderTemplate(rel, content, renderValues)
		if err != nil {
			// If template rendering fails, print raw content with a warning.
			fmt.Println(tui.ErrorStyle.Render("⚠ ") + tui.GrayStyle.Render("Could not render '"+rel+"' as Go template — showing raw content:"))
			fmt.Println()
			fmt.Println(content)
		} else {
			fmt.Println(tui.HeaderStyle.Render("--- # " + rel))
			fmt.Println()
			fmt.Println(rendered)
		}
	}
}

// buildRenderValues constructs a configuration map for template rendering.
// When a real values.yaml is available, it parses it to extract actual values.
// Otherwise, it falls back to sensible defaults.
func buildRenderValues(chartDir, valuesYAML string) map[string]any {
	defaults := map[string]any{
		"AppName":                           filepath.Base(chartDir),
		"Namespace":                         "default",
		"ImageRepository":                   "nginx",
		"ImageTag":                          "latest",
		"Port":                              80,
		"ReplicaCount":                      1,
		"IngressEnabled":                    false,
		"HPAEnabled":                        false,
		"HPAMinReplicas":                    1,
		"HPAMaxReplicas":                    3,
		"ProdProfile":                       false,
		"TemplateQuality":                   "basic",
		"SecretBackend":                     "vault",
		"StorageClass":                      "standard",
		"StorageSize":                       "1Gi",
		"StorageAccessMode":                 "ReadWriteOnce",
		"ServiceAccountName":                "default",
		"RbacLevel":                         "readonly",
		"RbacCustomResources":               []string{},
		"IngressTlsEnabled":                 false,
		"IngressTlsProvider":                "cert-manager",
		"NetworkPolicyPreset":               "defaultdeny",
		"GenerateDeployment":                true,
		"GenerateService":                   true,
		"GenerateIngress":                   false,
		"GenerateGateway":                   false,
		"GenerateConfigMap":                 false,
		"GenerateSecret":                    false,
		"GenerateExternalSecret":            false,
		"GenerateSealedSecret":              false,
		"GenerateHPA":                       false,
		"GenerateServiceMonitor":            false,
		"GeneratePDB":                       false,
		"GenerateVPA":                       false,
		"GenerateKEDA":                      false,
		"GenerateStatefulSet":               false,
		"GenerateCronJob":                   false,
		"GenerateArgoCD":                    false,
		"GenerateIstio":                     false,
		"GeneratePVC":                       false,
		"GenerateNetworkPolicy":             false,
		"GenerateDaemonSet":                 false,
		"GenerateJob":                       false,
		"GenerateServiceAccount":            false,
		"GenerateRbac":                      false,
		"GenerateRole":                      false,
		"GenerateRoleBinding":               false,
		"GenerateClusterRole":               false,
		"GenerateClusterRoleBinding":        false,
		"GeneratePriorityClass":             false,
		"GeneratePodMonitor":                false,
		"GeneratePrometheusRule":            false,
		"GenerateGrafanaDashboard":          false,
		"GenerateArgoCDSet":                 false,
		"GenerateFlux":                      false,
		"GeneratePodAntiAffinity":           false,
		"GenerateTopologySpreadConstraints": false,
	}

	if valuesYAML != "" {
		var parsed map[string]any
		if err := yaml.Unmarshal([]byte(valuesYAML), &parsed); err == nil {
			// Merge parsed values into defaults.
			for k, v := range parsed {
				defaults[k] = v
			}
			// Infer Generate* flags from what's present in values.yaml.
			if _, ok := parsed["ingress"]; ok {
				defaults["GenerateIngress"] = true
			}
			if _, ok := parsed["autoscaling"]; ok {
				defaults["GenerateHPA"] = true
			}
			if _, ok := parsed["serviceMonitor"]; ok {
				defaults["GenerateServiceMonitor"] = true
			}
			if _, ok := parsed["networkPolicy"]; ok {
				defaults["GenerateNetworkPolicy"] = true
			}
			if _, ok := parsed["podDisruptionBudget"]; ok {
				defaults["GeneratePDB"] = true
			}
		}
	}

	return defaults
}

// renderTemplate attempts to render a file as a Go text/template.
//
// For Chart.yaml and values.yaml, kgen uses text/template at generation time
// (with the Config struct as data). For templates/* files, the content is
// Helm template syntax — so we just return it as-is.
//
// When the user has an existing chart on disk (not generated by this run),
// we try to render it anyway but fall back to raw content on failure.
func renderTemplate(rel, content string, values map[string]any) (string, error) {
	// Only attempt template rendering for Chart.yaml and values.yaml.
	base := filepath.Base(rel)
	if base != "Chart.yaml" && base != "values.yaml" {
		return content, nil
	}

	// Try to parse and execute as Go template.
	tmpl, err := template.New(rel).Funcs(template.FuncMap{
		"quote": func(s string) string { return fmt.Sprintf("%q", s) },
	}).Parse(content)
	if err != nil {
		return "", err
	}

	// Use the passed values map for rendering (from buildRenderValues).
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, values); err != nil {
		return "", err
	}

	return buf.String(), nil
}
