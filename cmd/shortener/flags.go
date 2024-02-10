package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type osParam struct {
	SrvAdr  string `env:"SERVER_ADDRESS"`
	BaseURL string `env:"BASE_URL"`
}

var (
	flagRunAddr    string
	flagResultAddr string
)

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&flagResultAddr, "b", "http://localhost:8080", "address and port to run server")
	flag.Parse()

	var cfg osParam
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	if cfg.SrvAdr != "" {
		flagRunAddr = cfg.SrvAdr
	}
	if cfg.BaseURL != "" {
		flagResultAddr = cfg.BaseURL
	}
}
