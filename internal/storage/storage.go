package storage

import (
	"encoding/json"
	"os"
	"strings"
)

type StorageJSStruct struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type StorageJS struct {
	table []StorageJSStruct
	fname string
	// encoder *json.Encoder
	// file    *os.File
}

type StorageMap struct {
	m   map[string]string
	sjs StorageJS
}

func NewStorageMap(fname string) (*StorageMap, error) {
	// file, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	// if err != nil {
	// 	return nil, err
	// }

	sjs := StorageJS{
		table: make([]StorageJSStruct, 0),
		fname: fname,
		// encoder: json.NewEncoder(file),
		// file:    file,
	}

	sm := StorageMap{
		m:   make(map[string]string),
		sjs: sjs,
	}

	if err := readStorageFile(sm, fname); err != nil {
		return nil, err
	}

	return &sm, nil
}

// func (s *StorageMap) CloseStorageFile() error {
// 	return s.sjs.file.Close()
// }

func (s *StorageMap) GetURL(key string) string {
	return s.m[key]
}

func (s *StorageMap) PostURL(key string, ourl string) error {
	_, ok := s.m[key]
	if ok {
		return nil
	}
	s.m[key] = ourl
	uuid := len(s.sjs.table) + 1
	js := StorageJSStruct{
		UUID:        uuid,
		ShortURL:    key,
		OriginalURL: ourl,
	}
	s.sjs.table = append(s.sjs.table, js)

	file, err := os.OpenFile(s.sjs.fname, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(&js)
}

func readStorageFile(sm StorageMap, fname string) error {
	var js StorageJSStruct
	strData, err := os.ReadFile(fname)
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
		sm.m[js.ShortURL] = js.OriginalURL
	}
	return nil
}
