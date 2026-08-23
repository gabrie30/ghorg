package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/gabrie30/ghorg/scm"
)

// FilterByHook pipes the repo list as a JSON array to the executable set in
// GHORG_REPO_FILTER_HOOK and returns the list the hook writes to its stdout.
// The hook may remove repos and may also modify repo fields. Any hook failure
// aborts the run so ghorg never clones a list the hook did not approve.
func (rf *RepositoryFilter) FilterByHook(repos []scm.Repo) []scm.Repo {
	hookPath := os.Getenv("GHORG_REPO_FILTER_HOOK")
	if hookPath == "" {
		return repos
	}

	colorlog.PrintInfo("Filtering repos with repo filter hook...")

	filtered, err := runRepoFilterHook(hookPath, repos)
	if err != nil {
		colorlog.PrintErrorAndExit(fmt.Sprintf("Error running your GHORG_REPO_FILTER_HOOK, error: %v", err))
	}

	colorlog.PrintInfo(fmt.Sprintf("Repo filter hook returned %v of %v repos", len(filtered), len(repos)))

	return filtered
}

// runRepoFilterHook executes the hook directly (no shell). The hook's stderr
// is passed through so script diagnostics are visible while it runs.
func runRepoFilterHook(hookPath string, repos []scm.Repo) ([]scm.Repo, error) {
	if repos == nil {
		repos = []scm.Repo{}
	}

	input, err := json.Marshal(repos)
	if err != nil {
		return nil, fmt.Errorf("could not marshal repos to JSON: %w", err)
	}

	cmd := exec.Command(hookPath)
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("hook %v failed: %w", hookPath, err)
	}

	filtered := []scm.Repo{}
	if err := json.Unmarshal(output, &filtered); err != nil {
		return nil, fmt.Errorf("hook %v output is not a valid JSON array of repos: %w", hookPath, err)
	}

	return filtered, nil
}
