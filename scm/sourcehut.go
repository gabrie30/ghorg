package scm

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/gabrie30/ghorg/colorlog"
)

var (
	_ Client = Sourcehut{}
)

func init() {
	registerClient(Sourcehut{})
}

type Sourcehut struct {
	Client  *http.Client
	Token   string
	BaseURL string
}

type sourcehutCursor string

func (_ Sourcehut) GetType() string {
	return "sourcehut"
}

// // The ID of the repo that is assigned via the SCM provider. This is used for example with gitlab snippets on cloud gropus where we need to know the repo id to look up all he snippets it has.
// ID string
// // Name is the name of the repo https://www.github.com/gabrie30/ghorg.git the Name would be ghorg
// Name string
// // Path where the repo is located within the scm provider. Its mostly used with gitlab repos when the directory structure is preserved. In this case the path becomes where to locate the repo in relation to gitlab.com/group/group/group/repo.git => /group/group/group/repo
// Path string
// // URL is the web address of the repo
// URL string
// // CloneURL is the url for cloning the repo, will be different for ssh vs http clones and will have the .git extention
// CloneURL string
// // CloneBranch the branch to clone. This will be the default branch if not specified. It will always be main for snippets.
// CloneBranch string
// // IsWiki is set to true when the data is for a wiki page

// GetOrgRepos fetches repo data from a specific group. We emulate this by checking if the repo's owner
// matches the current user, which is identical to GetUserRepos. It's possible in the future that this
// will change as the manual (https://docs.sourcehut.org/git.sr.ht/#field-repositories) states that
// it will. This can be addressed once the changed behaviour is observable, rather than speculative.
func (c Sourcehut) GetOrgRepos(targetOrg string) ([]Repo, error) {
	return c.GetUserRepos(targetOrg)
}

// GetUserRepos gets all of a users sourcehut repos
func (c Sourcehut) GetUserRepos(targetUsername string) ([]Repo, error) {
	spinningSpinner.Start()
	defer spinningSpinner.Stop()

	repos := []Repo{}

	var cursor sourcehutCursor
	for {
		reposPage, nextCursor, err := c.queryRepositoriesPage(cursor, targetUsername)
		if err != nil {
			return nil, err
		}
		cursor = nextCursor
		repos = append(repos, reposPage...)
		if cursor == "" {
			break
		}
	}

	return repos, nil
}

// NewClient create new sourcehut scm client
func (_ Sourcehut) NewClient() (Client, error) {
	baseURL := os.Getenv("GHORG_SCM_BASE_URL")
	token := os.Getenv("GHORG_SOURCEHUT_TOKEN")

	if baseURL == "" {
		baseURL = "https://git.sr.ht"
	}

	isHTTP := strings.HasPrefix(baseURL, "http://")

	if isHTTP && (os.Getenv("GHORG_INSECURE_SOURCEHUT_CLIENT") != "true") {
		colorlog.PrintErrorAndExit("You are attempting clone from an insecure sourcehut instance. You must set the (--insecure-sourcehut-client) flag to proceed.")
	}

	var hc *http.Client
	if os.Getenv("GHORG_INSECURE_SOURCEHUT_CLIENT") == "true" {
		defaultTransport := http.DefaultTransport.(*http.Transport)
		// Create new Transport that ignores self-signed SSL
		customTransport := &http.Transport{
			Proxy:                 defaultTransport.Proxy,
			DialContext:           defaultTransport.DialContext,
			MaxIdleConns:          defaultTransport.MaxIdleConns,
			IdleConnTimeout:       defaultTransport.IdleConnTimeout,
			ExpectContinueTimeout: defaultTransport.ExpectContinueTimeout,
			TLSHandshakeTimeout:   defaultTransport.TLSHandshakeTimeout,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		}
		hc = &http.Client{Transport: customTransport}
		colorlog.PrintError("WARNING: USING AN INSECURE SOURCEHUT CLIENT")
	} else {
		hc = &http.Client{}
	}

	client := Sourcehut{
		BaseURL: baseURL,
		Client:  hc,
		Token:   token,
	}

	return client, nil
}

var sourcehutReposQuery = `
query repositories($cursor: Cursor, $filter: Filter) {
  repositories(cursor: $cursor, filter: $filter) {
    results {
      id
      name
	  visibility
      owner { canonicalName }
      HEAD { name }
    }
    cursor
  }
}
`

func (c Sourcehut) queryRepositoriesPage(cursor sourcehutCursor, targetUser string) ([]Repo, sourcehutCursor, error) {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, "", err
	}
	u.Path = "/query"

	vars := map[string]any{}
	if cursor != "" {
		vars["cursor"] = cursor
	}

	inputBody, err := json.Marshal(map[string]any{
		"query":     sourcehutReposQuery,
		"variables": vars,
	})
	if err != nil {
		return nil, "", err
	}

	rq, err := http.NewRequest("POST", u.String(), bytes.NewReader(inputBody))
	if err != nil {
		return nil, "", err
	}
	rq.Header.Set("Authorization", fmt.Sprintf("bearer %s", c.Token))
	rq.Header.Set("Content-Type", "application/json")
	rq.Header.Set("Accept", "application/json")

	rs, err := c.Client.Do(rq)
	if err != nil {
		return nil, "", err
	}
	defer rs.Body.Close()

	if rs.StatusCode != 200 {
		body, _ := io.ReadAll(io.LimitReader(rs.Body, 200))
		return nil, "", fmt.Errorf("unexpected response code %d from sourcehut: %q", rs.StatusCode, string(body))
	}

	type repository struct {
		ID    int64 `json:"id"`
		Owner struct {
			ID            string `json:"id"`
			CanonicalName string `json:"canonicalName"`
		} `json:"owner"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Visibility  string `json:"visibility"`
		HEAD        struct {
			Name   string `json:"name"`
			Target string `json:"target"`
		} `json:"HEAD"`
	}

	var response struct {
		Errors json.RawMessage `json:"errors"`
		Data   struct {
			Repositories struct {
				Results []repository `json:"results"`
				Cursor  string       `json:"cursor"`
			} `json:"repositories"`
		} `json:"data"`
	}

	body, err := io.ReadAll(rs.Body)
	if err != nil {
		return nil, "", err
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, "", err
	}
	if response.Errors != nil {
		return nil, "", fmt.Errorf("sourcehut api returned errors while listing repos: %s", string(response.Errors))
	}

	var repos []Repo
	for _, rp := range response.Data.Repositories.Results {
		if rp.Owner.CanonicalName != targetUser {
			continue
		}

		r := Repo{}
		r.Path = path.Join(rp.Owner.CanonicalName, rp.Name)
		r.Name = rp.Name
		r.Slug = rp.Name

		if os.Getenv("GHORG_BRANCH") == "" {
			var defaultBranch = ""
			if strings.HasPrefix(rp.HEAD.Name, "refs/heads/") {
				defaultBranch = rp.HEAD.Name[len("refs/heads/"):]
			}
			if defaultBranch == "" {
				defaultBranch = "master"
			}
			r.CloneBranch = defaultBranch
		} else {
			r.CloneBranch = os.Getenv("GHORG_BRANCH")
		}

		if rp.Visibility == "PUBLIC" || rp.Visibility == "UNLISTED" {
			r.CloneURL = fmt.Sprintf("%s/%s", c.BaseURL, r.Path)

		} else {
			var gitBase string
			isHTTP := strings.HasPrefix(c.BaseURL, "http://")
			if isHTTP {
				gitBase = strings.Replace(c.BaseURL, "http://", "git@", 1)
			} else {
				gitBase = strings.Replace(c.BaseURL, "https://", "git@", 1)
			}
			r.CloneURL = fmt.Sprintf("%s:%s", gitBase, r.Path)
		}

		r.URL = r.CloneURL
		repos = append(repos, r)
	}

	return repos, sourcehutCursor(response.Data.Repositories.Cursor), nil
}
