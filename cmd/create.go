package cmd

import (
	"fmt"
	"os"
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
		if cfg.GenerateDeployment {
			fmt.Println("    ├── deployment.yaml")
		}
		if cfg.GenerateService {
			fmt.Println("    ├── service.yaml")
		}
		if cfg.GenerateIngress {
			fmt.Println("    ├── ingress.yaml")
		}
		if cfg.GenerateHPA {
			fmt.Println("    └── hpa.yaml")
		}
		fmt.Println()
		fmt.Printf("Ready to deploy! You can validate it by running:\n  kgen validate %s\n\n", targetDir)
		_ = successStyle // silence compiler if unused
	},
}

func init() {
	createCmd.Flags().StringVarP(&profileFlag, "profile", "p", "dev", "Configuration profile to use: dev, prod")
	createCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Output directory path for the Helm chart")
	rootCmd.AddCommand(createCmd)
}
