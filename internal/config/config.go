package config

import (
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env/v6"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	SrvAdr   string `env:"SERVER_ADDRESS"`
	BaseURL  string `env:"BASE_URL"`
	LogLevel zapcore.Level
}

var (
	flagRunAddr    string
	flagResultAddr string
	flagLogLevel   string
)

func ClearCommandLine() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func Parse() (*Config, error) {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&flagResultAddr, "b", "http://localhost:8080", "address and port to run server")
	flag.StringVar(&flagLogLevel, "l", "info", "log level")
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

	cfg.LogLevel, err = zapcore.ParseLevel(flagLogLevel)
	if err != nil {
		return nil, err
	}

	return &cfg, err
}
