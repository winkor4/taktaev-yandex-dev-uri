package memory

import (
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/model"
	"errors"
)

type DB struct {
	dbMap map[string]string
}

func New() *DB {
	return &DB{
		dbMap: make(map[string]string),
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
}

func (db *DB) Close() error {
	return nil
}
