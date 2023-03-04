package main

import "github.com/thomasgouveia/go-config"

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

var loaderOptions = &config.Options[Config]{
	Format: config.YAML,

	// Configure the loader to lookup for environment
	// variables with the following pattern: ALPHA_*
	EnvEnabled: true,
	EnvPrefix:  "alpha",

	// Configure the loader to search for an alpha.yaml file
	// inside one or more locations defined in `FileLocations`
	FileName:      "alpha",
	FileLocations: []string{"/etc/alpha", "$HOME/.alpha", "."},

	// Inject a default configuration in the loader
	Default: &Config{
		Server: serverConfig{
			LogLevel: "INFO",
			Port:     8080,
		},
		Process: processConfig{
			Command:       "",
			DownstreamURL: "http://127.0.0.1:3000",
		},
	},
}
