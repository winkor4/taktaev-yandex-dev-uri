package model

import (
	"context"
	"crypto/md5"
	"encoding/hex"
)

type URLRepository interface {
	GetURL(key string) (*URL, error)
	SaveURL(urls []URL) error
	PingDB(ctx context.Context) error
	GetUsersURL(user string) ([]KeyAndOURL, error)
}

type URL struct {
	OriginalURL   string `json:"original_url"`
	Key           string
	CorrelationID string `json:"correlation_id"`
	Conflict      bool
	UserID        string
}

type KeyAndOURL struct {
	Key         string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func ShortKey(ourl string) string {
	hash := md5.Sum([]byte(ourl))
	return hex.EncodeToString(hash[:])
}
