package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/antavelos/blockchain/src/pkg/utils"
)

type Config map[string]string

func (c Config) GetInteger(key string, defaultVal int) int {
	value, err := strconv.Atoi(c[key])
	if err != nil {
		msg := fmt.Sprintf("Couldn't parse '%v' config value. Using default value: %v", key, defaultVal)
		utils.LogInfo(msg)
		return defaultVal
	}

	return value
}

func (c Config) GetFloat(key string, defaultVal float64) float64 {
	value, err := strconv.ParseFloat(c[key], 1)
	if err != nil {
		msg := fmt.Sprintf("Couldn't parse '%v' config value. Using default value: %v", key, defaultVal)
		utils.LogInfo(msg)
		return defaultVal
	}

	return value
}

func getEnvVar(envVarKey string) (string, error) {
	value, found := os.LookupEnv(envVarKey)
	if !found {
		return "", utils.GenericError{Msg: fmt.Sprintf("env variable '%v' not set", envVarKey)}
	}

	return value, nil
}

func LoadConfig(envVarKeys []string) (Config, error) {
	conf := make(Config)

	for _, key := range envVarKeys {
		value, err := getEnvVar(key)
		if err != nil {
			return conf, err
		}
		conf[key] = value
	}

	return conf, nil
}
