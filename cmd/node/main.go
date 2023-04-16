package main

import (
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	dns_client "github.com/antavelos/blockchain/pkg/clients/dns"
	node_client "github.com/antavelos/blockchain/pkg/clients/node"
	wallet_client "github.com/antavelos/blockchain/pkg/clients/wallet"
	"github.com/antavelos/blockchain/pkg/common"
	"github.com/antavelos/blockchain/pkg/db"
	"github.com/antavelos/blockchain/pkg/lib/rest"
	bc "github.com/antavelos/blockchain/pkg/models/blockchain"
	nd "github.com/antavelos/blockchain/pkg/models/node"
)

func getDnsHost() string {
	return fmt.Sprintf("http://%v:%v", os.Getenv("DNS_HOST"), os.Getenv("DNS_PORT"))
}

func getWalletsHost() string {
	return fmt.Sprintf("http://%v:%v", os.Getenv("WALLETS_HOST"), os.Getenv("WALLETS_PORT"))
}

func getSelfHost() string {
	return fmt.Sprintf("http://%v:%v", getSelfIP(), getSelfPort())
}

func getSelfPort() string {
	return os.Getenv("PORT")
}

func getSelfIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		common.ErrorLogger.Printf("Cannot get IP: " + err.Error() + "\n")
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return ""
}

func getSelfNode() nd.Node {
	return nd.Node{
		IP:   getSelfIP(),
		Port: getSelfPort(),
	}
}

func main() {
	mine := flag.Bool("mine", false, "Indicates whether it will run as miner")
	init := flag.Bool("init", false, "Initialises the blockchain. Existing blockchain will be overriden. Overrules other options.")

	flag.Parse()

	if *init {
		ioNewBlockchain()
	}

	initNode()

	if *mine {
		go runMiningLoop()
	}

	router := InitRouter()
	router.Run(fmt.Sprintf(":%v", getSelfPort()))
}

func initNode() {
	introduceToDns()

	retrieveDnsNodes()

	pingNodes()

	resolveLongestBlockchain()

	getNewWallet()
}

func introduceToDns() error {
	return dns_client.AddDnsNode(getDnsHost(), getSelfNode())
}

func retrieveDnsNodes() error {
	ndb := db.GetNodeDb()

	nodes, err := dns_client.GetDnsNodes(getDnsHost())
	if err != nil {
		return fmt.Errorf("Couldn't retrieve nodes from DNS %v", err.Error())
	}

	nodes = common.Filter(nodes, func(n nd.Node) bool {
		return n.Port != getSelfPort()
	})

	if err := ndb.SaveNodes(nodes); err != nil {
		return errors.New("Couldn't save nodes received from DNS.")
	}

	return nil
}

func pingNodes() error {
	ndb := db.GetNodeDb()

	nodes, err := ndb.LoadNodes()
	if err != nil {
		return fmt.Errorf("Couldn't load nodes: %v", err.Error())
	}

	responses := node_client.PingNodes(nodes, getSelfNode())

	if responses.ErrorsRatio() < 1 {
		return fmt.Errorf("Failed to ping all nodes: %v", strings.Join(responses.ErrorStrings(), "/n"))
	}

	return nil
}

func runMiningLoop() {
	i := 0
	for {
		block, err := Mine()
		if err != nil {
			common.ErrorLogger.Printf("New block [FAIL]: %v", err.Error())

			common.InfoLogger.Println("Resolving longest blockchain")
			err := resolveLongestBlockchain()
			if err != nil {
				common.ErrorLogger.Printf("Failed to resolve longest blockchain: %v", err.Error())
			}

		} else {
			// TODO: publish event
			common.InfoLogger.Printf("New block [OK]: %v", block.Idx)

			err := reward()
			if err != nil {
				common.ErrorLogger.Printf("Failed to reward node: %v", err.Error())
			}

		}

		time.Sleep(5 * time.Second)
		i = i + 1
	}
}

func reward() error {
	wallet, err := ioGetWallet()
	if err != nil {
		return fmt.Errorf("node wallet not available: %v", err)
	}

	tx := bc.Transaction{
		Body: bc.TransactionBody{
			Sender:    "0",
			Recipient: hex.EncodeToString(wallet.Address),
			Amount:    1.0,
		},
	}

	tx, err = ioAddTx(tx)
	if err != nil {
		return fmt.Errorf("failed to add reward transaction: %v", err.Error())
	}

	ndb := db.GetNodeDb()
	nodes, err := ndb.LoadNodes()
	if err != nil {
		return fmt.Errorf("failed to load nodes: %v", err.Error())
	}

	responses := node_client.ShareTx(nodes, tx)
	if responses.ErrorsRatio() > 0 {
		return fmt.Errorf("failed to share the transaction with other nodes: \n%v", strings.Join(responses.ErrorStrings(), "\n"))
	}

	return nil
}

func resolveLongestBlockchain() error {
	ndb := db.GetNodeDb()
	bdb := db.GetBlockchainDb()

	nodes, err := ndb.LoadNodes()
	if err != nil {
		return err
	}

	responses := node_client.GetBlockchains(nodes)
	blockchains := common.Map(responses, func(response rest.Response) *bc.Blockchain {
		if response.Err != nil {
			return &bc.Blockchain{}
		}

		return response.Body.(*bc.Blockchain)
	})

	localBlockchain, _ := bdb.GetBlockchainDb()
	blockchains := append(blockchains, localBlockchain)

	maxLengthBlockchain := bc.GetMaxLengthBlockchain(blockchains)

	if len(maxLengthBlockchain.Blocks) == 0 {
		return nil
	}

	err = bdb.UpdateBlockchain(maxLengthBlockchain)
	if err != nil {
		return fmt.Errorf("failed to update local blockchain: %v", err.Error())
	}

	return nil
}

func getMaxLengthBlockchain(blockchains []*bc.Blockchain) *bc.Blockchain {
	bdb := db.GetBlockchainDb()

	maxLengthBlockchain, _ := bdb.LoadBlockchain()

	for _, blockchain := range blockchains {
		if maxLengthBlockchain == nil || len(blockchain.Blocks) > len(maxLengthBlockchain.Blocks) {
			maxLengthBlockchain = blockchain
		}
	}

	return maxLengthBlockchain
}

func getNewWallet() error {
	wallet, err := wallet_client.GetNewWallet(getWalletsHost())
	if err != nil {
		return err
	}

	wdb := db.GetWalletDb()

	return wdb.SaveWallet(wallet)
}
