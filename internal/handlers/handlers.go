package handlers

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/config"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/storage"
)

type HandlerData struct {
	SM  *storage.StorageMap
	Cfg *config.Config
}

func (hd *HandlerData) URLRouter() chi.Router {
	r := chi.NewRouter()
	r.Post("/", hd.shortURL)
	r.Get("/{id}", hd.getURL)
	return r
}

func generateShortKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 8

	shortKey := make([]byte, keyLength)
	for i := range shortKey {
		shortKey[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortKey)
}

func invalidContentType(contentType string) bool {
	if contentType == "" {
		return true
	}
	out := true
	for _, v := range strings.Split(contentType, ";") {
		if v == "text/plain" {
			out = false
			break
		}
	}
	return out
}

func (hd *HandlerData) shortURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Invalid request method", http.StatusBadRequest)
		return
	}
	contentType := req.Header.Get("Content-Type")
	if invalidContentType(contentType) {
		http.Error(res, "Header: Content-Type must be text/plain", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Cant read body", http.StatusBadRequest)
		return
	}
	originalURL := string(body)
	if originalURL == "" {
		http.Error(res, "URL parameter is missing", http.StatusBadRequest)
		return
	}
	shortKey := generateShortKey()

	hd.SM.PostURL(shortKey, originalURL)
	shortenedURL := fmt.Sprintf(hd.Cfg.BaseURL+"/%s", shortKey)
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	data := []byte(shortenedURL)
	res.Write(data)
}

func (hd *HandlerData) getURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Invalid request method", http.StatusBadRequest)
		return
	}
	shortKey := chi.URLParam(req, "id")
	// shortKey := req.RequestURI[1:]
	originalURL := hd.SM.GetURL(shortKey)
	if originalURL == "" {
		http.Error(res, "Invalid url key", http.StatusBadRequest)
		return
	}
	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)

}
