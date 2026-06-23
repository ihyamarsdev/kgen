package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ihyamarsdev/kgen/internal/tui"
	"github.com/ihyamarsdev/kgen/internal/version"

	"github.com/spf13/cobra"
)

// httpClient is a shared client with a 15-second timeout for all HTTP calls.
var httpClient = &http.Client{Timeout: 15 * time.Second}

// updateYesFlag (-y / --yes) skips the interactive confirmation prompt.
var updateYesFlag bool

type githubRelease struct {
	TagName string `json:"tag_name"`
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update kgen to the latest released version",
	Long: `Check GitHub for the latest KGen release and replace the currently running
binary in-place. Mirrors install.sh: downloads the precompiled asset that
matches your OS/architecture and installs it into the same directory as the
running kgen, using sudo when that directory is not writable.`,
	Run: func(cmd *cobra.Command, args []string) {
		runUpdate()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolVarP(&updateYesFlag, "yes", "y", false, "skip the confirmation prompt")
}

func runUpdate() {
	fmt.Printf("Current version: %s\n", tui.GrayStyle.Render(version.Version))
	fmt.Println("Checking for the latest release...")

	latest, err := fetchLatestRelease()
	if err != nil {
		printErr("Error fetching latest release: %v", err)
		os.Exit(1)
	}

	if latest == version.Version {
		fmt.Println(tui.SuccessStyle.Render("You are already on the latest version of kgen."))
		return
	}

	fmt.Printf("Latest version:  %s\n", tui.ActiveInputStyle.Render(latest))

	// Locate the binary currently in use so we replace it in-place.
	currentBin, err := os.Executable()
	if err != nil {
		printErr("Error locating the running binary: %v", err)
		os.Exit(1)
	}
	currentBin, _ = filepath.EvalSymlinks(currentBin)
	binDir := filepath.Dir(currentBin)

	asset := assetName()
	downloadURL := fmt.Sprintf(
		"https://github.com/%s/%s/releases/download/%s/%s",
		version.RepoOwner, version.RepoName, latest, asset,
	)
	fmt.Printf("Downloading:     %s\n", downloadURL)

	tmpDir, err := os.MkdirTemp("", "kgen-update-*")
	if err != nil {
		printErr("Error creating temp directory: %v", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, "kgen")
	if err := downloadFile(downloadURL, tmpFile); err != nil {
		printErr("Error downloading update: %v", err)
		os.Exit(1)
	}
	if err := os.Chmod(tmpFile, 0o755); err != nil {
		printErr("Error marking binary as executable: %v", err)
		os.Exit(1)
	}

	targetPath := filepath.Join(binDir, "kgen")
	if !updateYesFlag {
		if !confirm(fmt.Sprintf("Install kgen %s to %s?", latest, targetPath)) {
			fmt.Println("Update cancelled.")
			return
		}
	}

	installPath, err := installBinary(tmpFile, targetPath)
	if err != nil {
		printErr("Error installing binary: %v", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println(tui.SuccessStyle.Render("KGen updated successfully."))
	fmt.Printf("Installed:       %s\n", installPath)
	fmt.Println("Verify with:")
	fmt.Println(tui.GrayStyle.Render("  kgen --version"))
}

// fetchLatestRelease queries the GitHub API for the latest release tag.
func fetchLatestRelease() (string, error) {
	req, err := http.NewRequest(http.MethodGet, version.LatestReleaseURL(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", version.UserAgent())
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github api returned status %s", resp.Status)
	}

	var rel githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", err
	}
	if rel.TagName == "" {
		return "", fmt.Errorf("could not determine the latest release tag")
	}
	return rel.TagName, nil
}

// assetName returns the release asset name for the current GOOS/GOARCH.
// It exits the process when the running OS/architecture is unsupported,
// matching the behavior of install.sh.
func assetName() string {
	osName := runtime.GOOS
	if osName != "linux" && osName != "darwin" {
		printErr("Error: unsupported OS: %s", runtime.GOOS)
		os.Exit(1)
	}
	arch := runtime.GOARCH
	if arch != "amd64" && arch != "arm64" {
		printErr("Error: unsupported architecture: %s", runtime.GOARCH)
		os.Exit(1)
	}
	return fmt.Sprintf("kgen-%s-%s", osName, arch)
}

// downloadFile streams url to dest, failing on non-2xx responses.
func downloadFile(url, dest string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", version.UserAgent())

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// installBinary moves src to dest, escalating via sudo when dest is not writable.
//
// To avoid a race condition where the old binary is removed but the new one
// isn't in place (if killed mid-operation), we download to a temp file in the
// *same directory* as dest, then use os.Rename() which is atomic on the same
// filesystem.
func installBinary(src, dest string) (string, error) {
	dir := filepath.Dir(dest)
	if _, err := os.Stat(dir); err != nil {
		return "", fmt.Errorf("install dir %s does not exist", dir)
	}

	// If we can write directly, download a temp file in the same directory
	// and atomically rename it into place.
	if isWritableDir(dir) {
		// Copy src to a temp file in the same directory so os.Rename is atomic.
		tmpDest := dest + ".tmp"
		if err := copyFile(src, tmpDest); err != nil {
			return "", err
		}
		if err := os.Chmod(tmpDest, 0o755); err != nil {
			os.Remove(tmpDest)
			return "", fmt.Errorf("failed to set permissions: %w", err)
		}
		if err := os.Rename(tmpDest, dest); err != nil {
			os.Remove(tmpDest)
			return "", fmt.Errorf("failed to replace binary: %w", err)
		}
		return dest, nil
	}

	// Need elevated privileges. Fall back to sudo mv/cp.
	if _, err := exec.LookPath("sudo"); err != nil {
		return "", fmt.Errorf("sudo is required to write to %s but was not found", dir)
	}
	fmt.Printf("sudo is required to write to %s\n", dir)
	if err := sudoReplaceBinary(src, dest); err != nil {
		return "", err
	}
	return dest, nil
}

func sudoReplaceBinary(src, dest string) error {
	// Copy to a sudo-writable staging path next to src, then sudo mv into place.
	stage := src + ".stage"
	if err := copyFile(src, stage); err != nil {
		return err
	}
	defer os.Remove(stage)
	cmd := exec.Command("sudo", "mv", stage, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// copyFile copies src to dest, preserving dest's existing mode (or 0o755).
func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	mode := os.FileMode(0o755)
	if info, err := os.Stat(dest); err == nil {
		mode = info.Mode()
	}
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// isWritableDir reports whether the current user can write to dir.
func isWritableDir(dir string) bool {
	f, err := os.CreateTemp(dir, ".kgen-write-test-*")
	if err != nil {
		return false
	}
	_ = f.Close()
	_ = os.Remove(f.Name())
	return true
}
