package bitbucket

import (
	"encoding/json"
	"os"

	"github.com/k0kubun/pp"
)

type BranchRestrictions struct {
	c *Client
}

func (b *BranchRestrictions) Gets(bo *BranchRestrictionsOptions) (interface{}, error) {
	urlStr := b.c.requestUrl("/repositories/%s/%s/branch-restrictions", bo.Owner, bo.RepoSlug)
	return b.c.execute("GET", urlStr, "")
}

func (b *BranchRestrictions) Create(bo *BranchRestrictionsOptions) (interface{}, error) {
	data := b.buildBranchRestrictionsBody(bo)
	urlStr := b.c.requestUrl("/repositories/%s/%s/branch-restrictions", bo.Owner, bo.RepoSlug)
	return b.c.execute("POST", urlStr, data)
}

func (b *BranchRestrictions) Get(bo *BranchRestrictionsOptions) (interface{}, error) {
	urlStr := b.c.requestUrl("/repositories/%s/%s/branch-restrictions/%s", bo.Owner, bo.RepoSlug, bo.ID)
	return b.c.execute("GET", urlStr, "")
}

func (b *BranchRestrictions) Update(bo *BranchRestrictionsOptions) (interface{}, error) {
	data := b.buildBranchRestrictionsBody(bo)
	urlStr := b.c.requestUrl("/repositories/%s/%s/branch-restrictions/%s", bo.Owner, bo.RepoSlug, bo.ID)
	return b.c.execute("PUT", urlStr, data)
}

func (b *BranchRestrictions) Delete(bo *BranchRestrictionsOptions) (interface{}, error) {
	urlStr := b.c.requestUrl("/repositories/%s/%s/branch-restrictions/%s", bo.Owner, bo.RepoSlug, bo.ID)
	return b.c.execute("DELETE", urlStr, "")
}

type branchRestrictionsBody struct {
	Kind    string `json:"kind"`
	Pattern string `json:"pattern"`
	Links   struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
	Value  interface{}                   `json:"value"`
	ID     int                           `json:"id"`
	Users  []branchRestrictionsBodyUser  `json:"users"`
	Groups []branchRestrictionsBodyGroup `json:"groups"`
}

type branchRestrictionsBodyGroup struct {
	Name  string `json:"name"`
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Html struct {
			Href string `json:"href"`
		} `json:"html"`
		FullSlug string `json:"full_slug"`
		Members  int    `json:"members"`
		Slug     string `json:"slug"`
	} `json:"links"`
}

type branchRestrictionsBodyUser struct {
	Username     string `json:"username"`
	Website      string `json:"website"`
	Display_name string `json:"display_name"`
	UUID         string `json:"uuid"`
	Created_on   string `json:"created_on"`
	Links        struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Repositories struct {
			Href string `json:"href"`
		} `json:"repositories"`
		Html struct {
			Href string `json:"href"`
		} `json:"html"`
		Followers struct {
			Href string `json:"href"`
		} `json:"followers"`
		Avatar struct {
			Href string `json:"href"`
		} `json:"avatar"`
		Following struct {
			Href string `json:"href"`
		} `json:"following"`
	} `json:"links"`
}

func (b *BranchRestrictions) buildBranchRestrictionsBody(bo *BranchRestrictionsOptions) string {

	var users []branchRestrictionsBodyUser
	var groups []branchRestrictionsBodyGroup
	for _, u := range bo.Users {
		user := branchRestrictionsBodyUser{
			Username: u,
		}
		users = append(users, user)
	}
	for _, g := range bo.Groups {
		group := branchRestrictionsBodyGroup{
			Name: g,
		}
		groups = append(groups, group)
	}

	body := branchRestrictionsBody{
		Kind:    bo.Kind,
		Pattern: bo.Pattern,
		Users:   users,
		Groups:  groups,
		Value:   bo.Value,
	}

	data, err := json.Marshal(body)
	if err != nil {
		pp.Println(err)
		os.Exit(9)
	}

	return string(data)
}
