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

var whitelist = []string{"alex"}

type Node struct {
	Host string `json:"host"`
}

type Transaction struct {
	Id        string  `json:"id"`
	Timestamp int64   `json:"timestamp"`
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

func (bc *Blockchain) removeTx(tx Transaction) {
	for i, bcTx := range bc.TxPool {
		if tx.Id == bcTx.Id {
			bc.TxPool = append(bc.TxPool[:i], bc.TxPool[i+1:]...)
			break
		}
	}
}

func (bc *Blockchain) removeTxs(txs []Transaction) {
	for _, tx := range txs {
		bc.removeTx(tx)
	}
}

func (bc *Blockchain) AddBlock(block Block) {
	if !verifyBlock(block) {
		return
	}

	bc.Blocks = append(bc.Blocks, block)
	bc.removeTxs(block.Txs)
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

func NewBlockchain() *Blockchain {
	var blockchain Blockchain
	blockchain.CreateGenesisBlock()

	return &blockchain
}

func senderIsWhitelisted(sender string) bool {
	for _, white := range whitelist {
		if white == sender {
			return true
		}
	}
	return false
}

func (bc *Blockchain) AddTx(tx Transaction) (Transaction, error) {
	if !bc.validateTransaction(tx) {
		return Transaction{}, fmt.Errorf(
			"transaction of %v units from %v to %v is not valid. Sender has not enough units", tx.Amount, tx.Sender, tx.Recipient)
	}
	tx.Id = newUuid()
	tx.Timestamp = time.Now().UnixMilli()
	bc.TxPool = append(bc.TxPool, tx)

	return tx, nil
}

func (bc *Blockchain) NewBlock() (Block, error) {
	txPoolLength := len(bc.TxPool)

	if txPoolLength == 0 {
		return Block{}, errors.New("no pending transactions found")
	}

	lastBlock := bc.Blocks[len(bc.Blocks)-1]
	var latestTxs []Transaction

	txCount := TxsPerBlock
	if txPoolLength < TxsPerBlock {
		txCount = txPoolLength
	}
	latestTxs = bc.TxPool[:txCount]

	hashedLastBlock, err := hashBlock(lastBlock)
	if err != nil {
		return Block{}, errors.New("failed to hash last block")
	}
	newBlock := Block{
		Idx:       lastBlock.Idx + 1,
		Timestamp: time.Now().UnixMilli(),
		Txs:       latestTxs,
		PrevHash:  hashedLastBlock,
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
	hashed, err := hashBlock(block)
	if err != nil {
		return false
	}

	prefix := []byte(strings.Repeat("0", difficulty))

	return bytes.Equal(hashed[:difficulty], prefix)
}

func getUserBalanceFromTransaction(user string, tx Transaction) float64 {
	switch user {
	case tx.Recipient:
		return tx.Amount
	case tx.Sender:
		return -tx.Amount
	default:
		return 0.0
	}
}

func (bc Blockchain) validateTransaction(tx Transaction) bool {
	if senderIsWhitelisted(tx.Sender) {
		return true
	}

	senderBalance := bc.getUserBalance(tx.Sender)

	return tx.Amount <= senderBalance
}

func (bc Blockchain) getUserBalance(sender string) float64 {
	senderBalance := 0.0

	for _, ptx := range bc.TxPool {
		senderBalance += getUserBalanceFromTransaction(sender, ptx)
	}

	for _, block := range bc.Blocks {
		for _, btx := range block.Txs {
			senderBalance += getUserBalanceFromTransaction(sender, btx)
		}
	}

	return senderBalance
}

func isValid(bc Blockchain) bool {
	if len(bc.Blocks) == 1 {
		return true
	}

	for i := 1; i < len(bc.Blocks); i++ {
		currBlock := bc.Blocks[i]
		prevBlock := bc.Blocks[i-1]
		hashedPrevBlock, err := hashBlock(prevBlock)

		if err != nil {
			return false
		}

		if !bytes.Equal(currBlock.PrevHash, hashedPrevBlock) {
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

func hashBlock(block Block) ([]byte, error) {
	jsonBlock, err := json.Marshal(block)
	if err != nil {
		return []byte{}, err
	}

	return hash(jsonBlock), nil
}
