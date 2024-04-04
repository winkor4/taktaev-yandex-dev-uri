package psql

import (
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/storage/psql"
)

type Repository struct {
	*psql.DB
}

func NewRepository(db *psql.DB) *Repository {
	return &Repository{
		DB: db,
	}
}
