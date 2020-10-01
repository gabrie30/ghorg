package internal

import "github.com/gabrie30/ghorg/internal/base"

var (
	clients []base.Client
)

// RegisterClient registers a client
func RegisterClient(c base.Client) {
	clients = append(clients, c)
}

func GetClient(cType string) base.Client {
	for i := range clients {
		if clients[i].GetType() == cType {
			return clients[i]
		}
	}
	return nil
}
