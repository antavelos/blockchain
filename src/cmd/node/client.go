package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	bc "github.com/antavelos/blockchain/src/blockchain"
)

func getBlockchain(node bc.Node) (*bc.Blockchain, error) {
	url := node.Host + "/blockchain"

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var blockchain bc.Blockchain
	if err := json.Unmarshal(body, &blockchain); err != nil {
		return nil, err
	}

	return &blockchain, nil
}

func GetDnsNodes() ([]bc.Node, error) {
	var nodes []bc.Node
	url := getDnsHost() + "/nodes"

	resp, err := http.Get(url)
	if err != nil {
		return nodes, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nodes, err
	}

	if err := json.Unmarshal(body, &nodes); err != nil {
		return nodes, err
	}

	return nodes, nil
}

func IntroduceToDns() error {
	url := getDnsHost() + "/nodes"
	selfNode := bc.Node{Host: fmt.Sprintf("http://%v:%v", getSelfHost(), getSelfPort())}

	selfNodesBytes, err := json.Marshal(selfNode)
	if err != nil {
		return err
	}

	body, err := postData(url, selfNodesBytes)
	if err != nil {
		return fmt.Errorf("failed to reach DNS: %v", string(body))
	}

	var nodes []bc.Node
	if err := json.Unmarshal(body, &nodes); err != nil {
		return err
	}

	return nil
}

func updateBlockchain(oldBlockchain *bc.Blockchain, newBlockchain *bc.Blockchain) *bc.Blockchain {
	if oldBlockchain == nil {
		return newBlockchain
	}

	// TODO: append the blocks diff
	oldBlockchain.Blocks = newBlockchain.Blocks

	// TODO: to refactor
	for i := len(oldBlockchain.Blocks) - 1; i > 0; i-- {
		for _, tx := range oldBlockchain.TxPool {
			if oldBlockchain.Blocks[i].HasTx(tx) {
				oldBlockchain.RemoveTx(tx)
			}
		}
	}

	return oldBlockchain
}

func ResolveLongestBlockchain() error {
	ndb := getNodeDb()
	bdb := getBlockchainDb()

	nodes, err := ndb.LoadNodes()
	if err != nil {
		return err
	}

	maxLengthBlockchain := getMaxLengthBlockchain(nodes)

	if len(maxLengthBlockchain.Blocks) == 0 {
		return nil
	}

	m := sync.Mutex{}

	m.Lock()
	defer m.Unlock()

	blockchain, _ := bdb.LoadBlockchain()

	blockchain = updateBlockchain(blockchain, maxLengthBlockchain)

	bdb.SaveBlockchain(*blockchain)

	return nil
}

func getMaxLengthBlockchain(nodes []bc.Node) *bc.Blockchain {
	bdb := getBlockchainDb()

	maxLengthBlockchain, _ := bdb.LoadBlockchain()

	for _, node := range nodes {
		nodeBlockchain, err := getBlockchain(node)
		if err != nil {
			ErrorLogger.Printf("Couldn't retrieve blockchain from node %v: %v", node.Host, err.Error())
			continue
		}

		if maxLengthBlockchain == nil || len(nodeBlockchain.Blocks) > len(maxLengthBlockchain.Blocks) {
			maxLengthBlockchain = nodeBlockchain
		}
	}

	return maxLengthBlockchain
}

type NodeBroadcaster interface {
	Broadcast() error
}

type PingNodesBroadcaster struct {
	Node bc.Node
	Url  string
}

func NewPingNodesBroadcaster(node bc.Node) PingNodesBroadcaster {
	return PingNodesBroadcaster{Node: node, Url: node.Host + pingURL}
}

func postData(url string, data []byte) ([]byte, error) {
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}

	return io.ReadAll(resp.Body)
}

func (pnb PingNodesBroadcaster) Broadcast() error {
	ndb := getNodeDb()
	selfNode := bc.Node{Host: fmt.Sprintf("http://%v:%v", getSelfHost(), getSelfPort())}

	jsonSelfNode, err := json.Marshal(selfNode)
	if err != nil {
		return err
	}

	body, err := postData(pnb.Url, jsonSelfNode)
	if err != nil {
		return fmt.Errorf("node %v: %v", pnb.Node.Host, string(body))
	}

	var nodes []bc.Node
	if err := json.Unmarshal(body, &nodes); err != nil {
		return err
	}

	for _, node := range nodes {
		if node.Host != selfNode.Host {
			ndb.AddNode(node)
		}
	}

	return nil
}

type TxNodeBroadcaster struct {
	Node bc.Node
	Tx   bc.Transaction
	Url  string
}

func NewTxNodeBroadcaster(node bc.Node, tx bc.Transaction) TxNodeBroadcaster {
	return TxNodeBroadcaster{Node: node, Tx: tx, Url: node.Host + sharedTransactionsURL}
}

func (txnb TxNodeBroadcaster) Broadcast() error {
	txBytes, err := json.Marshal(txnb.Tx)
	if err != nil {
		return err
	}

	_, err = postData(txnb.Url, txBytes)
	if err != nil {
		return fmt.Errorf("node %v: %v", txnb.Node.Host, err.Error())
	}

	return nil
}

type BlockNodeBroadcaster struct {
	Node  bc.Node
	Block bc.Block
	Url   string
}

func NewBlockNodeBroadcaster(node bc.Node, block bc.Block) BlockNodeBroadcaster {
	return BlockNodeBroadcaster{Node: node, Block: block, Url: node.Host + sharedBlocksURL}
}

func (bnb BlockNodeBroadcaster) Broadcast() error {
	blockBytes, err := json.Marshal(bnb.Block)
	if err != nil {
		return err
	}

	_, err = postData(bnb.Url, blockBytes)
	if err != nil {
		return fmt.Errorf("node %v: %v", bnb.Node.Host, err.Error())
	}

	return nil
}

func BroadcastNodes(nbs []NodeBroadcaster) []error {
	errorsChan := make(chan error)
	var wg sync.WaitGroup

	for _, nb := range nbs {
		wg.Add(1)
		go func(na NodeBroadcaster) {
			defer wg.Done()
			errorsChan <- na.Broadcast()
		}(nb)
	}

	go func() {
		wg.Wait()
		close(errorsChan)
	}()

	resultErrors := make([]error, 0)
	for ch := range errorsChan {
		resultErrors = append(resultErrors, ch)
	}

	return resultErrors
}

func ShareTx(tx bc.Transaction) []error {
	ndb := getNodeDb()

	nodes, err := ndb.LoadNodes()
	if err != nil {
		return []error{err}
	}

	var txnb []NodeBroadcaster
	for _, node := range nodes {
		txnb = append(txnb, NewTxNodeBroadcaster(node, tx))
	}

	return BroadcastNodes(txnb)
}

func ShareBlock(block bc.Block) []error {
	ndb := getNodeDb()

	nodes, err := ndb.LoadNodes()
	if err != nil {
		return []error{err}
	}

	var txnb []NodeBroadcaster
	for _, node := range nodes {
		txnb = append(txnb, NewBlockNodeBroadcaster(node, block))
	}

	return BroadcastNodes(txnb)
}

func Ping() []error {
	ndb := getNodeDb()

	nodes, err := ndb.LoadNodes()
	if err != nil {
		return []error{err}
	}

	var txnb []NodeBroadcaster
	for _, node := range nodes {
		txnb = append(txnb, NewPingNodesBroadcaster(node))
	}

	return BroadcastNodes(txnb)
}
