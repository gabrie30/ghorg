package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/gabrie30/ghorg/colorlog"
	gitlab "github.com/xanzy/go-gitlab"
)

func getGitLabOrgCloneUrls() ([]string, error) {
	cloneUrls := []string{}
	client := gitlab.NewClient(nil, os.Getenv("GHORG_GITLAB_TOKEN"))
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
			return []string{}, err
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

			if os.Getenv("GHORG_CLONE_PROTOCOL") == "https" {
				cloneUrls = append(cloneUrls, p.HTTPURLToRepo)
			} else {
				cloneUrls = append(cloneUrls, p.SSHURLToRepo)
			}
		}

		// Exit the loop when we've seen all pages.
		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		// Update the page number to get the next page.
		opt.Page = resp.NextPage
	}

	return cloneUrls, nil
}

// func getUsersUID(username string) int {

// }

func getGitLabUserCloneUrls() ([]string, error) {
	cloneUrls := []string{}
	client := gitlab.NewClient(nil, os.Getenv("GHORG_GITLAB_TOKEN"))

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
			return []string{}, err
		}

		// List all the projects we've found so far.
		for _, p := range ps {
			if os.Getenv("GHORG_CLONE_PROTOCOL") == "https" {
				cloneUrls = append(cloneUrls, p.HTTPURLToRepo)
			} else {
				cloneUrls = append(cloneUrls, p.SSHURLToRepo)
			}
		}

		// Exit the loop when we've seen all pages.
		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		// Update the page number to get the next page.
		opt.Page = resp.NextPage
	}

	return cloneUrls, nil
}
