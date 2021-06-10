package main

import (
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

func initConfig() (config *Config) {
	// Read config file
	constant := newConstant()
	configFilePath := getEnv(constant.EnvKeyConfigPath, "config.yaml")
	yamlFile := readFile(configFilePath)

	// Load config file
	unmarshalErr := yaml.Unmarshal(yamlFile, &config)
	if unmarshalErr != nil {
		log.Fatal().Err(unmarshalErr).Msgf("Failed to unmarshal config yaml file.")
	}
	return
}
