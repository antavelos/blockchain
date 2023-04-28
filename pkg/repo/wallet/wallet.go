package repowallet

import (
	"encoding/json"

	"github.com/antavelos/blockchain/pkg/common"
	database "github.com/antavelos/blockchain/pkg/db"
	w "github.com/antavelos/blockchain/pkg/models/wallet"
)

type WalletRepo struct {
	db *database.DB
}

func NewWalletRepo(db *database.DB) *WalletRepo {
	return &WalletRepo{db: db}
}

func (r *WalletRepo) GetWallets() (wallets []w.Wallet, err error) {
	data, err := r.db.Load()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &wallets)

	return wallets, err
}

func (r *WalletRepo) AddWallet(wallet w.Wallet) error {
	return r.db.WithLock(func(data []byte) (any, error) {
		wallets, _ := w.UnmarshalMany(data)

		return w.AddWallet(wallets, wallet)
	})
}

func (r *WalletRepo) CreateWallet() (*w.Wallet, error) {

	wallet, err := w.NewWallet()
	if err != nil {
		return nil, common.GenericError{Msg: "failed to create a new wallet", Extra: err}
	}

	err = r.db.WithLock(func(data []byte) (any, error) {
		wallets, _ := w.UnmarshalMany(data)

		return append(wallets, *wallet), nil
	})

	return wallet, err
}

func (r *WalletRepo) IsEmpty() bool {
	data, _ := r.db.Load()

	wallets, _ := w.UnmarshalMany(data)

	return len(wallets) == 0
}
