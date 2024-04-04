package memory

import (
	"context"
	"errors"

	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/model"
)

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

func (r *Repository) SaveURL(urls []model.URL) error {
	for i := range urls {
		r.Set(&urls[i])
	}
	return nil
}

func (r *Repository) PingDB(ctx context.Context) error {
	return errors.New("connection could't be established")
}

func (r *Repository) GetUsersURL(user string) ([]model.KeyAndOURL, error) {
	return r.GetByUser(user), nil
}

func (r *Repository) DeleteURL(user string, keys []string) {
	r.UpdateDeleteFlag(user, keys)
}
