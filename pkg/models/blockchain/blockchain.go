package blockchain

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/antavelos/blockchain/pkg/common"
	"github.com/antavelos/blockchain/pkg/lib/crypto"
	"github.com/antavelos/blockchain/pkg/lib/rest"
	"github.com/antavelos/blockchain/pkg/models/wallet"
	"github.com/google/uuid"
)

func getMiningDifficulty() int {
	diff, err := strconv.Atoi(os.Getenv("MINING_DIFFICULTY"))
	if err != nil {
		return 2
	}
	return diff
}

func getTxsPerBlock() int {
	txsPerBlock, err := strconv.Atoi(os.Getenv("TXS_PER_BLOCK"))
	if err != nil {
		return 5
	}

	return txsPerBlock
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

type TxMarshaller rest.ObjectMarshaller

func (tm TxMarshaller) Unmarshal(data []byte) (any, error) {
	var target any
	if tm.Many {
		target = make([]Transaction, 0)
	} else {
		target = Transaction{}
	}

	err := json.Unmarshal(data, &target)

	return target, err
}

type BlockchainMarshaller rest.ObjectMarshaller

func (bm BlockchainMarshaller) Unmarshal(data []byte) (any, error) {
	var target any
	if bm.Many {
		target = make([]Blockchain, 0)
	} else {
		target = Blockchain{}
	}

	err := json.Unmarshal(data, &target)

	return target, err
}

func NewTransaction(senderWallet wallet.Wallet, recipientWallet wallet.Wallet, amount float64) (Transaction, error) {

	txb := TransactionBody{
		Sender:    hex.EncodeToString(senderWallet.Address),
		Recipient: hex.EncodeToString(recipientWallet.Address),
		Amount:    amount,
	}

	txbBytes, err := json.Marshal(txb)
	if err != nil {
		return Transaction{}, common.GenericError{Msg: "failed to marshal transaction body", Extra: err}
	}

	signature, err := senderWallet.Sign(crypto.HashData(txbBytes))
	if err != nil {
		return Transaction{}, common.GenericError{Msg: "failed to sign transaction body", Extra: err}
	}

	return Transaction{
		Body:      txb,
		Signature: signature,
	}, nil
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

func (bc *Blockchain) AddBlock(block Block) error {
	if err := bc.validateBlock(block); err != nil {
		return common.GenericError{Msg: "failed to validate block"}
	}

	bc.Blocks = append(bc.Blocks, block)
	bc.RemoveTxs(block.Txs)

	return nil
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

	if tx.Id == "" {
		tx.Id = uuid.NewString()
	}

	if tx.Timestamp == 0 {
		tx.Timestamp = time.Now().UnixMilli()
	}
	bc.TxPool = append(bc.TxPool, tx)

	return tx, nil
}

func (bc *Blockchain) NewBlock() (Block, error) {
	txPoolLength := len(bc.TxPool)

	if txPoolLength == 0 {
		return Block{}, common.GenericError{Msg: "no pending transactions found"}
	}

	lastBlock := bc.Blocks[len(bc.Blocks)-1]

	txsPerBlock := getTxsPerBlock()
	txCount := txsPerBlock
	if txPoolLength < txsPerBlock {
		txCount = txPoolLength
	}

	latestTxs := make([]Transaction, txCount)
	copy(latestTxs, bc.TxPool[:txCount])

	hashedLastBlock, err := hashBlock(lastBlock)
	if err != nil {
		return Block{}, common.GenericError{Msg: "failed to hash last block"}
	}
	newBlock := Block{
		Idx:       lastBlock.Idx + 1,
		Timestamp: time.Now().UnixMilli(),
		Txs:       latestTxs,
		PrevHash:  hashedLastBlock,
		Nonce:     0,
	}

	common.LogInfo("Mining...")
	for !blockSatisfiesHashRule(newBlock) {
		newBlock.Nonce += 1
	}
	common.LogInfo("Found Nonce: %v", newBlock.Nonce)

	return newBlock, nil
}

func GetMaxLengthBlockchain(blockchains []*Blockchain) *Blockchain {
	if len(blockchains) == 0 {
		return &Blockchain{}
	}

	maxLengthBlockchain := blockchains[0]

	for _, blockchain := range blockchains[1:] {
		if maxLengthBlockchain == nil || len(blockchain.Blocks) > len(maxLengthBlockchain.Blocks) {
			maxLengthBlockchain = blockchain
		}
	}

	return maxLengthBlockchain
}

func blockSatisfiesHashRule(block Block) bool {
	hashed, _ := hashBlock(block)

	difficulty := getMiningDifficulty()

	prefix := []byte(strings.Repeat("0", difficulty))

	return bytes.Equal(hashed[:difficulty], prefix)
}

func (bc *Blockchain) validateBlock(block Block) error {
	difficulty := getMiningDifficulty()

	if !blockSatisfiesHashRule(block) {
		msg := fmt.Sprintf("block does not start with %v '0'", difficulty)
		return common.GenericError{Msg: msg}
	}

	lastBlockHashed, _ := hashBlock(bc.Blocks[len(bc.Blocks)-1])

	if !bytes.Equal(block.PrevHash, lastBlockHashed) {
		return common.GenericError{Msg: "block.PrevHash does not match with last block's hash"}
	}

	return nil
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
		return common.GenericError{Msg: "failed to marshal transaction body"}
	}

	txBodyHash := crypto.HashData(txBodyBytes)

	signatureBytes, err := hex.DecodeString(tx.Signature)
	if err != nil {
		return common.GenericError{Msg: "failed to decode signature"}
	}

	publicKeyBytes, err := crypto.PublicKeyFromSignature(txBodyHash, signatureBytes)
	if err != nil {
		return common.GenericError{Msg: "failed to retrieve public key from signature"}
	}

	publicKey, err := crypto.UnmarshalPublicKey(publicKeyBytes)
	if err != nil {
		return common.GenericError{Msg: "failed to unmarshal public key"}
	}

	senderBytes, err := hex.DecodeString(tx.Body.Sender)
	if err != nil {
		return common.GenericError{Msg: "failed to decode sender"}
	}

	senderAddress := crypto.AddressFromPublicKey(publicKey)
	if !bytes.Equal(senderAddress, senderBytes) {
		return common.GenericError{Msg: "sender address does not match with the public key of the signature"}
	}

	if !crypto.VerifySignature(txBodyHash, publicKeyBytes, signatureBytes) {
		return common.GenericError{Msg: "failed to verify signature"}
	}

	senderBalance := bc.getSenderBalance(tx.Body.Sender)
	if tx.Body.Amount <= senderBalance {
		return common.GenericError{Msg: "sender has not sufficient funds"}
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

// TODO: to be used
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
