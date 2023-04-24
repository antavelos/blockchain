package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"
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
)

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
}

var _nodeDB *db.NodeDB
var _blockchainDB *db.BlockchainDB
var _walletsDB *db.WalletDB

func getNodeDb() *db.NodeDB {
	if _nodeDB == nil {
		_nodeDB = db.GetNodeDb(config["NODES_FILENAME"])
	}
	return _nodeDB
}

func getBlockchainDb() *db.BlockchainDB {
	if _blockchainDB == nil {
		_blockchainDB = db.GetBlockchainDb(config["BLOCKCHAIN_FILENAME"])
	}
	return _blockchainDB
}

func getWalletDb() *db.WalletDB {
	if _walletsDB == nil {
		_walletsDB = db.GetWalletDb(config["WALLETS_FILENAME"])
	}
	return _walletsDB
}

func atoiConfigValue(key string, defaultVal int) int {
	value, err := strconv.Atoi(config[key])
	if err != nil {
		msg := fmt.Sprintf("Couldn't parse '%v' config value. Using default value: %v", key, defaultTxsPerBlock)
		common.LogInfo(msg)
		return defaultVal
	}

	return value
}

func atofConfigValue(key string, defaultVal float64) float64 {
	value, err := strconv.ParseFloat(config[key], 1)
	if err != nil {
		msg := fmt.Sprintf("Couldn't parse '%v' config value. Using default value: %v", key, defaultTxsPerBlock)
		common.LogInfo(msg)
		return defaultVal
	}

	return value
}

func getMiningDifficulty() int {
	return atoiConfigValue("MINING_DIFFICULTY", defaultMiningDifficulty)
}

func getTxsNumPerBlock() int {
	return atoiConfigValue("TXS_PER_BLOCK", defaultTxsPerBlock)
}

func getRewardAmount() float64 {
	return atofConfigValue("REWARD_AMOUNT", defaultRewardAmount)
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
		ioNewBlockchain()
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

	return nd.NewNode("", ip, config["PORT"]), nil
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

	// TODO: check if this step can be included in other calls
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

func retrieveDnsNodes() error {
	nodes, err := dns_client.GetDnsNodes(getDnsHost())
	if err != nil {
		return common.GenericError{Msg: "couldn't retrieve nodes from DNS", Extra: err}
	}

	nodes = common.Filter(nodes, func(n nd.Node) bool {
		return n.Port != config["PORT"]
	})

	ndb := getNodeDb()
	if err := ndb.SaveNodes(nodes); err != nil {
		return common.GenericError{Msg: "couldn't save nodes received from DNS", Extra: err}
	}

	return nil
}

func pingNodes() error {
	ndb := getNodeDb()

	nodes, err := ndb.LoadNodes()
	if err != nil {
		return common.GenericError{Msg: "couldn't load nodes", Extra: err}
	}

	selfNode, err := getSelfNode()
	if err != nil {
		return err
	}

	responses := node_client.PingNodes(nodes, selfNode)

	if responses.ErrorsRatio() < 1 {
		return common.GenericError{
			Msg: fmt.Sprintf("failed to share the transaction with other nodes\n %v", strings.Join(responses.ErrorStrings(), "\n")),
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
	wallet, err := ioGetWallet()
	if err != nil {
		return err
	}

	rewardTx := bc.Transaction{
		Body: bc.TransactionBody{
			Sender:    "0",
			Recipient: hex.EncodeToString(wallet.Address),
			Amount:    rewardAmount,
		},
	}

	bus.Publish(RewardTransaction, rewardTx)

	return nil
}

func resolveLongestBlockchain() error {
	ndb := getNodeDb()

	nodes, err := ndb.LoadNodes()
	if err != nil {
		return err
	}

	responses := node_client.GetBlockchains(nodes)
	blockchains := common.Map(responses, func(response rest.Response) *bc.Blockchain {
		if response.Err != nil {
			return &bc.Blockchain{}
		}

		blockchain, err := bc.UnmarshalBlockchain(response.Body)
		if err != nil {
			return &bc.Blockchain{}
		}

		return &blockchain
	})

	bdb := getBlockchainDb()
	localBlockchain, _ := bdb.LoadBlockchain()
	blockchains = append(blockchains, localBlockchain)

	maxLengthBlockchain := bc.GetMaxLengthBlockchain(blockchains)

	if len(maxLengthBlockchain.Blocks) == 0 {
		return nil
	}

	err = bdb.UpdateBlockchain(maxLengthBlockchain)
	if err != nil {
		return common.GenericError{Msg: "failed to update local blockchain", Extra: err}
	}

	return nil
}

func hasWallet() bool {
	wdb := getWalletDb()

	return !wdb.IsEmpty()
}

func createNewWallet() error {
	wallet, err := wallet_client.GetNewWallet(getWalletsHost())
	if err != nil {
		return err
	}

	wdb := getWalletDb()
	return wdb.SaveWallet(wallet)
}
