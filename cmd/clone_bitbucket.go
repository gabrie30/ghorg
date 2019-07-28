package cmd

import (
	bitbucket "github.com/ktrysmt/go-bitbucket"

	"os"
)

func getBitBucketOrgCloneUrls() ([]string, error) {

	client := bitbucket.NewBasicAuth(os.Getenv("GHORG_BITBUCKET_USERNAME"), os.Getenv("GHORG_BITBUCKET_APP_PASSWORD"))
	cloneUrls := []string{}

	resp, err := client.Users.Repositories(args[0])
	if err != nil {
		return []string{}, err
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

			if os.Getenv("GHORG_CLONE_PROTOCOL") == "ssh" && linkType == "ssh" {
				cloneUrls = append(cloneUrls, link.(string))
			} else if os.Getenv("GHORG_CLONE_PROTOCOL") == "https" && linkType == "https" {
				cloneUrls = append(cloneUrls, link.(string))
			}
		}
	}

	return cloneUrls, nil
}

func getBitBucketUserCloneUrls() ([]string, error) {

	client := bitbucket.NewBasicAuth(os.Getenv("GHORG_BITBUCKET_USERNAME"), os.Getenv("GHORG_BITBUCKET_APP_PASSWORD"))
	cloneUrls := []string{}

	resp, err := client.Teams.Repositories(args[0])
	if err != nil {
		return []string{}, err
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

			if os.Getenv("GHORG_CLONE_PROTOCOL") == "ssh" && linkType == "ssh" {
				cloneUrls = append(cloneUrls, link.(string))
			} else if os.Getenv("GHORG_CLONE_PROTOCOL") == "https" && linkType == "https" {
				cloneUrls = append(cloneUrls, link.(string))
			}
		}
	}

	return cloneUrls, nil
}
