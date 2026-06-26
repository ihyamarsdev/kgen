package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ihyamarsdev/kgen/internal/tui"

	"strings"

	"github.com/spf13/cobra"
)

// Deploy flags
var (
	deployRelease   string
	deployNamespace string
	deployValues    []string
	deploySet       []string
	deployDryRun    bool
	deployTimeout   string
	deployWait      bool
	deployYes       bool
)

var deployCmd = &cobra.Command{
	Use:   "deploy [chart-directory]",
	Short: "Deploy a Helm chart to a Kubernetes cluster",
	Long: `Install or upgrade a Helm chart on your Kubernetes cluster.

If the release already exists in the target namespace, it will be upgraded.
Otherwise, a fresh install is performed.

If no chart directory is provided, kgen will auto-select from ~/.kgen/.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runDeploy(args)
	},
}

var undeployCmd = &cobra.Command{
	Use:   "undeploy [chart-directory]",
	Short: "Uninstall a Helm release from a Kubernetes cluster",
	Long: `Uninstall (delete) a Helm release from your Kubernetes cluster.

The release name is derived from the chart directory name. Override with --release.

If no chart directory is provided, kgen will auto-select from ~/.kgen/.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runUndeploy(args)
	},
}

var statusCmd = &cobra.Command{
	Use:   "status [chart-directory]",
	Short: "Show the status of a deployed Helm release",
	Long: `Display the status of a Helm release in your Kubernetes cluster.

The release name is derived from the chart directory name. Override with --release.

If no chart directory is provided, kgen will auto-select from ~/.kgen/.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runStatus(args)
	},
}

func init() {
	// Deploy flags
	deployCmd.Flags().StringVarP(&deployRelease, "release", "r", "", "Helm release name (defaults to chart directory name)")
	deployCmd.Flags().StringVarP(&deployNamespace, "namespace", "n", "default", "Kubernetes namespace")
	deployCmd.Flags().StringArrayVarP(&deployValues, "values", "f", nil, "Override values from a file (can specify multiple)")
	deployCmd.Flags().StringArrayVar(&deploySet, "set", nil, "Override values (e.g. --set image.tag=v2)")
	deployCmd.Flags().BoolVarP(&deployDryRun, "dry-run", "d", false, "Simulate the deployment without applying changes")
	deployCmd.Flags().StringVarP(&deployTimeout, "timeout", "t", "5m", "Time to wait for deployment (e.g. 5m, 10m)")
	deployCmd.Flags().BoolVarP(&deployWait, "wait", "w", false, "Wait until all resources are ready")

	// Undeploy flags
	undeployCmd.Flags().StringVarP(&deployRelease, "release", "r", "", "Helm release name (defaults to chart directory name)")
	undeployCmd.Flags().StringVarP(&deployNamespace, "namespace", "n", "default", "Kubernetes namespace")
	undeployCmd.Flags().BoolVarP(&deployYes, "yes", "y", false, "Skip confirmation prompt")
	undeployCmd.Flags().BoolVarP(&deployDryRun, "dry-run", "d", false, "Simulate the uninstall without applying changes")

	// Status flags
	statusCmd.Flags().StringVarP(&deployRelease, "release", "r", "", "Helm release name (defaults to chart directory name)")
	statusCmd.Flags().StringVarP(&deployNamespace, "namespace", "n", "default", "Kubernetes namespace")

	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(undeployCmd)
	rootCmd.AddCommand(statusCmd)
}

// resolveDeployChart resolves the chart directory for deploy/undeploy/status commands.
func resolveDeployChart(args []string) string {
	if len(args) > 0 {
		return resolveChartPath(args[0])
	}

	charts := listAvailableCharts()
	if len(charts) == 0 {
		printErr("Error: No generated Helm charts found in ~/.kgen/.")
		fmt.Println("Run 'kgen create' first to generate a chart.")
		os.Exit(1)
	}

	if len(charts) == 1 {
		return filepath.Join(chartsDir(), charts[0])
	}

	fmt.Println(tui.HeaderStyle.Render("Select a chart to deploy:"))
	for i, c := range charts {
		fmt.Printf("  [%d] %s\n", i+1, c)
	}
	return promptChartChoice(charts)
}

// requireHelm checks that helm is available and exits with a helpful message if not.
func requireHelm() {
	if findHelm() == "" {
		printErr("Error: 'helm' binary not found in PATH.")
		fmt.Println("Please install Helm: https://helm.sh/docs/intro/install/")
		os.Exit(1)
	}
}

func runDeploy(args []string) {
	requireHelm()

	chartDir := resolveDeployChart(args)
	release := deployRelease
	if release == "" {
		release = releaseNameFromChart(chartDir)
	}

	// Read namespace from chart values if user didn't specify -n
	chartNS := readChartNamespace(chartDir)
	ns := deployNamespace
	if ns == "default" && chartNS != "default" {
		ns = chartNS
	}

	fmt.Println(tui.HeaderStyle.Render("Deploying Helm Chart"))
	fmt.Printf("  Chart:     %s\n", chartDir)
	fmt.Printf("  Release:   %s\n", release)
	fmt.Printf("  Namespace: %s\n", ns)

	if deployDryRun {
		fmt.Println(tui.GrayStyle.Render("  Mode:      dry-run"))
	}
	fmt.Println()

	// Build helm command arguments.
	helmArgs := buildDeployArgs(release, chartDir, ns)

	// Check if release exists to decide between install and upgrade.
	action := "install"
	if !deployDryRun && helmReleaseExists(release, ns) {
		action = "upgrade"
		helmArgs = buildUpgradeArgs(release, chartDir, ns)
		fmt.Printf("Release '%s' already exists in namespace '%s' — performing %s.\n", release, ns, action)
		fmt.Println()
	}

	fmt.Printf("Running: helm %s\n\n", strings.Join(helmArgs, " "))

	if err := helmRun(helmArgs...); err != nil {
		printErr("Helm %s failed: %v", action, err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println(tui.SuccessStyle.Render(fmt.Sprintf("Chart %s successfully as release '%s' in namespace '%s'.", action, release, ns)))
}

func buildDeployArgs(release, chartDir, ns string) []string {
	args := []string{"install", release, chartDir, "--namespace", ns, "--create-namespace"}
	if deployDryRun {
		args = append(args, "--dry-run")
	}
	args = append(args, "--timeout", deployTimeout)
	if deployWait {
		args = append(args, "--wait")
	}
	for _, v := range deployValues {
		args = append(args, "--values", v)
	}
	for _, s := range deploySet {
		args = append(args, "--set", s)
	}
	return args
}

func buildUpgradeArgs(release, chartDir, ns string) []string {
	args := []string{"upgrade", release, chartDir, "--namespace", ns}
	if deployDryRun {
		args = append(args, "--dry-run")
	}
	args = append(args, "--timeout", deployTimeout)
	if deployWait {
		args = append(args, "--wait")
	}
	for _, v := range deployValues {
		args = append(args, "--values", v)
	}
	for _, s := range deploySet {
		args = append(args, "--set", s)
	}
	return args
}

func runUndeploy(args []string) {
	requireHelm()

	chartDir := resolveDeployChart(args)
	release := deployRelease
	if release == "" {
		release = releaseNameFromChart(chartDir)
	}

	// Read namespace from chart values if user didn't specify -n
	chartNS := readChartNamespace(chartDir)
	ns := deployNamespace
	if ns == "default" && chartNS != "default" {
		ns = chartNS
	}

	fmt.Println(tui.HeaderStyle.Render("Uninstalling Helm Release"))
	fmt.Printf("  Release:   %s\n", release)
	fmt.Printf("  Namespace: %s\n", ns)
	fmt.Println()

	// Check if the release actually exists.
	if !helmReleaseExists(release, ns) {
		printErr("Error: Release '%s' not found in namespace '%s'.", release, ns)
		os.Exit(1)
	}

	if !deployYes {
		if !confirm(fmt.Sprintf("Uninstall release '%s' from namespace '%s'?", release, ns)) {
			fmt.Println("Uninstall cancelled.")
			return
		}
	}

	helmArgs := []string{"uninstall", release, "--namespace", ns}
	if deployDryRun {
		helmArgs = append(helmArgs, "--dry-run")
	}

	fmt.Printf("Running: helm %s\n\n", strings.Join(helmArgs, " "))

	if err := helmRun(helmArgs...); err != nil {
		printErr("Helm uninstall failed: %v", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println(tui.SuccessStyle.Render(fmt.Sprintf("Release '%s' uninstalled successfully from namespace '%s'.", release, ns)))
}

func runStatus(args []string) {
	requireHelm()

	chartDir := resolveDeployChart(args)
	release := deployRelease
	if release == "" {
		release = releaseNameFromChart(chartDir)
	}

	// Read namespace from chart values if user didn't specify -n
	chartNS := readChartNamespace(chartDir)
	ns := deployNamespace
	if ns == "default" && chartNS != "default" {
		ns = chartNS
	}

	fmt.Println(tui.HeaderStyle.Render("Helm Release Status"))
	fmt.Printf("  Release:   %s\n", release)
	fmt.Printf("  Namespace: %s\n", ns)
	fmt.Println()

	if !helmReleaseExists(release, ns) {
		printErr("Error: Release '%s' not found in namespace '%s'.", release, ns)
		fmt.Printf("Deploy it first with: kgen deploy %s\n", chartDir)
		os.Exit(1)
	}

	helmArgs := []string{"status", release, "--namespace", ns}
	fmt.Printf("Running: helm %s\n\n", strings.Join(helmArgs, " "))

	if err := helmRun(helmArgs...); err != nil {
		printErr("Helm status failed: %v", err)
		os.Exit(1)
	}
}
