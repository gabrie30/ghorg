package git

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/gabrie30/ghorg/scm"
)

// gitError reports a failed git command using the first meaningful line of git's stderr, which is usually the
// useful part (for example "Authentication failed for ..."). The underlying error is kept so callers can check
// the exit code with errors.As.
type gitError struct {
	msg string
	err error
}

func (e *gitError) Error() string { return e.msg }

func (e *gitError) Unwrap() error { return e.err }

// runGitOutput runs git in repo.HostPath and returns its trimmed stdout. Failures return a gitError. It does not use
// printDebugCmd because that runs the command itself, and the command could not then be run again to capture output.
func runGitOutput(repo scm.Repo, args ...string) (string, error) {
	return runGitOutputEnv(repo, nil, args...)
}

// runGitOutputEnv is runGitOutput with extra environment entries, in KEY=value form, added to the process environment.
func runGitOutputEnv(repo scm.Repo, env []string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repo.HostPath
	if env != nil {
		cmd.Env = append(os.Environ(), env...)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if os.Getenv("GHORG_DEBUG") != "" {
		fmt.Println("------------- GIT DEBUG -------------")
		fmt.Printf("Repo Path: %s\n", repo.HostPath)
		fmt.Printf("Command Ran: git %s\n", strings.Join(args, " "))
		fmt.Printf("Command Output: %s%s\n", stdout.String(), stderr.String())
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	if err != nil {
		return "", &gitError{msg: firstErrorLine(stderr.String(), err), err: err}
	}
	return strings.TrimSpace(stdout.String()), nil
}

// firstErrorLine returns the first line of stderr that is not a warning, hint, or fetch's "From " header, without
// git's "fatal: " or "error: " prefix. It falls back to err when git printed nothing useful.
func firstErrorLine(stderr string, err error) string {
	for _, line := range strings.Split(stderr, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "warning: ") || strings.HasPrefix(line, "hint: ") ||
			strings.HasPrefix(line, "From ") {
			continue
		}
		line = strings.TrimPrefix(line, "fatal: ")
		return strings.TrimPrefix(line, "error: ")
	}
	return err.Error()
}

// WorkingTreeStatus returns `git status --porcelain` output, which is empty when there are no uncommitted changes
// or untracked files. Unlike ShortStatus it runs git once, so it also works with GHORG_DEBUG set.
func (g GitClient) WorkingTreeStatus(repo scm.Repo) (string, error) {
	return runGitOutput(repo, "status", "--porcelain")
}

// AbsoluteGitDir returns the absolute path of the repo's git directory. For a linked worktree this is the
// worktree's own directory, which is where git keeps in-progress operation state such as MERGE_HEAD.
func (g GitClient) AbsoluteGitDir(repo scm.Repo) (string, error) {
	return runGitOutput(repo, "rev-parse", "--absolute-git-dir")
}

// HasRemote reports whether the repo has a remote called name.
func (g GitClient) HasRemote(repo scm.Repo, name string) bool {
	_, err := runGitOutput(repo, "remote", "get-url", name)
	return err == nil
}

// sshBatchModeEnv returns GIT_SSH_COMMAND set to batch mode, so a locked key or unknown host key fails instead of
// prompting. It returns nil when the user already chose an SSH command through GIT_SSH_COMMAND, GIT_SSH, or
// core.sshCommand in any git config that applies to the repo, including the repo's own.
func sshBatchModeEnv(repo scm.Repo) []string {
	if os.Getenv("GIT_SSH_COMMAND") != "" || os.Getenv("GIT_SSH") != "" {
		return nil
	}
	if sshCommand, err := runGitOutput(repo, "config", "--get", "core.sshCommand"); err == nil && sshCommand != "" {
		return nil
	}
	return []string{"GIT_SSH_COMMAND=ssh -o BatchMode=yes"}
}

// FetchOrigin fetches from origin. It only updates remote-tracking refs and never touches local branches or files.
// Unless the user configured an SSH command, SSH runs in batch mode so it fails instead of prompting.
func (g GitClient) FetchOrigin(repo scm.Repo) error {
	_, err := runGitOutputEnv(repo, sshBatchModeEnv(repo), "fetch", "origin")
	return err
}

// CurrentBranch returns the checked out branch. detached is true when HEAD points at a commit instead of a branch.
// Unlike GetCurrentBranch it works on a branch with no commits yet.
func (g GitClient) CurrentBranch(repo scm.Repo) (string, bool, error) {
	out, err := runGitOutput(repo, "symbolic-ref", "--quiet", "HEAD")
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return "", true, nil
	}
	if err != nil {
		return "", false, err
	}
	return strings.TrimPrefix(out, "refs/heads/"), false, nil
}

// LocalBranchExists reports whether refs/heads/<branch> exists.
func (g GitClient) LocalBranchExists(repo scm.Repo, branch string) bool {
	_, err := runGitOutput(repo, "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	return err == nil
}

// Upstream returns the short name of a local branch's upstream (for example origin/main) and whether its
// remote-tracking ref is gone. An empty name means no upstream is configured.
func (g GitClient) Upstream(repo scm.Repo, branch string) (string, bool, error) {
	out, err := runGitOutput(repo, "for-each-ref", "--format=%(upstream:short) %(upstream:track)", "refs/heads/"+branch)
	if err != nil {
		return "", false, err
	}
	name, track, _ := strings.Cut(out, " ")
	return name, track == "[gone]", nil
}

// AheadBehind counts commits on local that are not on upstream (ahead) and commits on upstream that are not on
// local (behind).
func (g GitClient) AheadBehind(repo scm.Repo, local, upstream string) (int, int, error) {
	out, err := runGitOutput(repo, "rev-list", "--left-right", "--count", local+"..."+upstream, "--")
	if err != nil {
		return 0, 0, err
	}
	fields := strings.Fields(out)
	if len(fields) != 2 {
		return 0, 0, fmt.Errorf("unexpected rev-list output %q", out)
	}
	ahead, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0, 0, err
	}
	behind, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0, 0, err
	}
	return ahead, behind, nil
}

// MergeFastForwardOnly moves the checked out branch to target. Git refuses if that is not a fast-forward or if it
// would overwrite local changes.
func (g GitClient) MergeFastForwardOnly(repo scm.Repo, target string) error {
	_, err := runGitOutput(repo, "merge", "--ff-only", target)
	return err
}

// DefaultBranch returns the remote's default branch name (for example main) from refs/remotes/origin/HEAD.
func (g GitClient) DefaultBranch(repo scm.Repo) (string, error) {
	out, err := runGitOutput(repo, "symbolic-ref", "refs/remotes/origin/HEAD")
	if err != nil {
		return "", err
	}
	branch := strings.TrimPrefix(out, "refs/remotes/origin/")
	if _, err := runGitOutput(repo, "show-ref", "--verify", "--quiet", "refs/remotes/origin/"+branch); err != nil {
		return "", fmt.Errorf("origin/HEAD points to missing origin/%s", branch)
	}
	return branch, nil
}

// BranchCheckedOutInWorktree reports whether branch is checked out in any worktree of the repo, including the main
// one. Pull checks this itself because older git versions do not refuse to move a branch checked out elsewhere.
func (g GitClient) BranchCheckedOutInWorktree(repo scm.Repo, branch string) (bool, error) {
	out, err := runGitOutput(repo, "worktree", "list", "--porcelain")
	if err != nil {
		return false, err
	}
	want := "branch refs/heads/" + branch
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == want {
			return true, nil
		}
	}
	return false, nil
}

// FastForwardBranch moves a local branch that is not checked out to origin/<branch> without touching any files.
// Git refuses the update if it is not a fast-forward.
func (g GitClient) FastForwardBranch(repo scm.Repo, branch string) error {
	_, err := runGitOutput(repo, "fetch", ".", "refs/remotes/origin/"+branch+":refs/heads/"+branch)
	return err
}

// ResetHardHead discards uncommitted changes to tracked files on the checked out branch.
func (g GitClient) ResetHardHead(repo scm.Repo) error {
	_, err := runGitOutput(repo, "reset", "--hard", "HEAD")
	return err
}

// DiffHead returns `git diff HEAD`: staged and unstaged changes to tracked files. External diff tools are disabled,
// and the diff is colored only when GHORG_COLOR is enabled and ghorg's output supports color (a terminal, without
// NO_COLOR).
func (g GitClient) DiffHead(repo scm.Repo) (string, error) {
	colorFlag := "--no-color"
	if os.Getenv("GHORG_COLOR") == "enabled" && !color.NoColor {
		colorFlag = "--color=always"
	}
	return runGitOutput(repo, "diff", "--no-ext-diff", colorFlag, "HEAD", "--")
}
