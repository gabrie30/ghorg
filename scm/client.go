package scm

import "fmt"

type Client interface {
	NewClient() (Client, error)

	GetUserRepos(targetUsername string) ([]Repo, error)
	GetOrgRepos(targetOrg string) ([]Repo, error)

	GetType() string
}

var (
	clients []Client
)

// registerClient registers a client
func registerClient(c Client) {
	clients = append(clients, c)
}

func GetClient(cType string) (Client, error) {
	for i := range clients {
		if clients[i].GetType() == cType {
			return clients[i].NewClient()
		}
	}
	return nil, fmt.Errorf("client type '%s' unsupported", cType)
}
