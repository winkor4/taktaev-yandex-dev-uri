package main

import (
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/config"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/databaseSQL"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/handlers"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/logger"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/storage"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		panic(err)
	}
	sm, err := storage.NewStorageMap(cfg.FileStoragePath)
	if err != nil {
		panic(err)
	}
	l, err := logger.NewLogZap()
	if err != nil {
		panic(err)
	}
	db, err := databasesql.CheckConn(cfg.DatabaseDSN)
	if err != nil {
		panic(err)
	}
	sm.DB = db
	hd := handlers.HandlerData{
		SM:  sm,
		Cfg: cfg,
		L:   l,
	}

	defer hd.SM.CloseStorageFile()
	defer db.Close()

	l.Logw(cfg.LogLevel, "Starting server", "SrvAdr", cfg.SrvAdr)

	err = http.ListenAndServe(cfg.SrvAdr, hd.URLRouter())
	if err != nil {
		panic(err)
	}
}
