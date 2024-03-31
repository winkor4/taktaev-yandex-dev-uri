package memory

import (
	"errors"

	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/model"
)

type DB struct {
	dbMap    map[string]string
	usersMap map[string][]model.KeyAndOURL
}

func New() *DB {
	return &DB{
		dbMap:    make(map[string]string),
		usersMap: make(map[string][]model.KeyAndOURL, 0),
	}
}

func (db *DB) Get(key string) (string, error) {
	ourl, ok := db.dbMap[key]
	if !ok {
		return "", errors.New("not found")
	}
	return ourl, nil
}

func (db *DB) Set(url *model.URL) {
	_, ok := db.dbMap[url.Key]
	if ok {
		url.Conflict = true
		return
	}
	db.dbMap[url.Key] = url.OriginalURL

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

func (db *DB) Close() error {
	return nil
}

func (db *DB) GetByUser(user string) []model.KeyAndOURL {
	usersURLS := db.usersMap[user]
	return usersURLS
}
