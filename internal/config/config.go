package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	SrvAdr          string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	LogLevel        zapcore.Level
}

var (
	flagRunAddr         string
	flagResultAddr      string
	flagLogLevel        string
	flagFileStoragePath string
)

func stringVar(p *string, name string, value string, usage string) {
	if flag.Lookup(name) == nil {
		flag.StringVar(p, name, value, usage)
	}
}

func Parse() (*Config, error) {
	stringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	stringVar(&flagResultAddr, "b", "http://localhost:8080", "address and port to run server")
	stringVar(&flagLogLevel, "l", "info", "log level")
	stringVar(&flagFileStoragePath, "f", "tmp/short-url-db.json", "file storage path")
	flag.Parse()

	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	if cfg.SrvAdr == "" {
		cfg.SrvAdr = flagRunAddr
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = flagResultAddr
	}
	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = flagFileStoragePath
	}
	cfg.LogLevel, err = zapcore.ParseLevel(flagLogLevel)
	if err != nil {
		return nil, err
	}

	return &cfg, err
}
