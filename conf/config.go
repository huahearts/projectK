package conf

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ListenAddress string `yaml:"listen_address"`
	Port          int    `yaml:"port"`
	LogFilePath   string `yaml:"logfilepath"`
	Name          string `yaml:"name"`
}

func ReadConfig() (*Config, error) {
	file, err := os.Open("conf/config.yaml")
	if err != nil {
		fmt.Println("file is not exist")
		return nil, err
	}

	defer file.Close()

	config := &Config{}
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}
	return config, nil
}

func (c Config) GetListenAddress() string {
	return c.ListenAddress
}

func (c Config) GetPort() int {
	return c.Port
}
