package memory

import "github.com/winkor4/taktaev-yandex-dev-uri.git/internal/storage/memory"

type Repository struct {
	*memory.DB
}

func NewRepository(db *memory.DB) *Repository {
	return &Repository{
		DB: db,
	}
}
