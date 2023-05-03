package eventhandlers

import (
	"fmt"

	cfg "github.com/antavelos/blockchain/src/internal/cmd/node/config"
	"github.com/antavelos/blockchain/src/internal/cmd/node/events"
	dns_client "github.com/antavelos/blockchain/src/internal/pkg/clients/dns"
	node_client "github.com/antavelos/blockchain/src/internal/pkg/clients/node"
	wallet_client "github.com/antavelos/blockchain/src/internal/pkg/clients/wallet"
	bc "github.com/antavelos/blockchain/src/internal/pkg/models/blockchain"
	nd "github.com/antavelos/blockchain/src/internal/pkg/models/node"
	rep "github.com/antavelos/blockchain/src/internal/pkg/repos"
	"github.com/antavelos/blockchain/src/pkg/eventbus"
	"github.com/antavelos/blockchain/src/pkg/rest"
	"github.com/antavelos/blockchain/src/pkg/utils"
)

type EventHandler struct {
	Bus    *eventbus.Bus
	Config *cfg.Config
	Repos  *rep.Repos
}

func getUrl(host string, port string) string {
	return fmt.Sprintf("http://%v:%v", host, port)
}

func (h EventHandler) getDNSHost() string {
	return getUrl(h.Config.Get("DNS_HOST"), h.Config.Get("DNS_PORT"))
}

func (h EventHandler) getWalletsHost() string {
	return getUrl(h.Config.Get("WALLETS_HOST"), h.Config.Get("WALLETS_PORT"))
}

func (h EventHandler) getSelfNode() (nd.Node, error) {
	ip, err := utils.GetSelfIP()
	if err != nil {
		return nd.Node{}, err
	}

	return nd.NewNode(h.Config.Get("NODE_NAME"), ip, h.Config.Get("PORT")), nil
}

func (h EventHandler) HandleInitNode(event eventbus.DataEvent) {
	err := h.initNode()
	if err != nil {
		utils.LogFatal("Failed to initialize node", err.Error())
	}
}

func (h EventHandler) initNode() error {
	err := h.introduceToDNS()
	if err != nil {
		return utils.GenericError{Msg: "failed to introduce self at DNS", Extra: err}
	}

	err = h.refreshDNSNodes()
	if err != nil {
		return utils.GenericError{Msg: "retrieve DNS nodes error", Extra: err}
	}

	err = h.pingNodes()
	if err != nil {
		utils.LogError("ping nodes error", err.Error())
	}

	h.resolveLongestBlockchain()

	if h.Repos.WalletRepo.IsEmpty() {
		if err := h.createNewWallet(); err != nil {
			return utils.GenericError{Msg: "failed to create new wallet", Extra: err}
		}
	}

	return nil
}

func (h EventHandler) introduceToDNS() error {
	selfNode, err := h.getSelfNode()
	if err != nil {
		return err
	}

	return dns_client.AddDNSNode(h.getDNSHost(), selfNode)
}

func (h EventHandler) refreshDNSNodes() error {
	nodes, err := dns_client.GetDNSNodes(h.getDNSHost())
	if err != nil {
		return utils.GenericError{Msg: "couldn't retrieve nodes from DNS", Extra: err}
	}

	nodes = utils.Filter(nodes, func(n nd.Node) bool {
		return n.Name != h.Config.Get("NODE_NAME")
	})

	for _, node := range nodes {
		if err := h.Repos.NodeRepo.AddNode(node); err != nil {
			return utils.GenericError{Msg: "couldn't save nodes received from DNS", Extra: err}
		}
	}

	return nil
}

func (h EventHandler) pingNodes() error {
	nodes, err := h.Repos.NodeRepo.GetNodes()
	if err != nil {
		return utils.GenericError{Msg: "couldn't load nodes", Extra: err}
	}

	selfNode, err := h.getSelfNode()
	if err != nil {
		return err
	}

	responses := node_client.PingNodes(nodes, selfNode)

	if responses.HasConnectionRefused() {
		h.Bus.Handle(eventbus.DataEvent{Ev: events.ConnectionRefusedEvent})
	}

	if responses.ErrorsRatio() < 1 {
		return utils.GenericError{
			Msg: fmt.Sprintf("failed to share the transaction with other nodes: %v", responses.Errors()),
		}
	}

	return nil
}

func (h EventHandler) getBlockchains(nodes []nd.Node) []*bc.Blockchain {

	responses := node_client.GetBlockchains(nodes)

	if responses.HasConnectionRefused() {
		h.Bus.Handle(eventbus.DataEvent{Ev: events.ConnectionRefusedEvent})
	}

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

func (h EventHandler) resolveLongestBlockchain() error {
	nodes, err := h.Repos.NodeRepo.GetNodes()
	if err != nil {
		return err
	}

	// TODO: include the below in a single lock

	blockchains := h.getBlockchains(nodes)
	utils.LogInfo("Retrieved blockchains", len(blockchains))

	localBlockchain, _ := h.Repos.BlockchainRepo.GetBlockchain()
	blockchains = append(blockchains, localBlockchain)

	maxLengthBlockchain := bc.GetMaxLengthBlockchain(blockchains)

	if len(maxLengthBlockchain.Blocks) == 0 {
		return nil
	}

	err = h.Repos.BlockchainRepo.UpdateBlockchain(maxLengthBlockchain)
	if err != nil {
		return utils.GenericError{Msg: "failed to update local blockchain", Extra: err}
	}

	return nil
}

func (h EventHandler) createNewWallet() error {
	wallet, err := wallet_client.GetNewWallet(h.getWalletsHost())
	if err != nil {
		return err
	}

	return h.Repos.WalletRepo.AddWallet(wallet)
}

func (h EventHandler) HandleTransactionReceivedEvent(event eventbus.DataEvent) {
	tx := event.Data.(bc.Transaction)

	nodes, _ := h.Repos.NodeRepo.GetNodes()
	responses := node_client.ShareTx(nodes, tx)
	if responses.ErrorsRatio() > 0 {
		utils.LogError("Failed to share the transaction with some nodes", responses.Errors())
	}
}

func (h EventHandler) HandleBlockMinedEvent(event eventbus.DataEvent) {

	rewardTx, err := h.makeRewardTx()
	if err != nil {
		utils.LogError("Failed to create reward transaction", err.Error())
		return
	}

	err = h.reward(rewardTx)
	if err != nil {
		utils.LogError("Failed to reward self", err.Error())
	}

	utils.LogInfo("Rewarded self with", h.Config.DefaultRewardAmount)
}

func (h EventHandler) HandleBlockMiningFailedEvent(event eventbus.DataEvent) {
	utils.LogInfo("Resolving longest blockchain")
	err := h.resolveLongestBlockchain()
	if err != nil {
		utils.LogError("Failed to resolve longest blockchain", err.Error())
	}
}

func (h EventHandler) makeRewardTx() (bc.Transaction, error) {
	if h.Repos.WalletRepo.IsEmpty() {
		return bc.Transaction{}, utils.GenericError{Msg: "node has no wallet"}
	}

	wallets, err := h.Repos.WalletRepo.GetWallets()
	if err != nil {
		return bc.Transaction{}, utils.GenericError{Msg: "failed to get wallets", Extra: err}
	}

	return bc.Transaction{
		Body: bc.TransactionBody{
			Sender:    h.Config.CoinBaseSenderAddress,
			Recipient: wallets[0].AddressString(),
			Amount:    h.Config.DefaultRewardAmount,
		},
	}, nil
}

func (h EventHandler) reward(tx bc.Transaction) error {

	tx, err := h.Repos.BlockchainRepo.AddTx(tx)
	if err != nil {
		return utils.GenericError{Msg: "failed to add reward transaction", Extra: err}
	}

	nodes, _ := h.Repos.NodeRepo.GetNodes()
	if err != nil {
		return utils.GenericError{Msg: "failed to load nodes", Extra: err}
	}

	responses := node_client.ShareTx(nodes, tx)
	if responses.ErrorsRatio() > 0 {
		return utils.GenericError{
			Msg: fmt.Sprintf("failed to share the transaction with other nodes: %v", responses.Errors()),
		}
	}

	return nil
}

func (h EventHandler) HandleConnectionRefusedEvent(event eventbus.DataEvent) {
	utils.LogInfo("Refresing DNS nodes")
	err := h.refreshDNSNodes()
	if err != nil {
		utils.LogError("Failed to refresh DNS nodes", err.Error())
	}
}

func NewEventBus(config *cfg.Config, repos *rep.Repos) *eventbus.Bus {
	bus := eventbus.NewBus()

	eh := EventHandler{Bus: bus, Config: config, Repos: repos}

	bus.RegisterEventHandler(events.InitNodeEvent, eh.HandleInitNode)
	bus.RegisterEventHandler(events.TransactionReceivedEvent, eh.HandleTransactionReceivedEvent)
	bus.RegisterEventHandler(events.BlockMinedEvent, eh.HandleBlockMinedEvent)
	bus.RegisterEventHandler(events.BlockMiningFailedEvent, eh.HandleBlockMiningFailedEvent)
	bus.RegisterEventHandler(events.ConnectionRefusedEvent, eh.HandleConnectionRefusedEvent)

	return bus
}
