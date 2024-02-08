package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

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

func shortURL(res http.ResponseWriter, req *http.Request) {
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
	UrlsID[shortKey] = originalURL
	shortenedURL := fmt.Sprintf("http://localhost:8080/%s", shortKey)
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	data := []byte(shortenedURL)
	res.Write(data)
}

func getURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Invalid request method", http.StatusBadRequest)
		return
	}
	shortKey := chi.URLParam(req, "id")
	// shortKey := req.RequestURI[1:]
	originalURL := UrlsID[shortKey]
	if originalURL == "" {
		http.Error(res, "Invalid url key", http.StatusBadRequest)
		return
	}
	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)

}

func URLRouter() chi.Router {
	r := chi.NewRouter()
	r.Post("/", shortURL)
	r.Get("/{id}", getURL)
	return r
}
