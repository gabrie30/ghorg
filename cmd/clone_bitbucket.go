package cmd

import (
	bitbucket "github.com/ktrysmt/go-bitbucket"

	"os"
)

func getBitBucketOrgCloneUrls() ([]Repo, error) {

	client := bitbucket.NewBasicAuth(os.Getenv("GHORG_BITBUCKET_USERNAME"), os.Getenv("GHORG_BITBUCKET_APP_PASSWORD"))
	cloneData := []Repo{}

	resp, err := client.Teams.Repositories(args[0])
	if err != nil {
		return []Repo{}, err
	}
	values := resp.(map[string]interface{})["values"].([]interface{})
	if err != nil {
		return nil, err
	}
	for _, a := range values {
		clone := a.(map[string]interface{})
		links := clone["links"].(map[string]interface{})["clone"].([]interface{})
		for _, l := range links {
			link := l.(map[string]interface{})["href"]
			linkType := l.(map[string]interface{})["name"]
			r := Repo{}
			if os.Getenv("GHORG_CLONE_PROTOCOL") == "ssh" && linkType == "ssh" {
				r.URL = link.(string)
				cloneData = append(cloneData, r)
			} else if os.Getenv("GHORG_CLONE_PROTOCOL") == "https" && linkType == "https" {
				r.URL = link.(string)
				cloneData = append(cloneData, r)
			}
		}
	}

	return cloneData, nil
}

func getBitBucketUserCloneUrls() ([]Repo, error) {

	client := bitbucket.NewBasicAuth(os.Getenv("GHORG_BITBUCKET_USERNAME"), os.Getenv("GHORG_BITBUCKET_APP_PASSWORD"))
	cloneData := []Repo{}

	resp, err := client.Users.Repositories(args[0])
	if err != nil {
		return []Repo{}, err
	}
	values := resp.(map[string]interface{})["values"].([]interface{})
	if err != nil {
		return nil, err
	}
	for _, a := range values {
		clone := a.(map[string]interface{})
		links := clone["links"].(map[string]interface{})["clone"].([]interface{})
		for _, l := range links {
			link := l.(map[string]interface{})["href"]
			linkType := l.(map[string]interface{})["name"]

			r := Repo{}
			if os.Getenv("GHORG_CLONE_PROTOCOL") == "ssh" && linkType == "ssh" {
				r.URL = link.(string)
				cloneData = append(cloneData, r)
			} else if os.Getenv("GHORG_CLONE_PROTOCOL") == "https" && linkType == "https" {
				r.URL = link.(string)
				cloneData = append(cloneData, r)
			}
		}
	}

	return cloneData, nil
}
