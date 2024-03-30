package file

import (
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/model"
	"encoding/json"
	"errors"
	"os"
	"strings"
)

type DB struct {
	file *os.File
	data map[string]fileURL
}

type fileURL struct {
	UUID        int    `json:"uuid"`
	ShortKey    string `json:"short_key"`
	OriginalURL string `json:"original_url"`
}

func New(fname string) (*DB, error) {
	file, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	data, err := readStorageFile(fname)
	if err != nil {
		return nil, err
	}

	out := new(DB)
	out.file = file
	out.data = data

	return out, nil
}

func (db *DB) CloseFile() error {
	return db.file.Close()
}

func readStorageFile(fname string) (map[string]fileURL, error) {
	out := make(map[string]fileURL)

	strData, err := os.ReadFile(fname)
	if err != nil {
		return nil, err
	}

	var schema fileURL
	sliceData := strings.Split(string(strData), "\n")
	for _, data := range sliceData {
		if data == "" {
			continue
		}
		if err := json.Unmarshal([]byte(data), &schema); err != nil {
			return nil, err
		}
		out[schema.ShortKey] = schema
	}

	return out, nil
}

func (db *DB) Get(key string) (string, error) {
	fileData, ok := db.data[key]
	if !ok {
		return "", errors.New("not found")
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
		}

		err := json.NewEncoder(db.file).Encode(&fileURL)
		if err != nil {
			return err
		}
		db.data[fileURL.ShortKey] = fileURL
		uuid++
	}

	return nil
}
