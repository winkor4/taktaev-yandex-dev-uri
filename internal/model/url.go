// Модуль model содержит общие типы, которые использует сервер.
package model

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
)

// ErrIsDeleted - ошибка "URL удален".
var ErrIsDeleted = errors.New("url is deleted")

// URLRepository интерфейс для хранения данных.
type URLRepository interface {
	GetURL(key string) (*URL, error)
	SaveURL(urls []URL) error
	PingDB(ctx context.Context) error
	GetUsersURL(user string) ([]KeyAndOURL, error)
	DeleteURL(user string, keys []string)
}

// URL - описание входящих ссылок.
type URL struct {
	OriginalURL   string `json:"original_url"`
	Key           string
	CorrelationID string `json:"correlation_id"`
	Conflict      bool
	UserID        string
}

// KeyAndOURL - описание хранения ссылок на сервере.
type KeyAndOURL struct {
	Key         string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// ShortKey сокращает ссылку и возвращает ключ.
func ShortKey(ourl string) string {
	var bb bytes.Buffer
	bb.Grow(len(ourl))
	bb.WriteString(ourl)

	hash := md5.Sum(bb.Bytes())
	// hash := md5.Sum([]byte(ourl))
	return hex.EncodeToString(hash[:])
}
