package cmd

import (
	"os"

	gitlab "github.com/xanzy/go-gitlab"
)

func getGitLabOrgCloneUrls() ([]string, error) {
	cloneUrls := []string{}
	client := gitlab.NewClient(nil, os.Getenv("GHORG_GITLAB_TOKEN"))

	opt := &gitlab.ListGroupProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 50,
			Page:    1,
		},
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
