package file

import "github.com/winkor4/taktaev-yandex-dev-uri.git/internal/storage/file"

type Repository struct {
	*file.DB
}

func NewRepository(db *file.DB) *Repository {
	return &Repository{
		DB: db,
	}
}
