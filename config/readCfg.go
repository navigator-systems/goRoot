package config

import (
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	// ServerConfig
	Port          int    `toml:"port"`
	Directory     string `toml:"directory"`
	Namespace     string `toml:"namespace"`
	Command       string `toml:"command"`
	Image         string
	SharedStorage string
}

const (
	FileName = "config.toml"
)

func ReadFile() Config {
	// Read the config file
	if _, err := os.Stat(FileName); err != nil {
		log.Fatal("Config file not found.")
	}

	var cfg Config
	if _, err := toml.DecodeFile(FileName, &cfg); err != nil {
		log.Fatal(err)
	}
	return cfg
}
