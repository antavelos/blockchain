package blockchain

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/antavelos/blockchain/pkg/common"
	"github.com/antavelos/blockchain/pkg/lib/crypto"
	"github.com/antavelos/blockchain/pkg/models/wallet"
	"github.com/google/uuid"
)

type TransactionBody struct {
	Sender    string  `json:"sender"`
	Recipient string  `json:"recipient"`
	Amount    float64 `json:"amount"`
}

func (txb TransactionBody) getBalanceForAddress(address string) float64 {
	switch address {
	case txb.Recipient:
		return txb.Amount
	case txb.Sender:
		return -txb.Amount
	default:
		return 0.0
	}
}

type Transaction struct {
	Id        string          `json:"id"`
	Timestamp int64           `json:"timestamp"`
	Body      TransactionBody `json:"body"`
	Signature string          `json:"signature"`
}

func NewTransaction(senderWallet wallet.Wallet, recipientWallet wallet.Wallet, amount float64) (Transaction, error) {

	txb := TransactionBody{
		Sender:    senderWallet.AddressString(),
		Recipient: recipientWallet.AddressString(),
		Amount:    amount,
	}

	txbBytes, err := json.Marshal(txb)
	if err != nil {
		return Transaction{}, common.GenericError{Msg: "failed to marshal transaction body", Extra: err}
	}

	signature, err := senderWallet.Sign(txbBytes)
	if err != nil {
		return Transaction{}, common.GenericError{Msg: "failed to sign transaction body", Extra: err}
	}

	return Transaction{
		Body:      txb,
		Signature: signature,
	}, nil
}

func (tx Transaction) isCoinbase() bool {
	return tx.Body.Sender == "0"
}

func (tx Transaction) Validate() error {
	if tx.isCoinbase() {
		return nil
	}

	txBodyBytes, err := json.Marshal(tx.Body)
	if err != nil {
		return common.GenericError{Msg: "failed to marshal transaction body"}
	}

	signatureBytes, err := hex.DecodeString(tx.Signature)
	if err != nil {
		return common.GenericError{Msg: "failed to decode signature"}
	}

	publicKeyBytes, err := crypto.PublicKeyFromSignature(txBodyBytes, signatureBytes)
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

	if !crypto.VerifySignature(txBodyBytes, publicKeyBytes, signatureBytes) {
		return common.GenericError{Msg: "failed to verify signature"}
	}

	return nil
}

type Block struct {
	Idx       int64         `json:"idx"`
	Timestamp int64         `json:"timestamp"`
	Txs       []Transaction `json:"txs"`
	PrevHash  []byte        `json:"prevHash"`
	Nonce     int64         `json:"nonce"`
}

func (b *Block) HasTx(tx Transaction) bool {
	for _, bTx := range b.Txs {
		if tx.Id == bTx.Id {
			return true
		}
	}
	return false
}

func (b *Block) IsValid(difficulty int) bool {
	hashedBlock := b.hash()

	prefix := []byte(strings.Repeat("0", difficulty))

	return bytes.Equal(hashedBlock[:difficulty], prefix)
}

func (b Block) hash() []byte {
	blockBytes, _ := json.Marshal(b)

	return crypto.HashData(blockBytes)
}

type Blockchain struct {
	Blocks []Block       `json:"block"`
	TxPool []Transaction `json:"txPool"`
}

func NewBlockchain() *Blockchain {
	var blockchain Blockchain
	blockchain.createGenesisBlock()

	return &blockchain
}

func UnmarshalBlockchain(data []byte) (blockchain Blockchain, err error) {
	err = json.Unmarshal(data, &blockchain)
	return
}

func UnmarshalTransaction(data []byte) (tx Transaction, err error) {
	err = json.Unmarshal(data, &tx)
	return
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

func (bc *Blockchain) AddBlock(block Block) error {
	if !bc.verifyBlockHash(block) {
		return common.GenericError{Msg: "block.PrevHash does not match with last block's hash"}
	}

	bc.Blocks = append(bc.Blocks, block)
	bc.removeTxs(block.Txs)

	return nil
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

func (bc *Blockchain) AddTx(tx Transaction) (Transaction, error) {
	if !bc.verifyTxSenderBalance(tx) {
		return Transaction{}, common.GenericError{Msg: "sender has not sufficient funds"}
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

func (bc *Blockchain) HasPendingTxs() bool {
	return len(bc.TxPool) > 0
}

func (bc *Blockchain) NewBlock(txsPerBlock int) (Block, error) {
	txPoolLength := len(bc.TxPool)

	txCount := txsPerBlock
	if txPoolLength < txsPerBlock {
		txCount = txPoolLength
	}

	latestTxs := make([]Transaction, txCount)
	copy(latestTxs, bc.TxPool[:txCount])

	lastBlock := bc.lastBlock()
	newBlock := Block{
		Idx:       lastBlock.Idx + 1,
		Timestamp: time.Now().UnixMilli(),
		Txs:       latestTxs,
		PrevHash:  lastBlock.hash(),
		Nonce:     0,
	}

	return newBlock, nil
}

func (bc *Blockchain) verifyBlockHash(block Block) bool {
	return bytes.Equal(block.PrevHash, bc.lastBlock().hash())
}

func (bc Blockchain) verifyTxSenderBalance(tx Transaction) bool {
	if tx.isCoinbase() {
		return true
	}

	return tx.Body.Amount <= bc.getSenderBalance(tx.Body.Sender)
}

func (bc Blockchain) getSenderBalance(sender string) float64 {
	senderBalance := 0.0

	for _, poolTx := range bc.TxPool {
		senderBalance += poolTx.Body.getBalanceForAddress(sender)
	}

	for _, block := range bc.Blocks {
		for _, blockTx := range block.Txs {
			senderBalance += blockTx.Body.getBalanceForAddress(sender)
		}
	}

	return senderBalance
}

// TODO: to be used
func (bc *Blockchain) IsValid() bool {
	if len(bc.Blocks) == 1 {
		return true
	}

	for i := 1; i < len(bc.Blocks); i++ {
		currBlock := bc.Blocks[i]
		prevBlock := bc.Blocks[i-1]

		if !bytes.Equal(currBlock.PrevHash, prevBlock.hash()) {
			return false
		}
	}

	return true
}

func (bc *Blockchain) lastBlock() Block {
	blocksNum := len(bc.Blocks)

	if blocksNum == 0 {
		return Block{}
	}

	return bc.Blocks[blocksNum-1]
}

func (bc *Blockchain) Update(other *Blockchain) {
	// TODO: append the blocks diff
	bc.Blocks = other.Blocks

	// TODO: to refactor
	for i := len(bc.Blocks) - 1; i > 0; i-- {
		for _, tx := range bc.TxPool {
			if bc.Blocks[i].HasTx(tx) {
				bc.removeTx(tx)
			}
		}
	}
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
