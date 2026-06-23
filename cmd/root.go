package cmd

import (
	"fmt"
	"os"

	"github.com/ihyamarsdev/kgen/internal/tui"
	"github.com/ihyamarsdev/kgen/internal/version"

	"github.com/spf13/cobra"
)

// versionFlag is set via the --version flag on the root command.
var versionFlag bool

var rootCmd = &cobra.Command{
	Use:   "kgen",
	Short: "KGen is an interactive CLI tool for generating Helm charts",
	Long: `KGen (Kubernetes & Helm Generator) helps developers, DevOps, and platform teams
generate standardized, production-ready Helm charts via a beautiful terminal TUI.`,
	Run: func(cmd *cobra.Command, args []string) {
		// `kgen --version` prints the version and exits.
		if versionFlag {
			fmt.Printf("kgen version %s\n", version.Version)
			return
		}
		// No subcommand and no --version: show help.
		_ = cmd.Help()
	},
}

func init() {
	// --version flag prints the current KGen version.
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "V", false, "print the kgen version and exit")

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

// Execute applies the styled help template to all subcommands before running.
func Execute() {
	// Propagate the help template to all subcommands.
	applyHelpToAll(rootCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// applyHelpToAll recursively applies the styled help template to every command.
func applyHelpToAll(cmd *cobra.Command) {
	cmd.SetHelpTemplate(rootCmd.HelpTemplate())
	for _, sub := range cmd.Commands() {
		applyHelpToAll(sub)
	}
}
