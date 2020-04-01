package main

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

func initConfig() (config *Config) {
	// Read config file
	path := "config.yaml"
	yamlFile, ioErr := ioutil.ReadFile(path)
	if ioErr != nil {
		log.Printf("Failed to read a config file. Expected path: %s. Error: %v", path, ioErr)
		return nil
	}

	// Load config file
	unmarshalErr := yaml.Unmarshal(yamlFile, &config)
	if unmarshalErr != nil {
		log.Fatalf("Unmarshal: %v", unmarshalErr)
		return nil
	}
	return
}
