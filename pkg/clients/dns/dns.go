package clientdns

import (
	"github.com/antavelos/blockchain/pkg/lib/rest"
	nd "github.com/antavelos/blockchain/pkg/models/node"
)

const nodesEndpoint = "/nodes"

func GetDnsNodes(host string) ([]nd.Node, error) {
	requester := rest.GetRequester{
		URL: host + nodesEndpoint,
	}

	response := requester.Request()

	if response.Err != nil {
		return nil, response.Err
	}

	return nd.UnmarshalMany(response.Body)
}

func AddDnsNode(host string, node nd.Node) error {
	requester := rest.PostRequester{
		URL:  host + nodesEndpoint,
		Body: node,
	}

	response := requester.Request()

	return response.Err
}
