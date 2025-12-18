package scm

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/ktrysmt/go-bitbucket"
)

var (
	_ Client = Bitbucket{}
)

func init() {
	registerClient(Bitbucket{})
}

type Bitbucket struct {
	// extend the bitbucket client
	*bitbucket.Client
	// Fields for Bitbucket Server support
	isServer   bool
	serverURL  string
	httpClient *http.Client
	username   string
	password   string
	// Fields for Bitbucket Cloud API Token support
	// When using API tokens, git operations use x-bitbucket-api-token-auth as the username
	useAPIToken bool
	apiToken    string
}

func (_ Bitbucket) GetType() string {
	return "bitbucket"
}

// GetOrgRepos gets org repos
func (c Bitbucket) GetOrgRepos(targetOrg string) ([]Repo, error) {
	spinningSpinner.Start()
	defer spinningSpinner.Stop()

	if c.isServer {
		return c.getServerProjectRepos(targetOrg)
	}

	// Use Cloud API (existing logic)
	resp, err := c.Repositories.ListForAccount(&bitbucket.RepositoriesOptions{Owner: targetOrg})
	if err != nil {
		return []Repo{}, err
	}

	return c.filter(resp.Items)
}

// GetUserRepos gets user repos from bitbucket
func (c Bitbucket) GetUserRepos(targetUser string) ([]Repo, error) {
	if c.isServer {
		return c.getServerUserRepos(targetUser)
	}

	// Use Cloud API (existing logic)
	resp, err := c.Repositories.ListForAccount(&bitbucket.RepositoriesOptions{Owner: targetUser})
	if err != nil {
		return []Repo{}, err
	}

	return c.filter(resp.Items)
}

// NewClient create new bitbucket scm client
func (_ Bitbucket) NewClient() (Client, error) {
	user := os.Getenv("GHORG_BITBUCKET_USERNAME")
	password := os.Getenv("GHORG_BITBUCKET_APP_PASSWORD")
	oAuth := os.Getenv("GHORG_BITBUCKET_OAUTH_TOKEN")
	apiToken := os.Getenv("GHORG_BITBUCKET_API_TOKEN")
	apiEmail := os.Getenv("GHORG_BITBUCKET_API_EMAIL")
	baseURL := os.Getenv("GHORG_SCM_BASE_URL")

	// Check if this is a Bitbucket Server instance
	isServer := baseURL != ""

	if isServer {
		// For Bitbucket Server, create a custom client
		httpClient := &http.Client{}

		// Handle insecure connections
		if strings.HasPrefix(baseURL, "http://") && os.Getenv("GHORG_INSECURE_BITBUCKET_CLIENT") != "true" {
			colorlog.PrintErrorAndExit("You are attempting to clone from an insecure Bitbucket instance. You must set GHORG_INSECURE_BITBUCKET_CLIENT environment variable to 'true' to proceed.")
		}

		if os.Getenv("GHORG_INSECURE_BITBUCKET_CLIENT") == "true" {
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			httpClient = &http.Client{Transport: tr}
			colorlog.PrintError("WARNING: USING AN INSECURE BITBUCKET CLIENT")
		}

		// Configured Bitbucket client for self-hosted instance

		return Bitbucket{
			Client:     nil, // Not using the Cloud client
			isServer:   true,
			serverURL:  baseURL,
			httpClient: httpClient,
			username:   user,
			password:   password,
		}, nil
	}

	// For Bitbucket Cloud, use the existing go-bitbucket library
	var c *bitbucket.Client
	var useAPIToken bool

	if oAuth != "" {
		// OAuth bearer token takes precedence
		c = bitbucket.NewOAuthbearerToken(oAuth)
	} else if apiToken != "" {
		// New API Token authentication
		// For API calls, use email (or username as fallback) with the API token
		apiUser := apiEmail
		if apiUser == "" {
			apiUser = user // Fall back to username if email not provided
		}
		c = bitbucket.NewBasicAuth(apiUser, apiToken)
		useAPIToken = true
	} else {
		// Legacy App Password authentication
		c = bitbucket.NewBasicAuth(user, password)
	}

	return Bitbucket{
		Client:      c,
		isServer:    false,
		useAPIToken: useAPIToken,
		apiToken:    apiToken,
	}, nil
}

// Bitbucket Server API structures
type ServerRepository struct {
	Name    string         `json:"name"`
	Slug    string         `json:"slug"`
	Links   map[string]any `json:"links"`
	Project struct {
		Key string `json:"key"`
	} `json:"project"`
}

type ServerProjectResponse struct {
	Values     []ServerRepository `json:"values"`
	Size       int                `json:"size"`
	IsLastPage bool               `json:"isLastPage"`
	Start      int                `json:"start"`
}

// getServerProjectRepos gets repositories from a Bitbucket Server project with parallel pagination
func (c Bitbucket) getServerProjectRepos(projectKey string) ([]Repo, error) {
	apiURL := strings.TrimSuffix(c.serverURL, "/") + fmt.Sprintf("/rest/api/1.0/projects/%s/repos", projectKey)
	limit := 25

	// Fetch first page to discover total size
	url := fmt.Sprintf("%s?start=0&limit=%d", apiURL, limit)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		colorlog.PrintError(fmt.Sprintf("API request failed with status %d: %s", resp.StatusCode, string(body)))
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response ServerProjectResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %v", err)
	}

	// If only one page, return immediately
	if response.IsLastPage {
		return c.filterServerRepos(response.Values), nil
	}

	// Multiple pages - fetch remaining pages in parallel
	return c.fetchServerProjectReposParallel(projectKey, response.Values, response.Size, limit)
}

// getServerUserRepos gets repositories for a specific user (personal repositories) with parallel pagination
func (c Bitbucket) getServerUserRepos(username string) ([]Repo, error) {
	// For Bitbucket Server, user repos are typically in projects prefixed with ~username
	apiURL := strings.TrimSuffix(c.serverURL, "/") + "/rest/api/1.0/repos"
	limit := 25

	// Fetch first page to discover total size
	url := fmt.Sprintf("%s?start=0&limit=%d", apiURL, limit)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response ServerProjectResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	// If only one page, return immediately
	if response.IsLastPage {
		return c.filterServerRepos(response.Values), nil
	}

	// Multiple pages - fetch remaining pages in parallel
	return c.fetchServerUserReposParallel(username, response.Values, response.Size, limit)
}

// filterServerRepos converts Bitbucket Server repo format to ghorg Repo format
func (c Bitbucket) filterServerRepos(repos []ServerRepository) []Repo {
	cloneData := []Repo{}
	// Starting to filter repositories from Bitbucket Server

	for _, repo := range repos {
		if repo.Links != nil && repo.Links["clone"] != nil {
			cloneLinks, ok := repo.Links["clone"].([]any)
			if !ok {
				continue
			}

			for _, linkInterface := range cloneLinks {
				link, ok := linkInterface.(map[string]any)
				if !ok {
					continue
				}

				href, ok := link["href"].(string)
				if !ok {
					continue
				}

				name, ok := link["name"].(string)
				if !ok {
					continue
				}

				r := Repo{
					Name: repo.Name,
					Path: fmt.Sprintf("%s/%s", repo.Project.Key, repo.Slug),
					URL:  href,
				}

				// Set clone branch to default (master/main)
				if os.Getenv("GHORG_BRANCH") == "" {
					r.CloneBranch = "master" // Default for Bitbucket Server
				} else {
					r.CloneBranch = os.Getenv("GHORG_BRANCH")
				}

				// Handle different protocol types with flexible matching
				cloneProtocol := os.Getenv("GHORG_CLONE_PROTOCOL")
				if cloneProtocol == "" {
					cloneProtocol = "https" // Default to HTTPS
				}

				// Processing clone link

				if cloneProtocol == "ssh" && name == "ssh" {
					r.CloneURL = href
					cloneData = append(cloneData, r)
					// Added SSH clone URL
				} else if cloneProtocol == "https" && (name == "http" || name == "https") {
					// For HTTPS with basic auth, embed credentials in URL
					r.CloneURL = c.addCredentialsToURL(href)
					cloneData = append(cloneData, r)
					// Added HTTPS clone URL
				} else {
					// Log unmatched protocols for debugging
					// Skipping incompatible clone link
				}
			}
		}
	}

	// Filtering complete, repositories ready for cloning
	return cloneData
}

// addCredentialsToURL adds basic auth credentials to HTTPS URLs for cloning
func (c Bitbucket) addCredentialsToURL(cloneURL string) string {
	if c.username != "" && c.password != "" {
		// Insert credentials into HTTPS URL
		if strings.HasPrefix(cloneURL, "https://") {
			return strings.Replace(cloneURL, "https://", fmt.Sprintf("https://%s:%s@", c.username, c.password), 1)
		} else if strings.HasPrefix(cloneURL, "http://") {
			return strings.Replace(cloneURL, "http://", fmt.Sprintf("http://%s:%s@", c.username, c.password), 1)
		}
	}
	return cloneURL
}

func (c Bitbucket) filter(resp []bitbucket.Repository) (repoData []Repo, err error) {
	cloneData := []Repo{}

	for _, a := range resp {
		links := a.Links["clone"].([]any)
		for _, l := range links {
			link := l.(map[string]any)["href"]
			linkType := l.(map[string]any)["name"]

			if os.Getenv("GHORG_TOPICS") != "" {
				colorlog.PrintError("WARNING: Filtering by topics is not supported for Bitbucket SCM")
			}

			r := Repo{}
			r.Name = a.Name
			r.Path = a.Full_name
			if os.Getenv("GHORG_BRANCH") == "" {
				r.CloneBranch = a.Mainbranch.Name
			} else {
				r.CloneBranch = os.Getenv("GHORG_BRANCH")
			}

			if os.Getenv("GHORG_CLONE_PROTOCOL") == "ssh" && linkType == "ssh" {
				r.URL = link.(string)
				r.CloneURL = link.(string)
				cloneData = append(cloneData, r)
			} else if os.Getenv("GHORG_CLONE_PROTOCOL") == "https" && linkType == "https" {
				r.URL = link.(string)
				r.CloneURL = link.(string)
				if os.Getenv("GHORG_BITBUCKET_OAUTH_TOKEN") != "" {
					// OAuth token - no credentials needed in URL, handled by git credential helper
				} else if c.useAPIToken {
					// API Token authentication - use special username x-bitbucket-api-token-auth
					r.CloneURL = insertAPITokenCredentialsIntoURL(r.CloneURL, c.apiToken)
				} else {
					// Legacy App Password authentication
					r.CloneURL = insertAppPasswordCredentialsIntoURL(r.CloneURL)
				}
				cloneData = append(cloneData, r)
			}
		}
	}

	return cloneData, nil
}

// insertAppPasswordCredentialsIntoURL inserts app password credentials into the clone URL
// Uses the format: https://username:app_password@bitbucket.org/...
func insertAppPasswordCredentialsIntoURL(url string) string {
	credentials := ":" + os.Getenv("GHORG_BITBUCKET_APP_PASSWORD") + "@"
	urlWithCredentials := strings.Replace(url, "@", credentials, 1)

	return urlWithCredentials
}

// insertAPITokenCredentialsIntoURL inserts API token credentials into the clone URL
// Uses the special username x-bitbucket-api-token-auth as per Bitbucket's API token documentation
// Format: https://x-bitbucket-api-token-auth:api_token@bitbucket.org/...
func insertAPITokenCredentialsIntoURL(cloneURL string, apiToken string) string {
	// Bitbucket Cloud clone URLs are in the format: https://username@bitbucket.org/workspace/repo.git
	// We need to replace username@ with x-bitbucket-api-token-auth:token@
	if strings.HasPrefix(cloneURL, "https://") {
		// Find the @ symbol that separates username from host
		atIndex := strings.Index(cloneURL[8:], "@") // Skip "https://"
		if atIndex != -1 {
			// Replace everything between https:// and @ with the API token credentials
			return "https://x-bitbucket-api-token-auth:" + apiToken + cloneURL[8+atIndex:]
		}
		// If no @ found, insert credentials after https://
		return strings.Replace(cloneURL, "https://", "https://x-bitbucket-api-token-auth:"+apiToken+"@", 1)
	}
	return cloneURL
}
