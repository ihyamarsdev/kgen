package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"kgen/internal/generator"
	"kgen/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	profileFlag string
	outputFlag  string
)

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
			if homeDir, err := os.UserHomeDir(); err == nil {
				targetDir = filepath.Join(homeDir, "kgen", appName)
			} else {
				targetDir = filepath.Join(".", appName)
			}
		}

		fmt.Printf("\nGenerating Helm chart for '%s' in '%s'...\n", appName, targetDir)

		err = generator.Generate(cfg, targetDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating Helm chart: %v\n", err)
			os.Exit(1)
		}

		successStyle := tui.SuccessStyle.Render
		titleStyle := tui.TitleStyle.Render
		fmt.Println("\n" + titleStyle(" Helm Chart Generated Successfully! ") + "\n")
		fmt.Printf("Created resources in: %s\n", targetDir)
		fmt.Println("├── Chart.yaml")
		fmt.Println("├── values.yaml")
		fmt.Println("└── templates/")
		fmt.Println("    ├── _helpers.tpl")

		printFile := func(name string, exists bool) {
			if exists {
				fmt.Printf("    ├── %s\n", name)
			}
		}

		printFile("deployment.yaml", cfg.GenerateDeployment)
		printFile("service.yaml", cfg.GenerateService)
		printFile("ingress.yaml", cfg.GenerateIngress)
		printFile("gateway.yaml", cfg.GenerateGateway)
		printFile("httproute.yaml", cfg.GenerateGateway)
		printFile("configmap.yaml", cfg.GenerateConfigMap)
		printFile("secret.yaml", cfg.GenerateSecret)
		printFile("externalsecret.yaml", cfg.GenerateExternalSecret)
		printFile("sealedsecret.yaml", cfg.GenerateSealedSecret)
		printFile("hpa.yaml", cfg.GenerateHPA)
		printFile("vpa.yaml", cfg.GenerateVPA)
		printFile("scaledobject.yaml", cfg.GenerateKEDA)
		printFile("statefulset.yaml", cfg.GenerateStatefulSet)
		printFile("cronjob.yaml", cfg.GenerateCronJob)
		printFile("application.yaml", cfg.GenerateArgoCD)
		printFile("applicationset.yaml", cfg.GenerateArgoCDSet)
		printFile("helmrelease.yaml", cfg.GenerateFlux)
		printFile("fluxkustomization.yaml", cfg.GenerateFlux)
		printFile("virtualservice.yaml", cfg.GenerateIstio)
		printFile("pvc.yaml", cfg.GeneratePVC)
		printFile("networkpolicy.yaml", cfg.GenerateNetworkPolicy)
		printFile("servicemonitor.yaml", cfg.GenerateServiceMonitor)
		printFile("podmonitor.yaml", cfg.GeneratePodMonitor)
		printFile("prometheusrule.yaml", cfg.GeneratePrometheusRule)
		printFile("grafanadashboard.yaml", cfg.GenerateGrafanaDashboard)
		printFile("pdb.yaml", cfg.GeneratePDB)
		printFile("priorityclass.yaml", cfg.GeneratePriorityClass)
		printFile("serviceaccount.yaml", cfg.GenerateServiceAccount)
		printFile("role.yaml", cfg.GenerateRole)
		printFile("rolebinding.yaml", cfg.GenerateRoleBinding)
		printFile("clusterrole.yaml", cfg.GenerateClusterRole)
		printFile("clusterrolebinding.yaml", cfg.GenerateClusterRoleBinding)
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

		fmt.Printf("Ready to deploy! You can validate it by running:\n  kgen validate %s\n\n", targetDir)

		// Compile the generated files list for editing
		var generatedFiles []string
		addGenFile := func(name string, exists bool) {
			if exists {
				generatedFiles = append(generatedFiles, name)
			}
		}

		// Always generated files
		generatedFiles = append(generatedFiles, "Chart.yaml", "values.yaml", "templates/_helpers.tpl")

		addGenFile("templates/deployment.yaml", cfg.GenerateDeployment)
		addGenFile("templates/service.yaml", cfg.GenerateService)
		addGenFile("templates/ingress.yaml", cfg.GenerateIngress)
		addGenFile("templates/gateway.yaml", cfg.GenerateGateway)
		addGenFile("templates/httproute.yaml", cfg.GenerateGateway)
		addGenFile("templates/configmap.yaml", cfg.GenerateConfigMap)
		addGenFile("templates/secret.yaml", cfg.GenerateSecret)
		addGenFile("templates/externalsecret.yaml", cfg.GenerateExternalSecret)
		addGenFile("templates/sealedsecret.yaml", cfg.GenerateSealedSecret)
		addGenFile("templates/hpa.yaml", cfg.GenerateHPA)
		addGenFile("templates/vpa.yaml", cfg.GenerateVPA)
		addGenFile("templates/scaledobject.yaml", cfg.GenerateKEDA)
		addGenFile("templates/statefulset.yaml", cfg.GenerateStatefulSet)
		addGenFile("templates/cronjob.yaml", cfg.GenerateCronJob)
		addGenFile("templates/application.yaml", cfg.GenerateArgoCD)
		addGenFile("templates/applicationset.yaml", cfg.GenerateArgoCDSet)
		addGenFile("templates/helmrelease.yaml", cfg.GenerateFlux)
		addGenFile("templates/fluxkustomization.yaml", cfg.GenerateFlux)
		addGenFile("templates/virtualservice.yaml", cfg.GenerateIstio)
		addGenFile("templates/pvc.yaml", cfg.GeneratePVC)
		addGenFile("templates/networkpolicy.yaml", cfg.GenerateNetworkPolicy)
		addGenFile("templates/servicemonitor.yaml", cfg.GenerateServiceMonitor)
		addGenFile("templates/podmonitor.yaml", cfg.GeneratePodMonitor)
		addGenFile("templates/prometheusrule.yaml", cfg.GeneratePrometheusRule)
		addGenFile("templates/grafanadashboard.yaml", cfg.GenerateGrafanaDashboard)
		addGenFile("templates/pdb.yaml", cfg.GeneratePDB)
		addGenFile("templates/priorityclass.yaml", cfg.GeneratePriorityClass)
		addGenFile("templates/serviceaccount.yaml", cfg.GenerateServiceAccount)
		addGenFile("templates/role.yaml", cfg.GenerateRole)
		addGenFile("templates/rolebinding.yaml", cfg.GenerateRoleBinding)
		addGenFile("templates/clusterrole.yaml", cfg.GenerateClusterRole)
		addGenFile("templates/clusterrolebinding.yaml", cfg.GenerateClusterRoleBinding)

		// Interactive File Editing Loop
		editor := os.Getenv("EDITOR")
		if editor == "" {
			// Find standard editors: nano, vim, vi
			for _, e := range []string{"nano", "vim", "vi"} {
				if _, err := exec.LookPath(e); err == nil {
					editor = e
					break
				}
			}
		}

		if editor != "" {
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

				// Launch Editor
				filePath := filepath.Join(targetDir, resModel.SelectedFile)
				cmdEdit := exec.Command(editor, filePath)
				cmdEdit.Stdin = os.Stdin
				cmdEdit.Stdout = os.Stdout
				cmdEdit.Stderr = os.Stderr
				_ = cmdEdit.Run()
			}
		} else {
			fmt.Println("No terminal editor ($EDITOR, nano, vim, vi) found in path. Skipping file edit option.")
		}

		_ = successStyle // silence compiler if unused
	},
}

func init() {
	createCmd.Flags().StringVarP(&profileFlag, "profile", "p", "dev", "Configuration profile to use: dev, prod")
	createCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Output directory path for the Helm chart")
	rootCmd.AddCommand(createCmd)
}
