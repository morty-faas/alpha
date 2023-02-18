package config

import (
	log "github.com/sirupsen/logrus"
	viper "github.com/spf13/viper"
)

type Config struct {
	InvokeInstruction string    `mapstructure:"ALPHA_INVOKE"`
	Remote            string    `mapstructure:"ALPHA_REMOTE"`
	Port              int       `mapstructure:"ALPHA_PORT"`
	LogLevel          log.Level `mapstructure:"ALPHA_LOG_LEVEL"`
}

var defaultConfig = &Config{
	InvokeInstruction: "",
	Remote:            "http://127.0.0.1:3000",
	Port:              8080,
	LogLevel:          log.InfoLevel,
}

const (
	configFileName = ".env"
)

// LoadConfig loads configuration variables from both config file and environment variables.
func LoadConfig() (*Config, error) {
	v := viper.New()

	// Get values from config file
	v.SetConfigFile(configFileName)
	err := v.ReadInConfig()
	if err != nil {
		// If the config file is not found, we don't return an error. We just log it as a debug message.
		currentLogLevel := log.GetLevel()
		log.SetLevel(log.DebugLevel)
		log.Debug(err)
		log.SetLevel(currentLogLevel)
	}

	// Get values from environment variables
	v.AutomaticEnv()

	// Set default values
	config := defaultConfig
	err = v.Unmarshal(defaultConfig)
	if err != nil {
		return nil, err
	}

	return config, nil
}
