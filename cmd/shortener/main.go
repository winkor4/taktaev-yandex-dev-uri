package main

import (
	"net/http"

	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/config"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/handlers"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/storage"
)

func main() {
	cfg := config.Parse()
	sm := storage.NewStorageMap()
	hd := handlers.HandlerData{
		SM:  sm,
		Cfg: cfg,
	}

	err := http.ListenAndServe(cfg.SrvAdr, hd.URLRouter())
	if err != nil {
		panic(err)
	}
}
