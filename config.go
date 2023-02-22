package main

import (
	"bytes"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"strings"
)

type (
	Config struct {
		Server  serverConfig  `yaml:"server"`
		Process processConfig `yaml:"process"`
	}

	serverConfig struct {
		Port     int    `yaml:"port"`
		LogLevel string `yaml:"logLevel"`
	}

	processConfig struct {
		Command       string `yaml:"command"`
		DownstreamURL string `yaml:"downstreamUrl"`
	}
)

var defaultConfig = &Config{
	Server: serverConfig{
		LogLevel: "INFO",
		Port:     8080,
	},
	Process: processConfig{
		Command:       "",
		DownstreamURL: "http://127.0.0.1:3000",
	},
}

const (
	configFileName = "alpha"
	configFileType = "yaml"
)

// LoadConfig loads configuration variables from both config file and environment variables.
func LoadConfig() (*Config, error) {
	v := viper.New()

	v.SetConfigType(configFileType)
	v.SetConfigName(configFileName)

	b, err := yaml.Marshal(defaultConfig)
	if err != nil {
		return nil, err
	}

	re := bytes.NewReader(b)
	if err := v.MergeConfig(re); err != nil {
		return nil, err
	}

	// Overwrite values from configuration files
	// This will check for an alpha.yaml file in the directories listed below
	for _, path := range []string{"/etc/alpha", "$HOME/.alpha", "."} {
		v.AddConfigPath(path)
	}

	// Get values from environment variables
	v.SetEnvPrefix(strings.ToUpper("ALPHA"))
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Parse configuration from all additional sources
	if err := v.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// Finally unmarshal the Viper loaded configuration into our config struct
	config := defaultConfig
	if err := v.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}
