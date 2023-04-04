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

func ping(node bc.Node) error {
	host := node.Host + "/ping"
	selfNode := bc.Node{Host: "http://localhost:" + *Port}
	jsonSelfNode, err := json.Marshal(selfNode)
	if err != nil {
		return err
	}

	resp, err := http.Post(host, "application/json", bytes.NewBuffer(jsonSelfNode))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var nodes []bc.Node
	if err := json.Unmarshal(body, &nodes); err != nil {
		return err
	}

	for _, node := range nodes {
		if node.Host != selfNode.Host {
			ioAddNode(node)
		}
	}

	return nil
}

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

func pingDns() ([]bc.Node, error) {
	var nodes []bc.Node
	url := dnsHost + "/nodes"

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

func resolveLongestBlockchain(nodes []bc.Node) {
	maxLengthBlockchain := getMaxLengthBlockchain(nodes)

	if len(maxLengthBlockchain.Blocks) == 0 {
		return
	}

	blockchain, _ := ioLoadBlockchain()

	if blockchain == nil {
		blockchain = maxLengthBlockchain
	} else {
		blockchain.Blocks = maxLengthBlockchain.Blocks
	}

	ioSaveBlockchain(*blockchain)
}

func getMaxLengthBlockchain(nodes []bc.Node) *bc.Blockchain {
	maxLengthBlockchain, _ := ioLoadBlockchain()

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

func shareTx(node bc.Node, tx bc.Transaction) error {
	txBytes, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	url := node.Host + "/shared-transactions"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(txBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 201 {
		return fmt.Errorf("node %v: %v", node.Host, string(body))
	}

	return nil
}

func ShareTx(tx bc.Transaction) []error {
	nodes, err := ioLoadNodes()
	if err != nil {
		return []error{err}
	}

	errorsChan := make(chan error)
	var wg sync.WaitGroup

	for _, node := range nodes {
		if node.GetPort() == *Port {
			continue
		}

		wg.Add(1)
		go func(node bc.Node, tx bc.Transaction) {
			defer wg.Done()
			errorsChan <- shareTx(node, tx)
		}(node, tx)
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
