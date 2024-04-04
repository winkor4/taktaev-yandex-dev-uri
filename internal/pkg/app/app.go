package app

import (
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/adapter/file"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/adapter/memory"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/adapter/psql"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/log"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/pkg/config"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/server"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/storage"
	sfile "github.com/winkor4/taktaev-yandex-dev-uri.git/internal/storage/file"
	smemory "github.com/winkor4/taktaev-yandex-dev-uri.git/internal/storage/memory"
	spsql "github.com/winkor4/taktaev-yandex-dev-uri.git/internal/storage/psql"
)

func Run() error {
	cfg, err := config.Parse()
	if err != nil {
		return err
	}

	repo, close, err := newRepo(cfg)
	if err != nil {
		return err
	}
	defer close()

	logger, err := log.New()
	if err != nil {
		return err
	}
	defer logger.Close()

	srv := server.New(server.Config{
		URLRepo:    repo,
		Cfg:        cfg,
		Logger:     logger,
	})

	return srv.Run()
}

type dbCloser func() error

func newRepo(cfg *config.Config) (storage.Repository, dbCloser, error) {
	var (
		repo  storage.Repository
		close dbCloser
	)

	switch {
	case cfg.DSN != "":
		db, err := spsql.New(cfg.DSN)
		if err != nil {
			return nil, nil, err
		}
		close = db.CloseDB
		repo = psql.NewRepository(db)
	case cfg.FileStoragePath != "":
		db, err := sfile.New(cfg.FileStoragePath)
		if err != nil {
			return nil, nil, err
		}
		close = db.CloseFile
		repo = file.NewRepository(db)
	default:
		db := smemory.New()
		close = db.Close
		repo = memory.NewRepository(db)
	}

	return repo, close, nil
}
