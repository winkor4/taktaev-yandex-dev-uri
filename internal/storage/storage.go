package storage

import (
	"encoding/json"
	"os"
	"strings"
)

type StorageJSStruct struct {
	Uuid         int
	Short_url    string
	Original_url string
}

type StorageJS struct {
	table   []StorageJSStruct
	encoder *json.Encoder
	file    *os.File
}

type StorageMap struct {
	m   map[string]string
	sjs StorageJS
}

func NewStorageMap(fname string) (*StorageMap, error) {
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
		m:   make(map[string]string),
		sjs: sjs,
	}

	if err := readStorageFile(sm, fname); err != nil {
		return nil, err
	}

	return &sm, err
}

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
		Uuid:         uuid,
		Short_url:    key,
		Original_url: ourl,
	}
	s.sjs.table = append(s.sjs.table, js)
	return s.sjs.encoder.Encode(&js)
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
		sm.m[js.Short_url] = js.Original_url
	}
	return nil
}
