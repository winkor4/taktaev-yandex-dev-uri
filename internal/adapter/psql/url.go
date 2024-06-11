package psql

import (
	"context"

	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/model"
)

// GetURL возвращает ссылку по ключу
func (r *Repository) GetURL(key string) (*model.URL, error) {
	ourl, err := r.Get(key)
	if err != nil {
		return nil, err
	}

	out := new(model.URL)
	out.OriginalURL = ourl
	out.Key = key

	return out, nil
}

// SaveURL созраняет ссылку в бд
func (r *Repository) SaveURL(urls []model.URL) error {
	return r.Set(urls)
}

// PingDB проверяет соединение с бд
func (r *Repository) PingDB(ctx context.Context) error {
	return r.Ping(ctx)
}

// GetUsersURL возвращает все ссылки пользователя
func (r *Repository) GetUsersURL(user string) ([]model.KeyAndOURL, error) {
	return r.GetByUser(user)
}

// DeleteURL удаляет ссылку из бд
func (r *Repository) DeleteURL(user string, keys []string) {
	r.UpdateDeleteFlag(user, keys)
}
