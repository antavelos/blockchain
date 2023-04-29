package repoblockchain

import (
	"encoding/json"
	"time"

	bc "github.com/antavelos/blockchain/src/internal/pkg/models/blockchain"
	database "github.com/antavelos/blockchain/src/pkg/db"
	"github.com/antavelos/blockchain/src/pkg/utils"
	"github.com/google/uuid"
)

type BlockchainRepo struct {
	db *database.DB
}

func NewBlockchainRepo(db *database.DB) *BlockchainRepo {
	return &BlockchainRepo{db: db}
}

func (r *BlockchainRepo) GetBlockchain() (blockchain *bc.Blockchain, err error) {
	data, err := r.db.Load()
	if err != nil {
		return &bc.Blockchain{}, err
	}

	err = json.Unmarshal(data, &blockchain)

	return
}

func (r *BlockchainRepo) UpdateBlockchain(other *bc.Blockchain) error {
	return r.db.WithLock(func(data []byte) (any, error) {
		blockchain, _ := bc.UnmarshalBlockchain(data)

		blockchain.Update(other)

		return blockchain, nil
	})
}

func (r *BlockchainRepo) ReplaceBlockchain(other bc.Blockchain) error {
	return r.db.Save(other)
}

func (r *BlockchainRepo) AddTx(tx bc.Transaction) (bc.Transaction, error) {

	if tx.Id == "" {
		tx.Id = uuid.NewString()
	}

	if tx.Timestamp == 0 {
		tx.Timestamp = time.Now().UnixMilli()
	}

	err := r.db.WithLock(func(data []byte) (any, error) {
		blockchain, _ := bc.UnmarshalBlockchain(data)

		_, err := blockchain.AddTx(tx)
		if err != nil {
			return nil, err
		}

		return blockchain, nil
	})

	return tx, err
}

func (r *BlockchainRepo) AddBlock(block bc.Block) error {
	err := r.db.WithLock(func(data []byte) (any, error) {
		blockchain, _ := bc.UnmarshalBlockchain(data)

		err := blockchain.AddBlock(block)
		if err != nil {
			return nil, err
		}

		return blockchain, nil
	})

	return err
}

func (r *BlockchainRepo) CreateBlockchain() (*bc.Blockchain, error) {

	blockchain := bc.NewBlockchain()

	err := r.db.Save(*blockchain)
	if err != nil {
		return nil, utils.GenericError{Msg: "failed to save new wallet"}
	}

	return blockchain, nil
}
