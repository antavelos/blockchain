package clientdns

import (
	"github.com/antavelos/blockchain/pkg/lib/rest"
	nd "github.com/antavelos/blockchain/pkg/models/node"
)

const nodesEndpoint = "/nodes"

func GetDnsNodes(host string) ([]nd.Node, error) {
	requester := rest.GetRequester{
		URL: host + nodesEndpoint,
		M:   nd.NodeMarshaller{Many: true},
	}

	response := requester.Request()

	return response.Body.([]nd.Node), response.Err
}

func AddDnsNode(host string, node nd.Node) error {
	requester := rest.PostRequester{
		URL:  host + nodesEndpoint,
		Body: node,
		M:    nd.NodeMarshaller{Many: true},
	}

	response := requester.Request()

	return response.Err
}
