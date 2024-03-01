package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/compression"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/config"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/models"
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
	r.Post("/", hd.gzipMiddleware(hd.WithLogging(hd.shortURL)))
	r.Get("/{id}", hd.gzipMiddleware(hd.WithLogging(hd.getURL)))
	r.Route("/api", func(r chi.Router) {
		r.Post("/shorten", hd.gzipMiddleware(hd.WithLogging(hd.shortURLJS)))
	})
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

func (hd *HandlerData) gzipMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// по умолчанию устанавливаем оригинальный http.ResponseWriter как тот,
		// который будем передавать следующей функции
		gzipW := w

		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			cw := compression.NewCompressWriter(w)
			// меняем оригинальный http.ResponseWriter на новый
			gzipW = cw
			// не забываем отправить клиенту все сжатые данные после завершения middleware
			defer cw.Close()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			gzipR, err := compression.NewCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// меняем тело запроса на новое
			r.Body = gzipR
			defer gzipR.Close()
		}

		h(gzipW, r)
	}
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

func generateShortKey(originalURL string) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 8

	shortKey := make([]byte, keyLength)
	for i := range shortKey {
		shortKey[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortKey)

	// hash := md5.Sum([]byte(originalURL))
	// return hex.EncodeToString(hash[:])
}

func badContentType(contentType string, expType string) bool {
	if contentType == "" {
		return true
	}
	out := true
	for _, v := range strings.Split(contentType, ";") {
		if v == expType || v == "application/x-gzip" {
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
	if badContentType(contentType, "text/plain") {
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
	shortKey := generateShortKey(originalURL)

	err = hd.SM.PostURL(shortKey, originalURL)
	if err != nil {
		http.Error(res, "Cant write data in file", http.StatusInternalServerError)
		return
	}
	shortenedURL := fmt.Sprintf(hd.Cfg.BaseURL+"/%s", shortKey)
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	data := []byte(shortenedURL)
	_, err = res.Write(data)
	if err != nil {
		http.Error(res, "Cant write response", http.StatusInternalServerError)
		return
	}
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

func (hd *HandlerData) shortURLJS(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Invalid request method", http.StatusBadRequest)
		return
	}
	contentType := req.Header.Get("Content-Type")
	if badContentType(contentType, "application/json") {
		http.Error(res, "Header: Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	var sreq models.ShortenRequest
	if err := json.NewDecoder(req.Body).Decode(&sreq); err != nil {
		http.Error(res, "Cant read body", http.StatusBadRequest)
		return
	}

	originalURL := sreq.URL
	if originalURL == "" {
		http.Error(res, "URL parameter is missing", http.StatusBadRequest)
		return
	}

	shortKey := generateShortKey(originalURL)
	err := hd.SM.PostURL(shortKey, originalURL)
	if err != nil {
		http.Error(res, "Cant write data in file", http.StatusInternalServerError)
		return
	}

	var sres models.ShortenResponse
	sres.Result = fmt.Sprintf(hd.Cfg.BaseURL+"/%s", shortKey)

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(res).Encode(sres); err != nil {
		http.Error(res, "error encoding response", http.StatusInternalServerError)
		return
	}
}
