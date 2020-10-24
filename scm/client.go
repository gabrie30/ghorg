package scm

type Client interface {
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

func GetClient(cType string) Client {
	for i := range clients {
		if clients[i].GetType() == cType {
			return clients[i]
		}
	}
	return nil
}
