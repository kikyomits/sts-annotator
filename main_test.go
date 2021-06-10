package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	dir, err := filepath.Abs("./")
	if err != nil {
		panic(err)
	}

	c := newConstant()
	configFilePath := fmt.Sprintf("%s/test/config.yaml", dir)
	setEnvErr := os.Setenv(c.EnvKeyConfigPath, configFilePath)
	if setEnvErr != nil {
		panic(setEnvErr)
	}
	exitVal := m.Run()
	os.Exit(exitVal)
}

func initTestConfig(k8sApiUrl string) *Config {
	config := initConfig()
	config.K8s.URL = k8sApiUrl

	dir, err := filepath.Abs("./")
	if err != nil {
		panic(err)
	}
	config.Server.Tls.Cert = fmt.Sprintf("%s/test/test.crt", dir)
	config.Server.Tls.Key = fmt.Sprintf("%s/test/test.key", dir)
	config.K8s.Tls.CaCert = fmt.Sprintf("%s/test/test.crt", dir)
	config.K8s.Token = fmt.Sprintf("%s/test/token", dir)
	return config
}
