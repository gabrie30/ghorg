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

// normalizeUsername ensures the username has the ~ prefix for API calls
// Sourcehut usernames require ~ for API queries
func normalizeUsername(username string) string {
	if !strings.HasPrefix(username, "~") {
		return "~" + username
	}
	return username
}

// stripUsernamePrefix removes the ~ prefix from sourcehut usernames for local paths
// This makes paths cleaner and avoids shell expansion issues
func stripUsernamePrefix(username string) string {
	return strings.TrimPrefix(username, "~")
}

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

	// Normalize username for API (ensure ~ prefix)
	apiUsername := normalizeUsername(targetUsername)
	// Strip prefix for local paths
	localUsername := stripUsernamePrefix(targetUsername)

	var cursor sourcehutCursor
	for {
		reposPage, nextCursor, err := c.queryRepositoriesPage(cursor, apiUsername, localUsername)
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

func (c Sourcehut) queryRepositoriesPage(cursor sourcehutCursor, apiUsername string, localUsername string) ([]Repo, sourcehutCursor, error) {
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

	repos, err := c.filter(response.Data.Repositories.Results, apiUsername, localUsername)
	if err != nil {
		return nil, "", err
	}

	return repos, sourcehutCursor(response.Data.Repositories.Cursor), nil
}

func (c Sourcehut) filter(rps []repository, apiUsername string, localUsername string) ([]Repo, error) {
	var repos []Repo

	for _, rp := range rps {
		// Filter by owner (Sourcehut API doesn't support server-side owner filtering)
		// API returns usernames with ~ prefix
		if rp.Owner.CanonicalName != apiUsername {
			continue
		}

		// Note: Sourcehut doesn't expose archived/fork status via GraphQL API
		// If these fields become available, add filtering here like other SCM providers

		r := Repo{}
		// Use localUsername (without ~) for local paths to avoid shell expansion issues
		r.Path = path.Join(localUsername, rp.Name)
		r.Name = rp.Name

		// Build the repo path WITH ~ for clone URLs (git needs this)
		repoPathWithTilde := path.Join(rp.Owner.CanonicalName, rp.Name)

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

		// Determine clone URL based on protocol and visibility
		protocol := os.Getenv("GHORG_CLONE_PROTOCOL")
		if protocol == "" {
			protocol = "https" // default to https
		}

		if protocol == "https" {
			// Use repoPathWithTilde for clone URL (git needs the ~ prefix)
			r.CloneURL = fmt.Sprintf("%s/%s", c.BaseURL, repoPathWithTilde)
		} else {
			// SSH protocol
			var gitBase string
			isHTTP := strings.HasPrefix(c.BaseURL, "http://")
			if isHTTP {
				gitBase = strings.Replace(c.BaseURL, "http://", "git@", 1)
			} else {
				gitBase = strings.Replace(c.BaseURL, "https://", "git@", 1)
			}
			// Use repoPathWithTilde for clone URL (git needs the ~ prefix)
			r.CloneURL = fmt.Sprintf("%s:%s", gitBase, repoPathWithTilde)
		}

		r.URL = r.CloneURL
		repos = append(repos, r)
	}

	return repos, nil
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
