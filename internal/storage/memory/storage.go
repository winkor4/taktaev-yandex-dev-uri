// Модуль memory описывает функции хранения данных через оперативную память.
package memory

import (
	"errors"
	"slices"

	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/model"
)

// DB - описание хранилища.
type DB struct {
	dbMap    map[string]memoryURL
	usersMap map[string][]model.KeyAndOURL
}

type memoryURL struct {
	OriginalURL string
	ShortKey    string
	UserID      string
	IsDeleted   bool
}

// New возвращает новое хранилище (map).
func New() *DB {
	return &DB{
		dbMap:    make(map[string]memoryURL),
		usersMap: make(map[string][]model.KeyAndOURL, 0),
	}
}

// Get возвращает ссылку по ключу.
func (db *DB) Get(key string) (string, error) {
	ourl, ok := db.dbMap[key]
	if !ok {
		return "", errors.New("not found")
	}
	if ourl.IsDeleted {
		return "", model.ErrIsDeleted
	}
	return ourl.OriginalURL, nil
}

// Set записывает ссылки в файл.
func (db *DB) Set(url *model.URL) {
	_, ok := db.dbMap[url.Key]
	if ok {
		url.Conflict = true
		return
	}
	db.dbMap[url.Key] = memoryURL{
		OriginalURL: url.OriginalURL,
		ShortKey:    url.Key,
		UserID:      url.UserID,
		IsDeleted:   false,
	}

	if url.UserID == "" {
		return
	}
	userURLS := db.usersMap[url.UserID]
	userURLS = append(userURLS, model.KeyAndOURL{
		Key:         url.Key,
		OriginalURL: url.OriginalURL,
	})
	db.usersMap[url.UserID] = userURLS
}

// Close бланк для интерфейса.
func (db *DB) Close() error {
	return nil
}

// GetByUser возвращает все ссылки пользователя.
func (db *DB) GetByUser(user string) []model.KeyAndOURL {
	usersURLS := db.usersMap[user]
	return usersURLS
}

// UpdateDeleteFlag удаляет ссылки.
func (db *DB) UpdateDeleteFlag(user string, keys []string) {
	userURLS, ok := db.usersMap[user]

	for _, key := range keys {
		url := db.dbMap[key]
		if url.UserID != user && user != "" {
			continue
		}
		url.IsDeleted = true
		db.dbMap[key] = url

		if ok {
			idx := slices.IndexFunc(userURLS, func(v model.KeyAndOURL) bool { return v.Key == key })
			if idx >= 0 {
				userURLS[idx] = userURLS[len(userURLS)-1]
				userURLS = userURLS[:len(userURLS)-1]
			}
		}
	}

	switch {
	case user != "":
		db.usersMap[user] = userURLS
	default:
		db.usersMap = make(map[string][]model.KeyAndOURL, 0)
	}

}
