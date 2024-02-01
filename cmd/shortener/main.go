package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
)

var urlsId = make(map[string]string)

func generateShortKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 8

	shortKey := make([]byte, keyLength)
	for i := range shortKey {
		shortKey[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortKey)
}

func shortUrl(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Invalid request method", http.StatusBadRequest)
		return
	}
	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		http.Error(res, "Missing Header: Content-Type", http.StatusBadRequest)
		return
	}
	if contentType != "text/plain" {
		http.Error(res, "Header: Content-Type must be text/plain", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Cant read body", http.StatusBadRequest)
		return
	}
	originalUrl := string(body)
	if originalUrl == "" {
		http.Error(res, "URL parameter is missing", http.StatusBadRequest)
		return
	}
	shortKey := generateShortKey()
	urlsId[shortKey] = originalUrl
	shortenedURL := fmt.Sprintf("http://localhost:8080/%s", shortKey)
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	data := []byte(shortenedURL)
	res.Write(data)
}

func getUrl(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Invalid request method", http.StatusBadRequest)
		return
	}
	shortKey := req.RequestURI[1:]
	originalUrl := urlsId[shortKey]
	if originalUrl == "" {
		http.Error(res, "Invalid url key", http.StatusBadRequest)
		return
	}
	res.Header().Set("Location", originalUrl)
	res.WriteHeader(http.StatusTemporaryRedirect)

}

func rootHandle(res http.ResponseWriter, req *http.Request) {
	if req.RequestURI == "/" {
		shortUrl(res, req)
	} else {
		getUrl(res, req)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, rootHandle)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
