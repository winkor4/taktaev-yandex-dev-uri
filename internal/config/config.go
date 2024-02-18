package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	SrvAdr  string `env:"SERVER_ADDRESS"`
	BaseURL string `env:"BASE_URL"`
}

var (
	flagRunAddr    string
	flagResultAddr string
)

func Parse() *Config {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&flagResultAddr, "b", "http://localhost:8080", "address and port to run server")
	flag.Parse()

	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	if cfg.SrvAdr == "" {
		cfg.SrvAdr = flagRunAddr
	}
	if cfg.SrvAdr == "" {
		cfg.BaseURL = flagResultAddr
	}

	return &cfg
}
