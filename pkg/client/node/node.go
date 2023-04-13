package clientnode

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	bc "github.com/antavelos/blockchain/pkg/blockchain"
)

const blockchainURL = "blockchain"

type Node struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
	Port string `json:"port"`
}

func (n Node) getHost() string {
	return fmt.Sprintf("%v:%v", n.IP, n.Port)
}

func (n Node) getBlockchainURL() string {
	return fmt.Sprintf("%v:%v/%v", n.IP, n.Port, blockchainURL)
}

func getData(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	return body, err
}

func (n Node) getBlockchain() (*bc.Blockchain, error) {
	body, err := getData(n.getBlockchainURL())
	if err != nil {
		return nil, fmt.Errorf("failed to get blockchain: %v", err.Error())
	}

	var blockchain bc.Blockchain
	if err := json.Unmarshal(body, &blockchain); err != nil {
		return nil, fmt.Errorf("failed to unmarshal blockchain data: %v", err.Error())
	}

	return &blockchain, nil
}
