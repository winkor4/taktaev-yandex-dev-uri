package handlers

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/config"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/storage"
	"go.uber.org/zap"
)

type (
	HandlerData struct {
		SM  *storage.StorageMap
		Cfg *config.Config
		L   *zap.SugaredLogger
	}
	responseData struct {
		status int
		size   int
	}
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (hd *HandlerData) URLRouter() chi.Router {
	r := chi.NewRouter()
	r.Post("/", hd.WithLogging(hd.shortURL))
	r.Get("/{id}", hd.WithLogging(hd.getURL))
	return r
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

// WithLogging добавляет дополнительный код для регистрации сведений о запросе
// и возвращает новый http.Handler.
func (hd *HandlerData) WithLogging(h http.HandlerFunc) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}
		h(&lw, r) // внедряем реализацию http.ResponseWriter

		duration := time.Since(start)

		hd.L.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status, // получаем перехваченный код статуса ответа
			"duration", duration,
			"size", responseData.size, // получаем перехваченный размер ответа
		)
	}
	return http.HandlerFunc(logFn)
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
	http.Redirect(res, req, originalURL, http.StatusTemporaryRedirect)
}
