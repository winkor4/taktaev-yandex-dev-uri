package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/dbsql"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/models"
)

type StorageJSStruct struct {
	UUID          int    `json:"uuid"`
	ShortKey      string `json:"short_key"`
	OriginalURL   string `json:"original_url"`
	CorrelationID string `json:"correlation_id"`
}

type StorageJS struct {
	table   []StorageJSStruct
	encoder *json.Encoder
	file    *os.File
}

type URLData struct {
	CorrelationID string
	OriginalURL   string
}

type StorageMap struct {
	m   map[string]URLData
	sjs StorageJS
	DB  dbsql.PSQLDB
}

func NewStorageMap(fname string) (*StorageMap, error) {
	err := os.MkdirAll(filepath.Dir(fname), 0666)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	sjs := StorageJS{
		table:   make([]StorageJSStruct, 0),
		encoder: json.NewEncoder(file),
		file:    file,
	}

	sm := StorageMap{
		m:   make(map[string]URLData),
		sjs: sjs,
	}

	if err := readStorageFile(sm, fname); err != nil {
		return nil, err
	}

	return &sm, nil
}

func (s *StorageMap) CloseStorageFile() error {
	return s.sjs.file.Close()
}

func (s *StorageMap) GetURL(key string) (string, error) {
	if s.DB.NotAvailable() {
		return s.m[key].OriginalURL, nil
	}
	ourl, err := s.DB.SelectURL(key)
	if err != nil {
		return "", err
	}
	return ourl, nil
}

func (s *StorageMap) PostURL(key string, ourl string) error {
	_, ok := s.m[key]
	if ok {
		return nil
	}
	err := s.DB.Insert(key, ourl)
	if err != nil {
		return err
	}
	s.m[key] = URLData{
		OriginalURL: ourl,
	}
	uuid := len(s.sjs.table) + 1
	js := StorageJSStruct{
		UUID:        uuid,
		ShortKey:    key,
		OriginalURL: ourl,
	}
	s.sjs.table = append(s.sjs.table, js)
	return json.NewEncoder(s.sjs.file).Encode(&js)
}

func (s *StorageMap) PostBatch(obj []models.ShortenBatchRequest) error {
	dataToWrite := make([]models.ShortenBatchRequest, 0)
	for _, data := range obj {
		_, ok := s.m[data.ShortKey]
		if ok {
			continue
		}
		dataToWrite = append(dataToWrite, data)
		s.m[data.ShortKey] = URLData{
			OriginalURL:   data.OriginalURL,
			CorrelationID: data.CorrelationID,
		}
		uuid := len(s.sjs.table) + 1
		js := StorageJSStruct{
			UUID:          uuid,
			ShortKey:      data.ShortKey,
			OriginalURL:   data.OriginalURL,
			CorrelationID: data.CorrelationID,
		}
		s.sjs.table = append(s.sjs.table, js)
		if err := json.NewEncoder(s.sjs.file).Encode(&js); err != nil {
			return err
		}
	}
	if len(dataToWrite) == 0 {
		return nil
	}
	err := s.DB.InsertBatch(dataToWrite)
	if err != nil {
		return err
	}
	return nil
}

func readStorageFile(sm StorageMap, fname string) error {
	var js StorageJSStruct
	strData, err := os.ReadFile(fname)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	sliceData := strings.Split(string(strData), "\n")
	for _, data := range sliceData {
		if data == "" {
			continue
		}
		if err := json.Unmarshal([]byte(data), &js); err != nil {
			return err
		}
		sm.sjs.table = append(sm.sjs.table, js)
		sm.m[js.ShortKey] = URLData{
			OriginalURL:   js.OriginalURL,
			CorrelationID: js.CorrelationID,
		}
	}
	return nil
}
