package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/gabrie30/ghorg/colorlog"
	gitlab "github.com/xanzy/go-gitlab"
)

func getGitLabOrgCloneUrls() ([]Repo, error) {
	repoData := []Repo{}
	client, err := determineClient()

	if err != nil {
		colorlog.PrintError(err)
	}

	namespace := os.Getenv("GHORG_GITLAB_DEFAULT_NAMESPACE")

	opt := &gitlab.ListGroupProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 50,
			Page:    1,
		},
		IncludeSubgroups: gitlab.Bool(true),
	}

	if namespace == "unset" {
		colorlog.PrintInfo("No namespace set, to reduce results use namespace flag e.g. --namespace=gitlab-org/security-products")
		fmt.Println("")
	}

	for {
		// Get the first page with projects.
		ps, resp, err := client.Groups.ListGroupProjects(args[0], opt)

		if err != nil {
			// TODO: check if 404, then we know group does not exist
			return []Repo{}, err
		}

		// List all the projects we've found so far.
		for _, p := range ps {

			// If it is set, then filter only repos from the namespace
			// if p.PathWithNamespace == "the namespace the user indicated" eg --namespace=org/namespace

			if namespace != "unset" {
				if strings.HasPrefix(p.PathWithNamespace, strings.ToLower(namespace)) == false {
					continue
				}
			}

			if os.Getenv("GHORG_SKIP_ARCHIVED") == "true" {
				if p.Archived == true {
					continue
				}
			}
			r := Repo{}

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

func getGitLabUserCloneUrls() ([]Repo, error) {
	cloneData := []Repo{}

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
		ps, resp, err := client.Projects.ListUserProjects(args[0], opt)
		if err != nil {
			// TODO: check if 404, then we know user does not exist
			return []Repo{}, err
		}

		// List all the projects we've found so far.
		for _, p := range ps {

			if os.Getenv("GHORG_SKIP_ARCHIVED") == "true" {
				if p.Archived == true {
					continue
				}
			}
			r := Repo{}
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
