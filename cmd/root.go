package cmd

import (
	"fmt"
	"os"

	"kgen/internal/tui"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kgen",
	Short: "KGen is an interactive CLI tool for generating Helm charts",
	Long: `KGen (Kubernetes & Helm Generator) helps developers, DevOps, and platform teams
generate standardized, production-ready Helm charts via a beautiful terminal TUI.`,
}

func init() {
	// Customize the Cobra help template to be beautiful
	cobra.AddTemplateFunc("styleHeading", func(s string) string {
		return tui.HeaderStyle.Render(s)
	})
	cobra.AddTemplateFunc("styleCommand", func(name string) string {
		return tui.ActiveInputStyle.Render(name)
	})
	cobra.AddTemplateFunc("styleDescription", func(desc string) string {
		return tui.GrayStyle.Render(desc)
	})

	rootCmd.SetHelpTemplate(`{{styleHeading "Description:"}}
  {{if .Long}}{{.Long}}{{else}}{{.Short}}{{end}}

{{styleHeading "Usage:"}}
  {{if .Runnable}}{{styleCommand .UseLine}}{{end}}{{if .HasAvailableSubCommands}}{{styleCommand .CommandPath}} [command]{{end}}

{{if .HasAvailableSubCommands}}{{styleHeading "Available Commands:"}}
{{range .Commands}}{{if (and .IsAvailableCommand (not (eq .Name "help")))}}  {{styleCommand (rpad .Name 12)}} {{styleDescription .Short}}
{{end}}{{end}}
{{end}}{{if .HasAvailableLocalFlags}}{{styleHeading "Flags:"}}
{{styleDescription (.LocalFlags.FlagUsages | trimTrailingWhitespaces)}}

{{end}}{{if .HasAvailableInheritedFlags}}{{styleHeading "Global Flags:"}}
{{styleDescription (.InheritedFlags.FlagUsages | trimTrailingWhitespaces)}}

{{end}}Use "{{.CommandPath}} [command] --help" for more information about a command.
`)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
