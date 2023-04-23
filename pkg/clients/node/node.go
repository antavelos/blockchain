package clientnode

import (
	"github.com/antavelos/blockchain/pkg/lib/rest"
	bc "github.com/antavelos/blockchain/pkg/models/blockchain"
	nd "github.com/antavelos/blockchain/pkg/models/node"
)

const sharedTransactionsEndpoint = "/shared-transactions"
const sharedBlocksEndpoint = "/shared-blocks"
const pingEndpoint = "/ping"
const blockchainEndpoint = "/blockchain"
const transactionsEndpoint = "/transactions"

func ShareTx(nodes []nd.Node, tx bc.Transaction) rest.BulkResponse {
	var requesters []rest.Requester
	for _, node := range nodes {
		requester := rest.PostRequester{
			URL:  node.GetHost() + sharedTransactionsEndpoint,
			Body: tx,
		}
		requesters = append(requesters, requester)
	}

	return rest.BulkRequest(requesters)
}

func ShareBlock(nodes []nd.Node, block bc.Block) rest.BulkResponse {
	var requesters []rest.Requester
	for _, node := range nodes {
		requester := rest.PostRequester{
			URL:  node.GetHost() + sharedBlocksEndpoint,
			Body: block,
		}
		requesters = append(requesters, requester)
	}

	return rest.BulkRequest(requesters)
}

func PingNodes(nodes []nd.Node, selfNode nd.Node) rest.BulkResponse {
	var requesters []rest.Requester
	for _, node := range nodes {
		requester := rest.PostRequester{
			URL:  node.GetHost() + pingEndpoint,
			Body: selfNode,
		}
		requesters = append(requesters, requester)
	}

	return rest.BulkRequest(requesters)
}

func GetBlockchains(nodes []nd.Node) rest.BulkResponse {
	var requesters []rest.Requester
	for _, node := range nodes {
		requester := rest.GetRequester{
			URL: node.GetHost() + blockchainEndpoint,
		}
		requesters = append(requesters, requester)
	}

	return rest.BulkRequest(requesters)
}

func SendTransaction(node nd.Node, tx bc.Transaction) (bc.Transaction, error) {
	requester := rest.PostRequester{
		URL:  node.GetHost() + transactionsEndpoint,
		Body: tx,
	}

	response := requester.Request()
	if response.Err != nil {
		return bc.Transaction{}, response.Err
	}

	return bc.UnmarshalTransaction(response.Body)
}
