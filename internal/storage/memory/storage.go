package memory

import (
	"errors"
	"slices"

	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/model"
)

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

func New() *DB {
	return &DB{
		dbMap:    make(map[string]memoryURL),
		usersMap: make(map[string][]model.KeyAndOURL, 0),
	}
}

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

func (db *DB) Close() error {
	return nil
}

func (db *DB) GetByUser(user string) []model.KeyAndOURL {
	usersURLS := db.usersMap[user]
	return usersURLS
}

func (db *DB) UpdateDeleteFlag(user string, keys []string) {
	userURLS := db.usersMap[user]

	for _, key := range keys {
		url := db.dbMap[key]
		if url.UserID != user {
			continue
		}
		url.IsDeleted = true
		db.dbMap[key] = url

		idx := slices.IndexFunc(userURLS, func(v model.KeyAndOURL) bool { return v.Key == key })
		if idx >= 0 {
			userURLS[idx] = userURLS[len(userURLS)-1]
			userURLS = userURLS[:len(userURLS)-1]
		}
	}

	db.usersMap[user] = userURLS
}
