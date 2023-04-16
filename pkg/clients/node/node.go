package clientnode

import (
	"github.com/antavelos/blockchain/pkg/lib/rest"
	bc "github.com/antavelos/blockchain/pkg/models/blockchain"
	nd "github.com/antavelos/blockchain/pkg/models/node"
)

const sharedTransactionsEndpoint = "/shared-transactions"
const sharedBlocksEndpoint = "/shared-blocks"
const pingEndpoint = "/ping"
const blockchainEndpoint = "/blockachain"
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
			M:    nd.NodeMarshaller{Many: true},
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
			M:   bc.BlockchainMarshaller{Many: true},
		}
		requesters = append(requesters, requester)
	}

	return rest.BulkRequest(requesters)
}

func SendTransaction(node nd.Node, tx bc.Transaction) rest.Response {
	requester := rest.PostRequester{
		URL: node.GetHost() + transactionsEndpoint,
		M:   bc.TxMarshaller{},
	}

	return requester.Request()
}
