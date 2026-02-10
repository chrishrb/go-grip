package ghissue

import (
	"os/exec"
	"regexp"
	"strings"
)

var gitHubRepoRegex = regexp.MustCompile(`github\.com[:/]([^/]+)/([^/\s]+?)(?:\.git)?$`)

// DetectRepository attempts to detect the GitHub repository from git remote
func DetectRepository() string {
	// Try to get the git remote URL
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	remoteURL := strings.TrimSpace(string(output))

	// Extract owner/repo from GitHub URL
	matches := gitHubRepoRegex.FindStringSubmatch(remoteURL)
	if len(matches) >= 3 {
		owner := matches[1]
		repo := strings.TrimSuffix(matches[2], ".git")
		return owner + "/" + repo
	}

	return ""
}
