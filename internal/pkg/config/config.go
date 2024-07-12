// Модуль config парсит конфигурацию для сервера.
package config

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/caarlos0/env/v6"
	"go.uber.org/zap/zapcore"
)

// Config параметры конфигурации
type Config struct {
	SrvAdr          string `env:"SERVER_ADDRESS"`
	ResSrvAdr       string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DSN             string `env:"DATABASE_DSN"`
	EnableHTTPS     bool   `env:"ENABLE_HTTPS"`
	Config          string `env:"CONFIG"`
	TrustedSubnet   string `env:"TRUSTED_SUBNET"`
	LogLevel        zapcore.Level
}

// JSConfig формат для чтения файла с конфигурацией
type JSConfig struct {
	SrvAdr          string `json:"server_address"`
	ResSrvAdr       string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DSN             string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
	TrustedSubnet   string `json:"trusted_subnet"`
}

var (
	flagSrvAdr          string
	flagResSrvAdr       string
	flagFileStoragePath string
	flagDSN             string
	flagEnableHTTPS     string
	flagConfig          string
	flagTrustedSubnet   string
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
	stringVar(&flagEnableHTTPS, "s", "", "возможность включения HTTPS в веб-сервере")
	stringVar(&flagConfig, "c", "", "config from json")
	stringVar(&flagTrustedSubnet, "t", "", "CIDR")
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
	if !cfg.EnableHTTPS {
		cfg.EnableHTTPS = flagEnableHTTPS == "true"
	}
	if cfg.TrustedSubnet == "" {
		cfg.TrustedSubnet = flagTrustedSubnet
	}
	if cfg.Config == "" {
		cfg.Config = flagConfig
	}

	cfg.LogLevel, err = zapcore.ParseLevel(flagLogLevel)
	if err != nil {
		return nil, err
	}

	if cfg.Config != "" {
		readJSConfig(cfg)
	}

	return cfg, nil
}

// readJSConfig парсит файл конфигурации в формате json, если он есть
func readJSConfig(cfg *Config) {

	fname := cfg.Config
	strData, err := os.ReadFile(fname)
	if err != nil {
		return
	}

	fcfg := new(JSConfig)
	if err := json.Unmarshal(strData, fcfg); err != nil {
		return
	}

	if cfg.SrvAdr == "" {
		cfg.SrvAdr = fcfg.SrvAdr
	}
	if cfg.ResSrvAdr == "" {
		cfg.ResSrvAdr = fcfg.ResSrvAdr
	}
	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = fcfg.FileStoragePath
	}
	if cfg.DSN == "" {
		cfg.DSN = fcfg.DSN
	}
	if cfg.TrustedSubnet == "" {
		cfg.TrustedSubnet = fcfg.TrustedSubnet
	}
	if !cfg.EnableHTTPS {
		cfg.EnableHTTPS = fcfg.EnableHTTPS
	}

}
