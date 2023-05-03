package config

import (
	cfg "github.com/antavelos/blockchain/src/pkg/config"
	"github.com/antavelos/blockchain/src/pkg/utils"
)

var envVars []string = []string{
	"PORT",
	"WALLET_CREATION_INTERVAL_IN_SEC",
	"TRANSACTION_CREATION_INTERVAL_IN_SEC",
	"WALLETS_FILENAME",
	"DNS_HOST",
	"DNS_PORT",
}

type Config struct {
	c                                cfg.Config
	WalletCreationIntervalInSec      int // 300
	TransactionCreationIntervalInSec int // 4
}

func NewConfig() (*Config, error) {
	config, err := cfg.LoadConfig(envVars)
	if err != nil {
		return nil, utils.GenericError{Msg: "Configuration error", Extra: err}
	}

	return &Config{
		c:                                config,
		WalletCreationIntervalInSec:      config.GetInteger("WALLET_CREATION_INTERVAL_IN_SEC", 300),
		TransactionCreationIntervalInSec: config.GetInteger("TRANSACTION_CREATION_INTERVAL_IN_SEC", 5),
	}, nil
}

func (c *Config) Get(key string) string {
	return c.c[key]
}
