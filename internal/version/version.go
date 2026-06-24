package version

import "fmt"

// RepoOwner and RepoName identify the GitHub source repository.
// These are shared by the update / uninstall commands and the install script.
const (
	RepoOwner = "ihyamarsdev"
	RepoName  = "kgen"
)

// Version holds the current KGen release tag.
//
// It is intentionally maintained by hand per release. Distributors may override
// it at build time via -ldflags:
//
//	go build -ldflags "-X github.com/ihyamarsdev/kgen/internal/version.Version=v0.4.4" -o kgen main.go
var Version = "v0.4.4"

// RepoURL returns the canonical GitHub URL for the project.
func RepoURL() string {
	return fmt.Sprintf("https://github.com/%s/%s", RepoOwner, RepoName)
}

// LatestReleaseURL returns the GitHub API endpoint for the latest release.
func LatestReleaseURL() string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", RepoOwner, RepoName)
}

// UserAgent returns a value suitable for the HTTP User-Agent header.
func UserAgent() string {
	return fmt.Sprintf("kgen/%s", Version)
}
