package psql

import (
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/storage/psql"
)

// Repository соответствует интерфйсу хранилища бд SQL
type Repository struct {
	*psql.DB
}

// NewRepository возвращает новый репозиторий
func NewRepository(db *psql.DB) *Repository {
	return &Repository{
		DB: db,
	}
}
