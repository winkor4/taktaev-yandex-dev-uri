package file

import "github.com/winkor4/taktaev-yandex-dev-uri.git/internal/storage/file"

// Repository соответствует интерфйсу хранилища для фала
type Repository struct {
	*file.DB
}

// NewRepository возвращает новый репозиторий
func NewRepository(db *file.DB) *Repository {
	return &Repository{
		DB: db,
	}
}
