package main

import (
	"encoding/hex"
	"flag"
	"fmt"
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

func getSelfPort() string {
	return os.Getenv("PORT")
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

	return nd.Node{
		IP:   ip,
		Port: getSelfPort(),
	}, nil
}

func main() {
	mine := flag.Bool("mine", false, "Indicates whether it will run as miner")
	init := flag.Bool("init", false, "Initialises the blockchain. Existing blockchain will be overriden. Overrules other options.")

	flag.Parse()

	if *init {
		ioNewBlockchain()
	}

	err := initNode()
	if err != nil {
		common.ErrorLogger.Fatalf(err.Error())
	}

	if *mine {
		go runMiningLoop()
	}

	router := InitRouter()
	router.Run(fmt.Sprintf(":%v", getSelfPort()))
}

type NodeError common.GenericError

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
		common.ErrorLogger.Printf("ping nodes error")
	}

	resolveLongestBlockchain()

	if !hasWallet() {
		err = createNewWallet()
		return common.GenericError{Msg: "failed to create new wallet", Extra: err}
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
		return n.Port != getSelfPort()
	})

	ndb := db.GetNodeDb()
	if err := ndb.SaveNodes(nodes); err != nil {
		return common.GenericError{Msg: "couldn't save nodes received from DNS"}
	}

	return nil
}

func pingNodes() error {
	ndb := db.GetNodeDb()

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
	i := 0
	for {
		block, err := Mine()
		if err != nil {
			common.ErrorLogger.Printf("New block [FAIL]")

			common.InfoLogger.Println("Resolving longest blockchain")
			err := resolveLongestBlockchain()
			if err != nil {
				common.ErrorLogger.Printf("Failed to resolve longest blockchain")
			}

		} else {
			// TODO: publish event
			common.InfoLogger.Printf("New block [OK]: %v", block.Idx)

			err := reward()
			if err != nil {
				common.ErrorLogger.Printf("Failed to reward node")
			}

		}

		time.Sleep(5 * time.Second)
		i = i + 1
	}
}

func reward() error {
	wallet, err := ioGetWallet()
	if err != nil {
		return common.GenericError{Msg: "node wallet not available", Extra: err}
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
		return common.GenericError{Msg: "failed to add reward transaction", Extra: err}
	}

	ndb := db.GetNodeDb()
	nodes, err := ndb.LoadNodes()
	if err != nil {
		return common.GenericError{Msg: "failed to load nodes", Extra: err}
	}

	responses := node_client.ShareTx(nodes, tx)
	if responses.ErrorsRatio() > 0 {
		return common.GenericError{
			Msg: fmt.Sprintf("failed to share the transaction with other nodes\n %v", strings.Join(responses.ErrorStrings(), "\n")),
		}
	}

	return nil
}

func resolveLongestBlockchain() error {
	ndb := db.GetNodeDb()

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

	bdb := db.GetBlockchainDb()
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
	wdb := db.GetWalletDb()

	wallets, _ := wdb.LoadWallets()

	return len(wallets) > 0
}

func createNewWallet() error {
	wallet, err := wallet_client.GetNewWallet(getWalletsHost())
	if err != nil {
		return err
	}

	wdb := db.GetWalletDb()
	return wdb.SaveWallet(wallet)
}
