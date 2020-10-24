package scm

import "fmt"

// Client define the interface a scm client has to have
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

// SupportedClients return list of all supported clients
func SupportedClients() []string {
	types := make([]string, 0, len(clients))
	for i := range clients {
		types = append(types, clients[i].GetType())
	}
	return types
}
