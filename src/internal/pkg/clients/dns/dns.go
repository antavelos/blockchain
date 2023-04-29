package clientdns

import (
	nd "github.com/antavelos/blockchain/src/internal/pkg/models/node"
	"github.com/antavelos/blockchain/src/pkg/rest"
)

const nodesEndpoint = "/nodes"

func GetDNSNodes(host string) ([]nd.Node, error) {
	requester := rest.GetRequester{
		URL: host + nodesEndpoint,
	}

	response := requester.Request()

	if response.Err != nil {
		return nil, response.Err
	}

	return nd.UnmarshalMany(response.Body)
}

func AddDNSNode(host string, node nd.Node) error {
	requester := rest.PostRequester{
		URL:  host + nodesEndpoint,
		Body: node,
	}

	response := requester.Request()

	return response.Err
}
