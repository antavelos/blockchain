package blockchain

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/antavelos/blockchain/src/crypto"
	"github.com/google/uuid"
)

const difficulty int = 2
const TxsPerBlock int = 5

type Node struct {
	Host string `json:"host"`
}

func (n Node) GetPort() string {
	return n.Host[len(n.Host)-4:]
}

type TransactionBody struct {
	Sender    string  `json:"sender"`
	Recipient string  `json:"recipient"`
	Amount    float64 `json:"amount"`
}

type Transaction struct {
	Id        string          `json:"id"`
	Timestamp int64           `json:"timestamp"`
	Body      TransactionBody `json:"body"`
	Signature string          `json:"signature"`
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

func NewTransactionBody(sender string, recipient string, amount float64) *TransactionBody {
	return &TransactionBody{
		Sender:    sender,
		Recipient: recipient,
		Amount:    amount,
	}
}

func (tx Transaction) GetBodyHash() ([]byte, error) {
	marshalled, err := json.Marshal(tx.Body)
	if err != nil {
		return nil, err
	}

	return crypto.HashData(marshalled), nil
}

func (tx Transaction) IsCoinbase() bool {
	return tx.Body.Sender == "0"
}

func (b *Block) HasTx(tx Transaction) bool {
	for _, bTx := range b.Txs {
		if tx.Id == bTx.Id {
			return true
		}
	}
	return false
}

func (bc *Blockchain) RemoveTx(tx Transaction) {
	for i, bcTx := range bc.TxPool {
		if tx.Id == bcTx.Id {
			bc.TxPool = append(bc.TxPool[:i], bc.TxPool[i+1:]...)
			break
		}
	}
}

func (bc *Blockchain) RemoveTxs(txs []Transaction) {
	for _, tx := range txs {
		bc.RemoveTx(tx)
	}
}

func (bc *Blockchain) AddBlock(block Block) {
	if !verifyBlock(block) {
		return
	}

	bc.Blocks = append(bc.Blocks, block)
	bc.RemoveTxs(block.Txs)
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

func (bc *Blockchain) AddTx(tx Transaction) (Transaction, error) {
	if err := bc.validateTransaction(tx); err != nil {
		return Transaction{}, err
	}

	tx.Id = uuid.NewString()
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

	txCount := TxsPerBlock
	if txPoolLength < TxsPerBlock {
		txCount = txPoolLength
	}

	latestTxs := make([]Transaction, txCount)
	copy(latestTxs, bc.TxPool[:txCount])

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

func getAddressBalanceFromTransactionBody(address string, txb TransactionBody) float64 {
	switch address {
	case txb.Recipient:
		return txb.Amount
	case txb.Sender:
		return -txb.Amount
	default:
		return 0.0
	}
}

func (bc Blockchain) validateTransaction(tx Transaction) error {
	if tx.IsCoinbase() {
		return nil
	}

	txBodyBytes, err := json.Marshal(tx.Body)
	if err != nil {
		return errors.New("failed to marshal transaction body")
	}

	txBodyHash := crypto.HashData(txBodyBytes)

	signatureBytes, err := hex.DecodeString(tx.Signature)
	if err != nil {
		return errors.New("failed to decode signature")
	}

	publicKeyBytes, err := crypto.PublicKeyFromSignature(txBodyHash, signatureBytes)
	if err != nil {
		return errors.New("failed to retrieve public key from signature")
	}

	publicKey, err := crypto.UnmarshalPublicKey(publicKeyBytes)
	if err != nil {
		return errors.New("failed to unmarshal public key")
	}

	senderBytes, err := hex.DecodeString(tx.Body.Sender)
	if err != nil {
		return errors.New("failed to decode sender")
	}

	senderAddress := crypto.AddressFromPublicKey(publicKey)
	if !bytes.Equal(senderAddress, senderBytes) {
		return errors.New("sender address does not match with the public key of the signature")
	}

	if !crypto.VerifySignature(txBodyHash, publicKeyBytes, signatureBytes) {
		return errors.New("failed to verify signature")
	}

	senderBalance := bc.getSenderBalance(tx.Body.Sender)
	if tx.Body.Amount <= senderBalance {
		return errors.New("sender has not sufficient funds")
	}

	return nil
}

func (bc Blockchain) getSenderBalance(sender string) float64 {
	senderBalance := 0.0

	for _, ptx := range bc.TxPool {
		senderBalance += getAddressBalanceFromTransactionBody(sender, ptx.Body)
	}

	for _, block := range bc.Blocks {
		for _, btx := range block.Txs {
			senderBalance += getAddressBalanceFromTransactionBody(sender, btx.Body)
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

	return crypto.HashData(jsonBlock), nil
}
