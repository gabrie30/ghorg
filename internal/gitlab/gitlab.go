package gitlab

import (
	"os"
	"strings"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/gabrie30/ghorg/internal/repo"

	gitlab "github.com/xanzy/go-gitlab"
)

// GetOrgRepos fetches repo data from a specific group
func GetOrgRepos(targetOrg string) ([]repo.Data, error) {
	repoData := []repo.Data{}
	client, err := determineClient()

	if err != nil {
		colorlog.PrintError(err)
	}

	opt := &gitlab.ListGroupProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
			Page:    1,
		},
		IncludeSubgroups: gitlab.Bool(true),
	}

	for {
		// Get the first page with projects.
		ps, resp, err := client.Groups.ListGroupProjects(targetOrg, opt)

		if err != nil {
			// TODO: check if 404, then we know group does not exist
			return []repo.Data{}, err
		}

		// List all the projects we've found so far.
		for _, p := range ps {

			if os.Getenv("GHORG_SKIP_ARCHIVED") == "true" {
				if p.Archived == true {
					continue
				}
			}

			if os.Getenv("GHORG_MATCH_PREFIX") != "" {
				repoName := strings.ToLower(p.Name)
				foundPrefix := false
				pfs := strings.Split(os.Getenv("GHORG_MATCH_PREFIX"), ",")
				for _, p := range pfs {
					if strings.HasPrefix(repoName, strings.ToLower(p)) {
						foundPrefix = true
					}
				}
				if foundPrefix == false {
					continue
				}
			}

			r := repo.Data{}

			r.Path = p.PathWithNamespace
			if os.Getenv("GHORG_CLONE_PROTOCOL") == "https" {
				r.CloneURL = addTokenToHTTPSCloneURL(p.HTTPURLToRepo, os.Getenv("GHORG_GITLAB_TOKEN"))
				r.URL = p.HTTPURLToRepo
				repoData = append(repoData, r)
			} else {
				r.CloneURL = p.SSHURLToRepo
				r.URL = p.SSHURLToRepo
				repoData = append(repoData, r)
			}
		}

		// Exit the loop when we've seen all pages.
		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		// Update the page number to get the next page.
		opt.Page = resp.NextPage
	}

	return repoData, nil
}

func determineClient() (*gitlab.Client, error) {
	baseURL := os.Getenv("GHORG_SCM_BASE_URL")
	token := os.Getenv("GHORG_GITLAB_TOKEN")

	if baseURL != "" {
		client, err := gitlab.NewClient(token, gitlab.WithBaseURL(baseURL))
		return client, err
	}

	return gitlab.NewClient(token)
}

// GetUserRepos gets all of a users gitlab repos
func GetUserRepos(targetUsername string) ([]repo.Data, error) {
	cloneData := []repo.Data{}

	client, err := determineClient()

	if err != nil {
		colorlog.PrintError(err)
	}

	opt := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 50,
			Page:    1,
		},
	}

	for {
		// Get the first page with projects.
		ps, resp, err := client.Projects.ListUserProjects(targetUsername, opt)
		if err != nil {
			// TODO: check if 404, then we know user does not exist
			return []repo.Data{}, err
		}

		// List all the projects we've found so far.
		for _, p := range ps {
			if os.Getenv("GHORG_SKIP_ARCHIVED") == "true" {
				if p.Archived == true {
					continue
				}
			}

			if os.Getenv("GHORG_MATCH_PREFIX") != "" {
				repoName := strings.ToLower(p.Name)
				foundPrefix := false
				pfs := strings.Split(os.Getenv("GHORG_MATCH_PREFIX"), ",")
				for _, p := range pfs {
					if strings.HasPrefix(repoName, strings.ToLower(p)) {
						foundPrefix = true
					}
				}
				if foundPrefix == false {
					continue
				}
			}

			r := repo.Data{}
			r.Path = p.PathWithNamespace
			if os.Getenv("GHORG_CLONE_PROTOCOL") == "https" {
				r.CloneURL = addTokenToHTTPSCloneURL(p.HTTPURLToRepo, os.Getenv("GHORG_GITLAB_TOKEN"))
				r.URL = p.HTTPURLToRepo
				cloneData = append(cloneData, r)
			} else {
				r.CloneURL = p.SSHURLToRepo
				r.URL = p.SSHURLToRepo
				cloneData = append(cloneData, r)
			}
		}

		// Exit the loop when we've seen all pages.
		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		// Update the page number to get the next page.
		opt.Page = resp.NextPage
	}

	return cloneData, nil
}

func addTokenToHTTPSCloneURL(url string, token string) string {
	splitURL := strings.Split(url, "https://")
	return "https://oauth2:" + token + "@" + splitURL[1]
}
