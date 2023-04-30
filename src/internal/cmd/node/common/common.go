package common

import (
	"fmt"

	cfg "github.com/antavelos/blockchain/src/internal/cmd/node/config"
	dns_client "github.com/antavelos/blockchain/src/internal/pkg/clients/dns"
	node_client "github.com/antavelos/blockchain/src/internal/pkg/clients/node"
	wallet_client "github.com/antavelos/blockchain/src/internal/pkg/clients/wallet"
	bc "github.com/antavelos/blockchain/src/internal/pkg/models/blockchain"
	nd "github.com/antavelos/blockchain/src/internal/pkg/models/node"
	bc_repo "github.com/antavelos/blockchain/src/internal/pkg/repos/blockchain"
	node_repo "github.com/antavelos/blockchain/src/internal/pkg/repos/node"
	wallet_repo "github.com/antavelos/blockchain/src/internal/pkg/repos/wallet"
	"github.com/antavelos/blockchain/src/pkg/rest"
	"github.com/antavelos/blockchain/src/pkg/utils"
)

type CommonHandler struct {
	Config         *cfg.Config
	BlockchainRepo *bc_repo.BlockchainRepo
	NodeRepo       *node_repo.NodeRepo
	WalletRepo     *wallet_repo.WalletRepo
}

func NewCommonHandler(config *cfg.Config, br *bc_repo.BlockchainRepo, nr *node_repo.NodeRepo, wr *wallet_repo.WalletRepo) *CommonHandler {
	return &CommonHandler{Config: config, BlockchainRepo: br, NodeRepo: nr, WalletRepo: wr}
}

func getUrl(host string, port string) string {
	return fmt.Sprintf("http://%v:%v", host, port)
}

func (h CommonHandler) getDNSHost() string {
	return getUrl(h.Config.Get("DNS_HOST"), h.Config.Get("DNS_PORT"))
}

func (h CommonHandler) getWalletsHost() string {
	return getUrl(h.Config.Get("WALLETS_HOST"), h.Config.Get("WALLETS_PORT"))
}

func (h CommonHandler) getSelfNode() (nd.Node, error) {
	ip, err := utils.GetSelfIP()
	if err != nil {
		return nd.Node{}, err
	}

	return nd.NewNode(h.Config.Get("NODE_NAME"), ip, h.Config.Get("PORT")), nil
}

func (h CommonHandler) InitNode() error {
	err := h.IntroduceToDNS()
	if err != nil {
		return utils.GenericError{Msg: "failed to introduce self at DNS", Extra: err}
	}

	err = h.RefreshDNSNodes()
	if err != nil {
		return utils.GenericError{Msg: "retrieve DNS nodes error", Extra: err}
	}

	err = h.PingNodes()
	if err != nil {
		utils.LogError("ping nodes error", err.Error())
	}

	h.ResolveLongestBlockchain()

	if h.WalletRepo.IsEmpty() {
		if err := h.CreateNewWallet(); err != nil {
			return utils.GenericError{Msg: "failed to create new wallet", Extra: err}
		}
	}

	return nil
}

func (h CommonHandler) IntroduceToDNS() error {
	selfNode, err := h.getSelfNode()
	if err != nil {
		return err
	}

	return dns_client.AddDNSNode(h.getDNSHost(), selfNode)
}

func (h CommonHandler) RefreshDNSNodes() error {
	nodes, err := dns_client.GetDNSNodes(h.getDNSHost())
	if err != nil {
		return utils.GenericError{Msg: "couldn't retrieve nodes from DNS", Extra: err}
	}

	nodes = utils.Filter(nodes, func(n nd.Node) bool {
		return n.Name != h.Config.Get("NODE_NAME")
	})

	for _, node := range nodes {
		if err := h.NodeRepo.AddNode(node); err != nil {
			return utils.GenericError{Msg: "couldn't save nodes received from DNS", Extra: err}
		}
	}

	return nil
}

func (h CommonHandler) PingNodes() error {
	nodes, err := h.NodeRepo.GetNodes()
	if err != nil {
		return utils.GenericError{Msg: "couldn't load nodes", Extra: err}
	}

	selfNode, err := h.getSelfNode()
	if err != nil {
		return err
	}

	responses := node_client.PingNodes(nodes, selfNode)

	// if responses.HasConnectionRefused() {
	// 	h.Bus.Handle(eventbus.DataEvent{Ev: events.ConnectionRefusedEvent})
	// }

	if responses.ErrorsRatio() < 1 {
		return utils.GenericError{
			Msg: fmt.Sprintf("failed to share the transaction with other nodes: %v", responses.Errors()),
		}
	}

	return nil
}

func (h CommonHandler) getBlockchains(nodes []nd.Node) []*bc.Blockchain {

	responses := node_client.GetBlockchains(nodes)

	// if responses.HasConnectionRefused() {
	// 	h.Bus.Handle(eventbus.DataEvent{Ev: events.ConnectionRefusedEvent})
	// }

	noErrorResponses := utils.Filter(responses, func(response rest.Response) bool {
		return response.Err == nil
	})

	blockchains := utils.Map(noErrorResponses, func(response rest.Response) *bc.Blockchain {
		blockchain, err := bc.UnmarshalBlockchain(response.Body)
		if err != nil {
			return &bc.Blockchain{}
		}

		return &blockchain
	})

	return blockchains
}

func (h CommonHandler) ResolveLongestBlockchain() error {
	nodes, err := h.NodeRepo.GetNodes()
	if err != nil {
		return err
	}

	// TODO: include the below in a single lock

	blockchains := h.getBlockchains(nodes)
	utils.LogInfo("Retrieved blockchains", len(blockchains))

	localBlockchain, _ := h.BlockchainRepo.GetBlockchain()
	blockchains = append(blockchains, localBlockchain)

	maxLengthBlockchain := bc.GetMaxLengthBlockchain(blockchains)

	if len(maxLengthBlockchain.Blocks) == 0 {
		return nil
	}

	err = h.BlockchainRepo.UpdateBlockchain(maxLengthBlockchain)
	if err != nil {
		return utils.GenericError{Msg: "failed to update local blockchain", Extra: err}
	}

	return nil
}

func (h CommonHandler) CreateNewWallet() error {
	wallet, err := wallet_client.GetNewWallet(h.getWalletsHost())
	if err != nil {
		return err
	}

	return h.WalletRepo.AddWallet(wallet)
}
