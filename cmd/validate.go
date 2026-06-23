package cmd

import (
	"fmt"
	"os"

	"github.com/ihyamarsdev/kgen/internal/tui"
	"github.com/ihyamarsdev/kgen/internal/validator"

	"github.com/spf13/cobra"
)

var validateStrict bool

var validateCmd = &cobra.Command{
	Use:   "validate [path]",
	Short: "Validate Kubernetes best practices in a directory",
	Long: `Scan Helm charts or Kubernetes YAML files in the specified directory
for standard best practices (Limits, Requests, Probes, Security Context).

With --strict, the command exits with code 1 when any warnings are found,
making it suitable for CI/CD quality gates.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		fmt.Printf("Validating resources in '%s'...\n\n", path)

		results, err := validator.ValidateDir(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error performing validation: %v\n", err)
			os.Exit(1)
		}

		passCount := 0
		warnCount := 0

		for _, res := range results {
			var statusStr string
			if res.Status == "PASS" {
				statusStr = tui.SuccessStyle.Render("PASS")
				passCount++
			} else {
				statusStr = tui.ErrorStyle.Render("WARN")
				warnCount++
			}
			fmt.Printf("%s: %s (%s)\n", statusStr, res.Message, res.Check)
		}

		fmt.Println()
		if warnCount > 0 {
			fmt.Printf("Validation finished with %d warnings and %d passes.\n", warnCount, passCount)
			fmt.Println("We recommend fixing warnings before deploying to a production cluster.")
			if validateStrict {
				os.Exit(1)
			}
		} else {
			fmt.Println(tui.SuccessStyle.Render("All checks passed! Your configuration adheres to Kubernetes best practices."))
		}
	},
}

func init() {
	validateCmd.Flags().BoolVarP(&validateStrict, "strict", "s", false, "exit with code 1 when any warnings are found (useful for CI/CD)")
	rootCmd.AddCommand(validateCmd)
}
