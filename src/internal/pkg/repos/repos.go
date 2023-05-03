package repos

import "github.com/antavelos/blockchain/src/pkg/db"

type DBFilenames struct {
	BlockchainFilename string
	NodeFilename       string
	WalletFilename     string
}

type Repos struct {
	BlockchainRepo *BlockchainRepo
	NodeRepo       *NodeRepo
	WalletRepo     *WalletRepo
}

func InitRepos(filenames DBFilenames) *Repos {
	return &Repos{
		BlockchainRepo: NewBlockchainRepo(db.NewDB(filenames.BlockchainFilename)),
		NodeRepo:       NewNodeRepo(db.NewDB(filenames.NodeFilename)),
		WalletRepo:     NewWalletRepo(db.NewDB(filenames.WalletFilename)),
	}
}
