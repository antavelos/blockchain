package main

import (
	"fmt"
	"os"

	"github.com/antavelos/blockchain/pkg/common"
)

type Config map[string]string

func getEnvVar(conf Config, envVar string) error {
	value, found := os.LookupEnv(envVar)
	if !found {
		return common.GenericError{Msg: fmt.Sprintf("env variable '%v' not set", value)}
	}
	conf[envVar] = value

	return nil
}

func getConfig() (Config, error) {
	conf := make(Config)

	envVars := []string{
		"PORT",
		"DNS_HOST",
		"DNS_PORT",
		"WALLETS_HOST",
		"WALLETS_PORT",
	}

	for _, envVar := range envVars {
		err := getEnvVar(conf, envVar)
		if err != nil {
			return conf, err
		}
	}

	return conf, nil
}
