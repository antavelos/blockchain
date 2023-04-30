package config

import (
	cfg "github.com/antavelos/blockchain/src/pkg/config"
	"github.com/antavelos/blockchain/src/pkg/utils"
)

var envVars []string = []string{
	"PORT",
	"DNS_HOST",
	"DNS_PORT",
	"WALLETS_HOST",
	"WALLETS_PORT",
	"NODES_FILENAME",
	"BLOCKCHAIN_FILENAME",
	"WALLETS_FILENAME",
	"MINING_DIFFICULTY",
	"TXS_PER_BLOCK",
	"REWARD_AMOUNT",
	"NODE_NAME",
}

type Config struct {
	c                       cfg.Config
	CoinBaseSenderAddress   string  //= "0"
	DefaultTxsPerBlock      int     //= 10
	DefaultMiningDifficulty int     //= 2
	DefaultRewardAmount     float64 //= 1.0
}

func NewConfig() (*Config, error) {
	config, err := cfg.LoadConfig(envVars)
	if err != nil {
		return nil, utils.GenericError{Msg: "Configuration error", Extra: err}
	}

	return &Config{
		c:                       config,
		CoinBaseSenderAddress:   "0",
		DefaultTxsPerBlock:      config.GetInteger("TXS_PER_BLOCK", 10),
		DefaultMiningDifficulty: config.GetInteger("MINING_DIFFICULTY", 2),
		DefaultRewardAmount:     config.GetFloat("REWARD_AMOUNT", 1.0),
	}, nil
}

func (c *Config) Get(key string) string {
	return c.c[key]
}
