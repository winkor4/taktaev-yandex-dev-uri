// Модуль config парсит конфигурацию для сервера.
package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	SrvAdr          string `env:"SERVER_ADDRESS"`
	ResSrvAdr       string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DSN             string `env:"DATABASE_DSN"`
	LogLevel        zapcore.Level
}

var (
	flagSrvAdr          string
	flagResSrvAdr       string
	flagFileStoragePath string
	flagDSN             string
	flagLogLevel        string
)

func stringVar(p *string, name string, value string, usage string) {
	if flag.Lookup(name) == nil {
		flag.StringVar(p, name, value, usage)
	}
}

// Parse парсит флаги и параметры ОС.
func Parse() (*Config, error) {

	// dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
	// 	`localhost`, `postgres`, `123`, `shorten_dev`)

	// tmp/short-url-db.json

	stringVar(&flagSrvAdr, "a", "localhost:8080", "address and port to run server")
	stringVar(&flagResSrvAdr, "b", "http://localhost:8080", "address and port to run server")
	stringVar(&flagFileStoragePath, "f", "", "file storage path")
	stringVar(&flagDSN, "d", "", "PostgresSQL path")
	stringVar(&flagLogLevel, "l", "info", "log level")
	flag.Parse()

	cfg := new(Config)
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	if cfg.SrvAdr == "" {
		cfg.SrvAdr = flagSrvAdr
	}
	if cfg.ResSrvAdr == "" {
		cfg.ResSrvAdr = flagResSrvAdr
	}
	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = flagFileStoragePath
	}
	if cfg.DSN == "" {
		cfg.DSN = flagDSN
	}

	cfg.LogLevel, err = zapcore.ParseLevel(flagLogLevel)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
