package main

import (
	"flag"
	"fmt"
	"net"
	"time"

	dns_client "github.com/antavelos/blockchain/pkg/clients/dns"
	node_client "github.com/antavelos/blockchain/pkg/clients/node"
	wallet_client "github.com/antavelos/blockchain/pkg/clients/wallet"

	"github.com/antavelos/blockchain/pkg/common"
	"github.com/antavelos/blockchain/pkg/db"
	"github.com/antavelos/blockchain/pkg/lib/bus"
	cfg "github.com/antavelos/blockchain/pkg/lib/config"
	"github.com/antavelos/blockchain/pkg/lib/rest"
	bc "github.com/antavelos/blockchain/pkg/models/blockchain"
	nd "github.com/antavelos/blockchain/pkg/models/node"
	bc_repo "github.com/antavelos/blockchain/pkg/repo/blockchain"
	node_repo "github.com/antavelos/blockchain/pkg/repo/node"
	wallet_repo "github.com/antavelos/blockchain/pkg/repo/wallet"
)

const coinBaseSenderAddress = "0"
const defaultTxsPerBlock = 10
const defaultMiningDifficulty = 2
const defaultRewardAmount = 1.0

var config cfg.Config

var envVars []string = []string{
	"PORT",
	"DNS_HOST",
	"DNS_PORT",
	"WALLETS_HOST",
	"WALLETS_PORT",
	"NODES_FILENAME",
	"BLOCKCHAIN_FILENAME",
	"WALLETS_FILENAME",
	"MINING_DIFFICULTY",
	"TXS_PER_BLOCK",
	"REWARD_AMOUNT",
	"NODE_NAME",
}

var _nodeDB *db.DB
var _blockchainDB *db.DB
var _walletsDB *db.DB

func getWalletDb() *db.DB {
	if _walletsDB == nil {
		_walletsDB = db.NewDB(config["WALLETS_FILENAME"])
	}
	return _walletsDB
}

func getNodeDb() *db.DB {
	if _nodeDB == nil {
		_nodeDB = db.NewDB(config["NODES_FILENAME"])
	}
	return _nodeDB
}

func getBlockchainDb() *db.DB {
	if _blockchainDB == nil {
		_blockchainDB = db.NewDB(config["BLOCKCHAIN_FILENAME"])
	}
	return _blockchainDB
}

func getBlockchainRepo() *bc_repo.BlockchainRepo {
	return bc_repo.NewBlockchainRepo(getBlockchainDb())
}

func getNodeRepo() *node_repo.NodeRepo {
	return node_repo.NewNodeRepo(getNodeDb())
}

func getWalletRepo() *wallet_repo.WalletRepo {
	return wallet_repo.NewWalletRepo(getWalletDb())
}

func getMiningDifficulty() int {
	return config.GetValueAtoi("MINING_DIFFICULTY", defaultMiningDifficulty)
}

func getTxsNumPerBlock() int {
	return config.GetValueAtoi("TXS_PER_BLOCK", defaultTxsPerBlock)
}

func getRewardAmount() float64 {
	return config.GetValueAtof("REWARD_AMOUNT", defaultRewardAmount)
}

func main() {
	mine := flag.Bool("mine", false, "Indicates whether it will run as miner")
	init := flag.Bool("init", false, "Initialises the blockchain. Existing blockchain will be overriden. Overrules other options.")

	flag.Parse()

	var err error

	config, err = cfg.LoadConfig(envVars)
	if err != nil {
		common.LogFatal("Configuration error", err.Error())
	}

	if *init {
		brepo := getBlockchainRepo()
		if _, err := brepo.CreateBlockchain(); err != nil {
			common.LogFatal("failed to create blockchain", err.Error())
		}
	}

	err = initNode()
	if err != nil {
		common.LogFatal(err.Error())
	}

	if *mine {
		go runMiningLoop()
	}

	go startEventLoop()

	// TODO: add a periodic longest blockchain resolve

	router := InitRouter()
	router.Run(fmt.Sprintf(":%v", config["PORT"]))
}

func getUrl(host string, port string) string {
	return fmt.Sprintf("http://%v:%v", host, port)
}
func getDnsHost() string {
	return getUrl(config["DNS_HOST"], config["DNS_PORT"])
}

func getWalletsHost() string {
	return getUrl(config["WALLETS_HOST"], config["WALLETS_PORT"])
}

func getSelfIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", common.GenericError{Msg: "IP not found", Extra: err}
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", common.GenericError{Msg: "IP not found"}
}

func getSelfNode() (nd.Node, error) {
	ip, err := getSelfIP()
	if err != nil {
		return nd.Node{}, err
	}

	return nd.NewNode(config["NODE_NAME"], ip, config["PORT"]), nil
}

func initNode() error {
	err := introduceToDns()
	if err != nil {
		return common.GenericError{Msg: "failed to introduce self at DNS", Extra: err}
	}

	err = retrieveDnsNodes()
	if err != nil {
		return common.GenericError{Msg: "retrieve DNS nodes error", Extra: err}
	}

	err = pingNodes()
	if err != nil {
		common.LogError("ping nodes error", err.Error())
	}

	resolveLongestBlockchain()

	if !hasWallet() {
		if err := createNewWallet(); err != nil {
			return common.GenericError{Msg: "failed to create new wallet", Extra: err}
		}
	}

	return nil
}

func introduceToDns() error {
	selfNode, err := getSelfNode()
	if err != nil {
		return err
	}

	return dns_client.AddDnsNode(getDnsHost(), selfNode)
}

// TODO: move to dedicated module
func retrieveDnsNodes() error {
	nodes, err := dns_client.GetDnsNodes(getDnsHost())
	if err != nil {
		return common.GenericError{Msg: "couldn't retrieve nodes from DNS", Extra: err}
	}

	nodes = common.Filter(nodes, func(n nd.Node) bool {
		return n.Name != config["NODE_NAME"]
	})

	nrepo := getNodeRepo()
	for _, node := range nodes {
		if err := nrepo.AddNode(node); err != nil {
			return common.GenericError{Msg: "couldn't save nodes received from DNS", Extra: err}
		}
	}

	return nil
}

func pingNodes() error {
	nrepo := getNodeRepo()

	nodes, err := nrepo.GetNodes()
	if err != nil {
		return common.GenericError{Msg: "couldn't load nodes", Extra: err}
	}

	selfNode, err := getSelfNode()
	if err != nil {
		return err
	}

	responses := node_client.PingNodes(nodes, selfNode)

	if responses.HasConnectionRefused() {
		bus.Publish(RefreshDnsNodesTopic, nil)
	}

	if responses.ErrorsRatio() < 1 {
		return common.GenericError{
			Msg: fmt.Sprintf("failed to share the transaction with other nodes: %v", responses.Errors()),
		}
	}

	return nil
}

func runMiningLoop() {
	for {
		block, err := Mine()
		if err != nil {
			common.LogError("New block [FAIL]", err.Error())

			common.LogInfo("Resolving longest blockchain")
			err := resolveLongestBlockchain()
			if err != nil {
				common.LogError("Failed to resolve longest blockchain", err.Error())
			}

		} else {
			common.LogInfo("New block [OK]", block.Idx)

			// TODO: check who should do the reward
			//
			rewardAmount := getRewardAmount()
			err := rewardSelf(rewardAmount)
			if err != nil {
				common.LogError("failed to create reward transaction", err.Error())
			} else {
				common.LogInfo("rewarded self with", rewardAmount)
			}
		}

		time.Sleep(5 * time.Second)
	}
}

func rewardSelf(rewardAmount float64) error {
	walletRepo := getWalletRepo()
	wallets, err := walletRepo.GetWallets()
	if err != nil {
		return err
	}

	wallet := wallets[0]
	rewardTx := bc.Transaction{
		Body: bc.TransactionBody{
			Sender:    coinBaseSenderAddress,
			Recipient: wallet.AddressString(),
			Amount:    rewardAmount,
		},
	}

	bus.Publish(RewardTransactionTopic, rewardTx)

	return nil
}

func getBlockchains(nodes []nd.Node) []*bc.Blockchain {

	responses := node_client.GetBlockchains(nodes)

	if responses.HasConnectionRefused() {
		bus.Publish(RefreshDnsNodesTopic, nil)
	}

	noErrorResponses := common.Filter(responses, func(response rest.Response) bool {
		return response.Err == nil
	})

	blockchains := common.Map(noErrorResponses, func(response rest.Response) *bc.Blockchain {
		blockchain, err := bc.UnmarshalBlockchain(response.Body)
		if err != nil {
			return &bc.Blockchain{}
		}

		return &blockchain
	})

	return blockchains
}

func resolveLongestBlockchain() error {
	nrepo := getNodeRepo()

	nodes, err := nrepo.GetNodes()
	if err != nil {
		return err
	}

	// TODO: include the below in a single lock

	blockchains := getBlockchains(nodes)
	common.LogInfo("Retrieved blockchains", len(blockchains))

	brepo := getBlockchainRepo()
	localBlockchain, _ := brepo.GetBlockchain()
	blockchains = append(blockchains, localBlockchain)

	maxLengthBlockchain := bc.GetMaxLengthBlockchain(blockchains)

	if len(maxLengthBlockchain.Blocks) == 0 {
		return nil
	}

	err = brepo.UpdateBlockchain(maxLengthBlockchain)
	if err != nil {
		return common.GenericError{Msg: "failed to update local blockchain", Extra: err}
	}

	return nil
}

func hasWallet() bool {
	wrepo := getWalletRepo()

	return !wrepo.IsEmpty()
}

func createNewWallet() error {
	wallet, err := wallet_client.GetNewWallet(getWalletsHost())
	if err != nil {
		return err
	}

	wrepo := getWalletRepo()
	return wrepo.AddWallet(wallet)
}
