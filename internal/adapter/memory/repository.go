package memory

import "github.com/winkor4/taktaev-yandex-dev-uri.git/internal/storage/memory"

// Repository соответствует интерфйсу хранилища в оперативной памяти
type Repository struct {
	*memory.DB
}

// NewRepository возвращает новый репозиторий
func NewRepository(db *memory.DB) *Repository {
	return &Repository{
		DB: db,
	}
}
