package file

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"strings"

	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/model"
)

type DB struct {
	file     *os.File
	data     map[string]fileURL
	usersMap map[string][]model.KeyAndOURL
}

type fileURL struct {
	UUID        int    `json:"uuid"`
	ShortKey    string `json:"short_key"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
	IsDeleted   bool   `json:"is_deleted"`
}

func New(fname string) (*DB, error) {
	file, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	out := new(DB)
	out.file = file

	err = readStorageFile(out, fname)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (db *DB) CloseFile() error {
	return db.file.Close()
}

func readStorageFile(db *DB, fname string) error {
	fileData := make(map[string]fileURL)
	usersMap := make(map[string][]model.KeyAndOURL)

	strData, err := os.ReadFile(fname)
	if err != nil {
		return err
	}

	var schema fileURL
	sliceData := strings.Split(string(strData), "\n")
	for _, data := range sliceData {
		if data == "" {
			continue
		}
		if err := json.Unmarshal([]byte(data), &schema); err != nil {
			return err
		}
		fileData[schema.ShortKey] = schema

		userURLS := usersMap[schema.UserID]
		userURLS = append(userURLS, model.KeyAndOURL{
			Key:         schema.ShortKey,
			OriginalURL: schema.OriginalURL,
		})
		usersMap[schema.UserID] = userURLS
	}

	db.data = fileData
	db.usersMap = usersMap

	return nil
}

func (db *DB) Get(key string) (string, error) {
	fileData, ok := db.data[key]
	if !ok {
		return "", errors.New("not found")
	}
	if fileData.IsDeleted {
		return "", model.ErrIsDeleted
	}
	return fileData.OriginalURL, nil
}

func (db *DB) Set(urls []model.URL) error {

	uuid := len(db.data) + 1

	for i, url := range urls {
		_, ok := db.data[url.Key]
		if ok {
			urls[i].Conflict = true
			continue
		}

		fileURL := fileURL{
			UUID:        uuid,
			ShortKey:    url.Key,
			OriginalURL: url.OriginalURL,
			UserID:      url.UserID,
			IsDeleted:   false,
		}

		err := json.NewEncoder(db.file).Encode(&fileURL)
		if err != nil {
			return err
		}
		db.data[fileURL.ShortKey] = fileURL
		uuid++

		if url.UserID == "" {
			continue
		}
		userURLS := db.usersMap[url.UserID]
		userURLS = append(userURLS, model.KeyAndOURL{
			Key:         url.Key,
			OriginalURL: url.OriginalURL,
		})
		db.usersMap[url.UserID] = userURLS
	}

	return nil
}

func (db *DB) GetByUser(user string) []model.KeyAndOURL {
	usersURLS := db.usersMap[user]
	return usersURLS
}

func (db *DB) UpdateDeleteFlag(user string, keys []string) {
	userURLS := db.usersMap[user]

	for _, key := range keys {
		url := db.data[key]
		if url.UserID != user {
			continue
		}
		url.IsDeleted = true
		db.data[key] = url

		idx := slices.IndexFunc(userURLS, func(v model.KeyAndOURL) bool { return v.Key == key })
		if idx >= 0 {
			userURLS[idx] = userURLS[len(userURLS)-1]
			userURLS = userURLS[:len(userURLS)-1]
		}
	}

	db.usersMap[user] = userURLS

	fname := db.file.Name()

	err := db.file.Close()
	if err != nil {
		return
	}
	err = os.Remove(fname)
	if err != nil {
		return
	}

	file, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return
	}

	db.file = file

	for _, fileURL := range db.data {
		err := json.NewEncoder(db.file).Encode(&fileURL)
		if err != nil {
			return
		}
	}
}
