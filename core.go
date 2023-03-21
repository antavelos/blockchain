package blockchain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

const difficulty int = 2
const TxsPerBlock int = 5
const god = "9e7fb2cd-ddc2-4824-8d87-c0238752255b"

type Node struct {
	Host string `json:"host"`
}

type Transaction struct {
	Id        string  `json:"id"`
	Sender    string  `json:"sender"`
	Recipient string  `json:"recipient"`
	Amount    float64 `json:"amount"`
}

type Block struct {
	Idx       int64         `json:"idx"`
	Timestamp int64         `json:"timestamp"`
	Txs       []Transaction `json:"txs"`
	PrevHash  []byte        `json:"prevHash"`
	Nonce     int64         `json:"nonce"`
}

type Blockchain struct {
	Blocks []Block       `json:"block"`
	TxPool []Transaction `json:"txPool"`
}

func (bc *Blockchain) RemoveTx(tx Transaction) {
	index := -1
	for i, bcTx := range bc.TxPool {
		if tx.Id == bcTx.Id {
			index = i
			break
		}
	}

	if index > -1 {
		bc.TxPool = append(bc.TxPool[:index], bc.TxPool[index+1:]...)
	}
}

func (bc *Blockchain) AddBlock(block Block) {
	if !verifyBlock(block) {
		return
	}

	bc.Blocks = append(bc.Blocks, block)

	for _, tx := range block.Txs {
		bc.RemoveTx(tx)
	}
}

func (bc *Blockchain) CreateGenesisBlock() {
	genesisBlock := Block{
		Idx:       1,
		Timestamp: time.Now().UnixMilli(),
		PrevHash:  []byte{},
		Nonce:     0,
	}

	bc.Blocks = append(bc.Blocks, genesisBlock)
}

func (bc *Blockchain) AddTx(tx Transaction) (Transaction, error) {
	if tx.Sender != god && !bc.validateTransaction(tx) {
		return Transaction{}, fmt.Errorf(
			"transaction of %v units from %v to %v is not valid. Sender has not enough units", tx.Amount, tx.Sender, tx.Recipient)
	}
	tx.Id = NewUuid()
	bc.TxPool = append(bc.TxPool, tx)

	return tx, nil
}

func (bc *Blockchain) NewBlock() (Block, error) {
	txPoolLength := len(bc.TxPool)

	if txPoolLength == 0 {
		return Block{}, errors.New("no pending transactions found")
	}

	lastBlock := bc.Blocks[len(bc.Blocks)-1]

	txCount := TxsPerBlock
	if txPoolLength < TxsPerBlock {
		txCount = txPoolLength
	}
	var latestTxs []Transaction
	for i := 0; i < txCount; i++ {
		latestTxs = append(latestTxs, bc.TxPool[i])
	}

	newBlock := Block{
		Idx:       lastBlock.Idx + 1,
		Timestamp: time.Now().UnixMilli(),
		Txs:       latestTxs,
		PrevHash:  hashBlock(lastBlock),
		Nonce:     0,
	}

	log.Printf("Mining...")
	for !verifyBlock(newBlock) {
		newBlock.Nonce += 1
	}
	log.Printf("Found Nonce: %v", newBlock.Nonce)

	return newBlock, nil
}

func verifyBlock(block Block) bool {
	hashed := hashBlock(block)

	prefix := []byte(strings.Repeat("0", difficulty))

	return bytes.Equal(hashed[:difficulty], prefix)
}

func userBalanceFromTransaction(user string, tx Transaction) float64 {
	if user == tx.Recipient {
		return tx.Amount
	}

	if user == tx.Sender {
		return -tx.Amount
	}

	return 0.0
}

func (bc Blockchain) validateTransaction(tx Transaction) bool {
	senderBalance := 0.0

	for _, ptx := range bc.TxPool {
		senderBalance += userBalanceFromTransaction(tx.Sender, ptx)
	}

	for _, block := range bc.Blocks {
		for _, btx := range block.Txs {
			senderBalance += userBalanceFromTransaction(tx.Sender, btx)
		}
	}

	return tx.Amount <= senderBalance
}

func isValid(bc Blockchain) bool {
	if len(bc.Blocks) == 1 {
		return true
	}

	for i := 1; i < len(bc.Blocks); i++ {
		if !bytes.Equal(bc.Blocks[i].PrevHash, hashBlock(bc.Blocks[i-1])) {
			return false
		}
	}

	return true
}

func getLastBlock(bc Blockchain) Block {
	blocksNum := len(bc.Blocks)

	if blocksNum == 0 {
		return Block{}
	}

	return bc.Blocks[blocksNum-1]
}

func hashBlock(block Block) []byte {
	jsonBlock, err := json.Marshal(block)
	if err != nil {
		return []byte{}
	}

	return hash(jsonBlock)
}
