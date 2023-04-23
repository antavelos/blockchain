package config

import (
	"fmt"
	"os"

	"github.com/antavelos/blockchain/pkg/common"
)

type Config map[string]string

func getEnvVar(conf Config, envVar string) error {
	value, found := os.LookupEnv(envVar)
	if !found {
		return common.GenericError{Msg: fmt.Sprintf("env variable '%v' not set", envVar)}
	}
	conf[envVar] = value

	return nil
}

func LoadConfig(envVars []string) (Config, error) {
	conf := make(Config)

	for _, envVar := range envVars {
		err := getEnvVar(conf, envVar)
		if err != nil {
			return conf, err
		}
	}

	return conf, nil
}
