package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ihyamarsdev/kgen/internal/generator"
	"github.com/ihyamarsdev/kgen/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	profileFlag string
	outputFlag  string
	forceFlag   bool
)

// templateFile defines a file that may be generated for a Helm chart.
type templateFile struct {
	name    string
	enabled bool
}

// templateFiles returns the canonical list of template files and their enabled
// status based on the given config.  The order here determines the display
// order in the generated tree output and the edit menu.
func templateFiles(cfg generator.Config) []templateFile {
	return []templateFile{
		{"deployment.yaml", cfg.GenerateDeployment},
		{"service.yaml", cfg.GenerateService},
		{"ingress.yaml", cfg.GenerateIngress},
		{"gateway.yaml", cfg.GenerateGateway},
		{"httproute.yaml", cfg.GenerateGateway},
		{"configmap.yaml", cfg.GenerateConfigMap},
		{"secret.yaml", cfg.GenerateSecret},
		{"externalsecret.yaml", cfg.GenerateExternalSecret},
		{"sealedsecret.yaml", cfg.GenerateSealedSecret},
		{"hpa.yaml", cfg.GenerateHPA},
		{"vpa.yaml", cfg.GenerateVPA},
		{"scaledobject.yaml", cfg.GenerateKEDA},
		{"statefulset.yaml", cfg.GenerateStatefulSet},
		{"cronjob.yaml", cfg.GenerateCronJob},
		{"daemonset.yaml", cfg.GenerateDaemonSet},
		{"job.yaml", cfg.GenerateJob},
		{"application.yaml", cfg.GenerateArgoCD},
		{"applicationset.yaml", cfg.GenerateArgoCDSet},
		{"helmrelease.yaml", cfg.GenerateFlux},
		{"fluxkustomization.yaml", cfg.GenerateFlux},
		{"virtualservice.yaml", cfg.GenerateIstio},
		{"pvc.yaml", cfg.GeneratePVC},
		{"networkpolicy.yaml", cfg.GenerateNetworkPolicy},
		{"servicemonitor.yaml", cfg.GenerateServiceMonitor},
		{"podmonitor.yaml", cfg.GeneratePodMonitor},
		{"prometheusrule.yaml", cfg.GeneratePrometheusRule},
		{"grafanadashboard.yaml", cfg.GenerateGrafanaDashboard},
		{"pdb.yaml", cfg.GeneratePDB},
		{"priorityclass.yaml", cfg.GeneratePriorityClass},
		{"serviceaccount.yaml", cfg.GenerateServiceAccount},
		{"role.yaml", cfg.GenerateRole},
		{"rolebinding.yaml", cfg.GenerateRoleBinding},
		{"clusterrole.yaml", cfg.GenerateClusterRole},
		{"clusterrolebinding.yaml", cfg.GenerateClusterRoleBinding},
	}
}

// printFileTree prints the generated file tree with proper └── for the last item.
func printFileTree(files []templateFile) {
	enabled := make([]string, 0, len(files))
	for _, f := range files {
		if f.enabled {
			enabled = append(enabled, f.name)
		}
	}
	lastIdx := len(enabled) - 1
	for i, name := range enabled {
		if i == lastIdx {
			fmt.Printf("    └── %s\n", name)
		} else {
			fmt.Printf("    ├── %s\n", name)
		}
	}
}

// generatedFilePaths returns only the paths of enabled template files (for the edit menu).
func generatedFilePaths(files []templateFile) []string {
	// Include root-level files so users can edit Chart.yaml and values.yaml.
	var paths []string
	paths = append(paths, "Chart.yaml", "values.yaml")
	for _, f := range files {
		if f.enabled {
			paths = append(paths, "templates/"+f.name)
		}
	}
	return paths
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Helm chart interactively",
	Long:  `Start the interactive terminal wizard to configure and generate a Helm chart.`,
	Run: func(cmd *cobra.Command, args []string) {
		if profileFlag != "dev" && profileFlag != "prod" {
			if profileFlag == "enterprise" {
				fmt.Fprintf(os.Stderr, "Error: Enterprise profile is not supported in MVP v0.1.\n")
			} else {
				fmt.Fprintf(os.Stderr, "Error: Profile '%s' is not supported. Use 'dev' or 'prod'.\n", profileFlag)
			}
			os.Exit(1)
		}

		initialAppName := "my-app"
		if outputFlag != "" {
			base := filepath.Base(outputFlag)
			if base != "." && base != "/" {
				initialAppName = base
			}
		}

		model := tui.InitialModel(profileFlag, initialAppName)
		p := tea.NewProgram(&model)
		m, err := p.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error running wizard: %v\n", err)
			os.Exit(1)
		}

		wizardModel, ok := m.(*tui.WizardModel)
		if !ok {
			fmt.Fprintf(os.Stderr, "Error: Invalid program model returned.\n")
			os.Exit(1)
		}

		if wizardModel.Quitted || !wizardModel.Confirmed {
			fmt.Println("Generation cancelled.")
			os.Exit(0)
		}

		cfg, appName := wizardModel.GetConfig()

		targetDir := outputFlag
		if targetDir == "" {
			if hd := homeDir(); hd != "" {
				targetDir = filepath.Join(hd, "kgen", appName)
			} else {
				targetDir = filepath.Join(".", appName)
			}
		}

		// Check if target directory already exists
		if _, err := os.Stat(targetDir); err == nil {
			if !forceFlag {
				fmt.Fprintf(os.Stderr, "Error: Directory '%s' already exists.\n", targetDir)
				fmt.Println("Use --force (-f) to overwrite, or specify a different output directory with -o.")
				os.Exit(1)
			}
			// Clean the existing directory to avoid stale files.
			if err := os.RemoveAll(targetDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to remove existing directory: %v\n", err)
				os.Exit(1)
			}
		}

		fmt.Printf("\nGenerating Helm chart for '%s' in '%s'...\n", appName, targetDir)

		err = generator.Generate(cfg, targetDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating Helm chart: %v\n", err)
			os.Exit(1)
		}

		titleStyle := tui.TitleStyle.Render
		fmt.Println("\n" + titleStyle(" Helm Chart Generated Successfully! ") + "\n")
		fmt.Printf("Created resources in: %s\n", targetDir)
		fmt.Println("├── Chart.yaml")
		fmt.Println("├── values.yaml")
		fmt.Println("└── templates/")
		fmt.Println("    ├── _helpers.tpl")

		// Use single source of truth for template files
		tmplFiles := templateFiles(cfg)
		printFileTree(tmplFiles)
		fmt.Println()

		// Calculate Production Readiness Score
		if cfg.GenerateDeployment || cfg.GenerateStatefulSet || cfg.GenerateDaemonSet {
			score := 0
			var scoreDetails []string

			addCheck := func(name string, passed bool, points int) {
				if passed {
					score += points
					scoreDetails = append(scoreDetails, fmt.Sprintf("  %s %s", tui.SuccessStyle.Render("✓"), name))
				} else {
					scoreDetails = append(scoreDetails, fmt.Sprintf("  %s %s", tui.ErrorStyle.Render("✗"), name))
				}
			}

			isProdOrEnt := cfg.TemplateQuality == "production" || cfg.TemplateQuality == "enterprise"

			addCheck("Resource Requests", isProdOrEnt, 15)
			addCheck("Resource Limits", isProdOrEnt, 15)
			addCheck("Readiness Probe", isProdOrEnt, 15)
			addCheck("Liveness Probe", isProdOrEnt, 15)
			addCheck("HPA", cfg.GenerateHPA, 10)
			addCheck("PDB", cfg.GeneratePDB, 10)
			addCheck("NetworkPolicy", cfg.GenerateNetworkPolicy, 10)
			addCheck("Topology Spread Constraints", cfg.GenerateTopologySpreadConstraints, 5)
			addCheck("Pod Anti Affinity", cfg.GeneratePodAntiAffinity, 5)

			fmt.Println(tui.HeaderStyle.Render("Production Readiness Score"))
			fmt.Printf("  Score: %d/100\n\n", score)
			for _, detail := range scoreDetails {
				fmt.Println(detail)
			}
			fmt.Println()
		}

		// Offer interactive editing
		generatedFiles := generatedFilePaths(tmplFiles)

		if editor := findEditor(); editor != "" {
			for {
				selModel := tui.InitialSelectorModel(generatedFiles)
				selProg := tea.NewProgram(&selModel)
				mRes, err := selProg.Run()
				if err != nil {
					break
				}
				resModel, ok := mRes.(*tui.SelectorModel)
				if !ok || resModel.Quitted || resModel.SelectedFile == "" {
					break
				}

				filePath := filepath.Join(targetDir, resModel.SelectedFile)
				cmdEdit := exec.Command(editor, filePath)
				cmdEdit.Stdin = os.Stdin
				cmdEdit.Stdout = os.Stdout
				cmdEdit.Stderr = os.Stderr
				if err := cmdEdit.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: editor exited with error: %v\n", err)
			}
			}
		} else {
			fmt.Println("No terminal editor ($EDITOR, nano, vim, vi) found in path. Skipping file edit option.")
		}
	},
}

func init() {
	createCmd.Flags().StringVarP(&profileFlag, "profile", "p", "dev", "Configuration profile to use: dev, prod")
	createCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Output directory path for the Helm chart")
	createCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Overwrite existing output directory")
	rootCmd.AddCommand(createCmd)
}
