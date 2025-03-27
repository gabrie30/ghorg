package git

import (
	"os"
	"strconv"

	"github.com/gabrie30/ghorg/scm"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

type Gitter interface {
	Clone(scm.Repo) error
	Reset(scm.Repo) error
	Pull(scm.Repo) error
	SetOrigin(scm.Repo) error
	SetOriginWithCredentials(scm.Repo) error
	Clean(scm.Repo) error
	Checkout(scm.Repo) error
	UpdateRemote(scm.Repo) error
	FetchAll(scm.Repo) error
}

type GitClient struct{}

func NewGit() GitClient {
	return GitClient{}
}

func getCloneDepth() int {
	cloneDepthStr := os.Getenv("GHORG_CLONE_DEPTH")
	if cloneDepthStr != "" {
		if depth, err := strconv.Atoi(cloneDepthStr); err == nil && depth > 0 {
			return depth
		}
	}
	return 1 // Default depth
}

func (g GitClient) Clone(repo scm.Repo) error {
	recurseSubmodules := os.Getenv("GHORG_INCLUDE_SUBMODULES") == "true"

	// 	if os.Getenv("GHORG_GIT_FILTER") != "" {
	// 		index := 1
	// 		args = append(args[:index+1], args[index:]...)
	// 		args[index] = fmt.Sprintf("--filter=%v", os.Getenv("GHORG_GIT_FILTER"))
	// 	}

	// currently go-git does not support --filter, it should be coming in v6
	// see https://github.com/go-git/go-git/issues/1381
	_, err := git.PlainClone(repo.HostPath, false, &git.CloneOptions{
		URL:   repo.CloneURL,
		Depth: getCloneDepth(),
		RecurseSubmodules: git.SubmoduleRescursivity(func() int {
			if recurseSubmodules {
				return int(git.DefaultSubmoduleRecursionDepth)
			}
			return 0
		}()),
		Mirror: os.Getenv("GHORG_BACKUP") == "true",
	})

	return err
}

func (g GitClient) SetOriginWithCredentials(repo scm.Repo) error {
	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return err
	}

	cfg, _ := r.Config()
	cfg.Remotes[git.DefaultRemoteName] = &config.RemoteConfig{
		Name: git.DefaultRemoteName,
		URLs: []string{repo.CloneURL},
	}

	r.Storer.SetConfig(cfg)
	return nil
}

func (g GitClient) SetOrigin(repo scm.Repo) error {
	return g.SetOriginWithCredentials(repo)
}

func (g GitClient) Checkout(repo scm.Repo) error {
	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(repo.CloneBranch),
	})

	return err
}

func (g GitClient) Clean(repo scm.Repo) error {
	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	err = w.Clean(&git.CleanOptions{
		Dir: true,
	})

	return err
}

func (g GitClient) UpdateRemote(repo scm.Repo) error {
	return g.SetOriginWithCredentials(repo)
}

func (g GitClient) Pull(repo scm.Repo) error {
	recurseSubmodules := os.Getenv("GHORG_INCLUDE_SUBMODULES") == "true"

	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	err = w.Pull(&git.PullOptions{
		RemoteName: git.DefaultRemoteName,
		Force:      true,
		Depth:      getCloneDepth(),
		RecurseSubmodules: git.SubmoduleRescursivity(func() int {
			if recurseSubmodules {
				return int(git.DefaultSubmoduleRecursionDepth)
			}
			return 0
		}()),
	})

	return err
}

func (g GitClient) Reset(repo scm.Repo) error {
	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	err = w.Reset(&git.ResetOptions{
		Mode: git.HardReset,
	})

	return err
}

func (g GitClient) FetchAll(repo scm.Repo) error {
	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return err
	}

	err = r.Fetch(&git.FetchOptions{
		RemoteName: git.DefaultRemoteName,
		RemoteURL:  repo.URL,
		Depth:      getCloneDepth(),
	})
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}

	return err
}
