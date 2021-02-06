package scm

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gabrie30/ghorg/configs"
)

// ErrIncorrectScmType indicates an unsupported scm type being used
var ErrIncorrectScmType = errors.New("GHORG_SCM or --scm must be one of " + strings.Join(SupportedClients(), ", "))

// Client define the interface a scm client has to have
type Client interface {
	NewClient(config *configs.Config) (Client, error)

	GetUserRepos(config *configs.Config, targetUsername string) ([]Repo, error)
	GetOrgRepos(config *configs.Config, targetOrg string) ([]Repo, error)

	GetType() string
}

var (
	clients []Client
)

// registerClient registers a client
func registerClient(c Client) {
	clients = append(clients, c)
}

func GetClient(config *configs.Config, cType string) (Client, error) {
	for i := range clients {
		if clients[i].GetType() == cType {
			return clients[i].NewClient(config)
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

// VerifyScmType makes sure flags are set to appropriate values
func VerifyScmType(config *configs.Config) error {
	if !isStringInSlice(config.ScmType, SupportedClients()) {
		return ErrIncorrectScmType
	}

	return nil
}

// isStringInSlice check if a string is in a given slice
func isStringInSlice(s string, sl []string) bool {
	for i := range sl {
		if sl[i] == s {
			return true
		}
	}
	return false
}
