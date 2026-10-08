package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/gabrie30/ghorg/git"
	"github.com/korovkin/limiter"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull [dir]",
	Short: "Update every git repo in a directory without discarding local changes",
	Long: `Update every git repo found under dir (default: the current directory).

No token or SCM API calls are needed. Git uses your existing credentials (SSH agent
or credential helper). HTTPS credential prompts are disabled, and SSH runs in batch
mode unless you configured your own SSH command (GIT_SSH_COMMAND, GIT_SSH, or
core.sshCommand), so a missing credential shows up as a failure instead of hanging.

SAFE MODE (default):
  Only fast-forward updates are made and nothing local is discarded.
  - The checked out branch is fast-forwarded from its upstream.
  - The local default branch is fast-forwarded without being checked out.
  - Repos with uncommitted changes, untracked files, a detached HEAD, no upstream,
    a diverged branch, or an unfinished merge or rebase are skipped.
  Every skipped or failed repo is listed at the end with the reason.
  With --verbose, each repo skipped for local changes also lists its untracked
  files and the first 100 lines of git diff HEAD.

RESET HARD MODE (--reset-hard), DESTRUCTIVE:
  Discards local changes, deletes untracked files, and resets each repo to origin's
  default branch: reset, clean, check out the default branch, then reset it to
  origin. Ignored files are kept. Other branches are not modified, and a repo whose
  default branch is checked out in another worktree is skipped. --reset-hard has no
  short form and cannot be set in conf.yaml.

EXIT CODE:
  0 when no repo failed (skipped repos are expected), 1 when any repo failed.

EXAMPLES:
  # Update every repo in the current directory
  $ ghorg pull .

  # Update every repo in a ghorg clone directory
  $ ghorg pull ~/ghorg/my-org

  # Make every repo match origin's default branch, discarding local changes
  $ ghorg pull ~/ghorg/my-org --reset-hard

  # See the untracked files and diff of each repo skipped for local changes
  $ ghorg pull . --verbose
`,
	Args: cobra.MaximumNArgs(1),
	Run:  pullFunc,
}

func pullFunc(cmd *cobra.Command, argz []string) {
	if cmd.Flags().Changed("concurrency") {
		_ = os.Setenv("GHORG_CONCURRENCY", cmd.Flag("concurrency").Value.String())
	}
	force, _ := cmd.Flags().GetBool("reset-hard")
	verbose, _ := cmd.Flags().GetBool("verbose")

	dir := "."
	if len(argz) == 1 {
		dir = argz[0]
	}
	root, err := resolvePullRoot(dir)
	if err != nil {
		colorlog.PrintErrorAndExit(err.Error())
	}

	concurrency, err := strconv.Atoi(os.Getenv("GHORG_CONCURRENCY"))
	if err != nil || concurrency < 1 {
		colorlog.PrintErrorAndExit(fmt.Sprintf("GHORG_CONCURRENCY must be a positive integer, got %q", os.Getenv("GHORG_CONCURRENCY")))
	}

	repos, bare, err := discoverRepos(root)
	if err != nil {
		colorlog.PrintErrorAndExit(fmt.Sprintf("Could not search %s for repos: %v", root, err))
	}
	if len(repos) == 0 && len(bare) == 0 {
		colorlog.PrintInfo(fmt.Sprintf("No git repos found in %s", root))
		return
	}

	disablePullPrompts()
	if len(repos) > 0 {
		colorlog.PrintInfo(fmt.Sprintf("Pulling %d repos in %s\n", len(repos), root))
	}

	results := runPulls(git.NewGit(), root, repos, pullOptions{force: force, verbose: verbose}, concurrency)
	for _, path := range bare {
		r := pullResult{Path: pullDisplayPath(root, path)}
		r.markSkipped("bare repository")
		printPullProgress(r)
		results = append(results, r)
	}

	fmt.Println(pullSummary(results))
	os.Exit(pullExitCode(results))
}

// resolvePullRoot makes dir absolute and resolves symlinks when possible so the walk starts from a real directory.
func resolvePullRoot(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("could not resolve %s: %v", dir, err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("could not resolve %s: %v", dir, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", dir)
	}
	// Some Windows drives cannot be resolved, so fall back to the absolute path.
	if root, err := filepath.EvalSymlinks(abs); err == nil {
		return root, nil
	}
	return abs, nil
}

// disablePullPrompts stops git from asking for credentials on the terminal, and Git Credential Manager from opening
// GUI prompts, so a missing credential fails with a reason instead of hanging the run. SSH batch mode is set per
// repo by FetchOrigin so a repo's own core.sshCommand is respected.
func disablePullPrompts() {
	_ = os.Setenv("GIT_TERMINAL_PROMPT", "0")
	_ = os.Setenv("GCM_INTERACTIVE", "never")
}

// pullOptions are the command-line choices that change how each repo is handled and reported.
type pullOptions struct {
	force   bool
	verbose bool
}

// runPulls pulls repos concurrently and prints a progress line as each one finishes. Worktrees of one repo share refs,
// so they run one after another instead of concurrently. Results keep the input order.
// With verbose, repos skipped for local changes also get their untracked files and diff for the summary.
func runPulls(g pullGitter, root string, repos []string, opts pullOptions, concurrency int) []pullResult {
	results := make([]pullResult, len(repos))
	limit := limiter.NewConcurrencyLimiter(concurrency)
	for _, group := range groupBySharedGitDir(repos) {
		_, _ = limit.Execute(func() {
			for _, i := range group {
				results[i] = pullRepo(g, repos[i], pullDisplayPath(root, repos[i]), opts.force)
				if opts.verbose && results[i].Dirty {
					results[i].LocalChanges = localChangeLines(g, repos[i])
				}
				printPullProgress(results[i])
			}
		})
	}
	_ = limit.WaitAndClose()
	return results
}

// groupBySharedGitDir returns repo indexes grouped by git common dir, in first-seen order. A repo whose common dir
// cannot be read gets its own group.
func groupBySharedGitDir(repos []string) [][]int {
	var groups [][]int
	byDir := map[string]int{}
	for i, repoPath := range repos {
		dir := gitCommonDir(repoPath)
		if dir == "" {
			groups = append(groups, []int{i})
			continue
		}
		if g, ok := byDir[dir]; ok {
			groups[g] = append(groups[g], i)
			continue
		}
		byDir[dir] = len(groups)
		groups = append(groups, []int{i})
	}
	return groups
}

// gitCommonDir returns the git directory shared by all worktrees of the repo at dir, read from the filesystem: the
// .git directory itself, or for a linked worktree the directory named by the .git file's gitdir line and that
// directory's commondir file. Symlinks are resolved so the same repo always yields the same path. It returns "" when
// the layout cannot be read, so the repo is pulled on its own.
func gitCommonDir(dir string) string {
	dotGit := filepath.Join(dir, ".git")
	info, err := os.Stat(dotGit)
	if err != nil {
		return ""
	}
	common := dotGit
	if !info.IsDir() {
		content, err := os.ReadFile(dotGit)
		if err != nil {
			return ""
		}
		gitDir, ok := strings.CutPrefix(strings.TrimSpace(string(content)), "gitdir:")
		if !ok {
			return ""
		}
		gitDir = strings.TrimSpace(gitDir)
		if !filepath.IsAbs(gitDir) {
			gitDir = filepath.Join(dir, gitDir)
		}
		common = gitDir
		if rel, err := os.ReadFile(filepath.Join(gitDir, "commondir")); err == nil {
			common = strings.TrimSpace(string(rel))
			if !filepath.IsAbs(common) {
				common = filepath.Join(gitDir, common)
			}
		}
	}
	common = filepath.Clean(common)
	if resolved, err := filepath.EvalSymlinks(common); err == nil {
		return resolved
	}
	return common
}

// discoverRepos walks root and returns the absolute paths of git work trees, plus bare repos, which pull skips.
// A directory containing .git (a directory, or a file for worktrees) is a repo and is not descended into, so repos
// nested inside it are not reported. Symlinks are not followed.
func discoverRepos(root string) ([]string, []string, error) {
	var repos, bare []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if path == root {
				return walkErr
			}
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if isWorkTree(path) {
			repos = append(repos, path)
			return fs.SkipDir
		}
		if isBareRepo(path) {
			bare = append(bare, path)
			return fs.SkipDir
		}
		return nil
	})
	return repos, bare, err
}

// isWorkTree reports whether dir is the root of a git work tree.
func isWorkTree(dir string) bool {
	_, err := os.Lstat(filepath.Join(dir, ".git"))
	return err == nil
}

// isBareRepo reports whether dir looks like a bare git repo: a HEAD file next to objects and refs directories.
func isBareRepo(dir string) bool {
	head, err := os.Stat(filepath.Join(dir, "HEAD"))
	if err != nil || head.IsDir() {
		return false
	}
	for _, sub := range []string{"objects", "refs"} {
		info, err := os.Stat(filepath.Join(dir, sub))
		if err != nil || !info.IsDir() {
			return false
		}
	}
	return true
}

// pullDisplayPath returns repoPath relative to root for output, or the repo's own name when root is the repo.
func pullDisplayPath(root, repoPath string) string {
	rel, err := filepath.Rel(root, repoPath)
	if err != nil || rel == "." {
		return filepath.Base(repoPath)
	}
	return rel
}

// pullProgressLine is the line printed as each repo finishes. Notes are joined with "; " because some contain commas.
func pullProgressLine(r pullResult) string {
	notes := strings.Join(r.Notes, "; ")
	switch r.Status {
	case pullUpdated:
		return fmt.Sprintf("Success pulling %s: %s", r.Path, notes)
	case pullSkipped:
		return fmt.Sprintf("Skipped %s: %s", r.Path, notes)
	case pullFailed:
		return fmt.Sprintf("Failed %s: %s", r.Path, notes)
	default:
		if notes == "" {
			return "Up to date " + r.Path
		}
		return fmt.Sprintf("Up to date %s: %s", r.Path, notes)
	}
}

// printPullProgress prints r's progress line in the color for its status.
func printPullProgress(r pullResult) {
	line := pullProgressLine(r)
	switch r.Status {
	case pullUpdated:
		colorlog.PrintSuccess(line)
	case pullSkipped:
		colorlog.PrintInfo(line)
	case pullFailed:
		colorlog.PrintError(line)
	default:
		colorlog.PrintSubtleInfo(line)
	}
}

// pullSummary lists skipped and failed repos sorted by path, with any --verbose local changes under each repo,
// then the count for each status.
func pullSummary(results []pullResult) string {
	sorted := slices.Clone(results)
	slices.SortFunc(sorted, func(a, b pullResult) int { return strings.Compare(a.Path, b.Path) })

	counts := map[pullStatus]int{}
	for _, r := range sorted {
		counts[r.Status]++
	}

	var b strings.Builder
	sections := []struct {
		title  string
		status pullStatus
	}{
		{"Skipped", pullSkipped},
		{"Failed", pullFailed},
	}
	for _, section := range sections {
		if counts[section.status] == 0 {
			continue
		}
		width := 0
		for _, r := range sorted {
			if r.Status == section.status {
				width = max(width, len(r.Path))
			}
		}
		fmt.Fprintf(&b, "\n%s (%d)\n", section.title, counts[section.status])
		for _, r := range sorted {
			if r.Status == section.status {
				fmt.Fprintf(&b, "  %-*s  %s\n", width, r.Path, strings.Join(r.Notes, "; "))
				for _, line := range r.LocalChanges {
					fmt.Fprintf(&b, "    %s\n", line)
				}
			}
		}
	}
	fmt.Fprintf(&b, "\nUpdated %d, up to date %d, skipped %d, failed %d",
		counts[pullUpdated], counts[pullUpToDate], counts[pullSkipped], counts[pullFailed])
	return b.String()
}

// pullExitCode is 1 when any repo failed. Skipped repos are expected and do not change it.
func pullExitCode(results []pullResult) int {
	for _, r := range results {
		if r.Status == pullFailed {
			return 1
		}
	}
	return 0
}
