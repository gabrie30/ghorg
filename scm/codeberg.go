package scm

import (
	"os"
)

// compile-time assertion that Codeberg implements the Client interface
var _ Client = Codeberg{}

func init() {
	registerClient(Codeberg{})
}

// Codeberg is backed by the Gitea client. Codeberg runs Forgejo, which is
// API-compatible with Gitea, so the entire Gitea backend is reused. Codeberg
// differs by defaulting the base URL to https://codeberg.org and using its own
// GHORG_CODEBERG_TOKEN.
type Codeberg struct {
	Gitea
}

func (Codeberg) GetType() string {
	return "codeberg"
}

// NewClient creates a new Codeberg scm client backed by the Gitea client,
// defaulting the base URL to Codeberg's public instance when none is provided.
func (Codeberg) NewClient() (Client, error) {
	baseURL := os.Getenv("GHORG_SCM_BASE_URL")
	if baseURL == "" {
		baseURL = "https://codeberg.org"
	}

	token := os.Getenv("GHORG_CODEBERG_TOKEN")
	insecure := os.Getenv("GHORG_INSECURE_CODEBERG_CLIENT") == "true"

	c, err := newGiteaClient(baseURL, token, insecure, "--insecure-codeberg-client")
	if err != nil {
		return nil, err
	}

	return Codeberg{Gitea: c.(Gitea)}, nil
}
