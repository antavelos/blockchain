package main

import (
	"bytes"
	"crypto/sha512"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"
)

const difficulty int = 2
const TxsPerBlock int = 5
const god = "9e7fb2cd-ddc2-4824-8d87-c0238752255b"

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

func (bc Blockchain) hasTx(tx Transaction) bool {
	for _, bcTx := range bc.TxPool {
		if tx.Id == bcTx.Id {
			return true
		}
	}

	return false
}

func (bc *Blockchain) removeTx(tx Transaction) {
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

func (bc *Blockchain) addBlock(block Block) {
	if !verifyBlock(block) {
		return
	}

	bc.Blocks = append(bc.Blocks, block)

	for _, tx := range block.Txs {
		bc.removeTx(tx)
	}
}

func (bc *Blockchain) createGenesisBlock() {
	genesisBlock := Block{
		Idx:       1,
		Timestamp: time.Now().UnixMilli(),
		PrevHash:  []byte{},
		Nonce:     0,
	}

	bc.Blocks = append(bc.Blocks, genesisBlock)
}

func (bc *Blockchain) addTx(tx Transaction) {
	if tx.Sender != god && !bc.validateTransaction(tx) {
		log.Printf("Transaction of %v units from %v to %v is not valid. Sender has not enough units.", tx.Amount, tx.Sender, tx.Recipient)
		return
	}
	bc.TxPool = append(bc.TxPool, tx)
}

func (bc *Blockchain) newBlock() (Block, error) {
	if len(bc.TxPool) < TxsPerBlock {
		return Block{}, errors.New("not enough transactions yet to create a block")
	}

	lastBlock := bc.Blocks[len(bc.Blocks)-1]

	var latestTxs []Transaction
	for i := 0; i < TxsPerBlock; i++ {
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

	if !bytes.Equal(hashed[:difficulty], prefix) {
		return false
	}

	return true
}

func (bc Blockchain) validateTransaction(tx Transaction) bool {
	senderBalance := 0.0
	for _, block := range bc.Blocks {
		for _, btx := range block.Txs {
			if tx.Sender == btx.Recipient {
				senderBalance += btx.Amount
			}
		}
	}

	if tx.Amount > senderBalance {
		return false
	}

	return true
}

func (bc Blockchain) isValid() bool {
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

func hash(hashable []byte) []byte {
	hash := sha512.New()

	hash.Write(hashable)

	return hash.Sum(nil)
}

func hashBlock(block Block) []byte {
	jsonBlock, err := json.Marshal(block)
	if err != nil {
		return []byte{}
	}

	return hash(jsonBlock)
}
