package server

import (
	"net/http"
	"time"
)

// responseData дополнительные поля для http.ResponseWriter
type responseData struct {
	status int
	size   int
}

// loggingResponseWriter описывает http.ResponseWriter с доп полями
type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

// Write команда соответствия интерфейсу
func (lw *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := lw.ResponseWriter.Write(b)
	lw.responseData.size += size
	return size, err
}

// WriteHeader команда соответствия интерфейсу
func (lw *loggingResponseWriter) WriteHeader(statusCode int) {
	lw.ResponseWriter.WriteHeader(statusCode)
	lw.responseData.status = statusCode
}

// newLoggingResponseWriter возвращает новый loggingResponseWriter
func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	responseData := new(responseData)
	lw := new(loggingResponseWriter)
	lw.ResponseWriter = w
	lw.responseData = responseData
	return lw
}

// logMiddleware обработчик логирования при запросе
func logMiddleware(s *Server) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			logW := newLoggingResponseWriter(w)
			h.ServeHTTP(logW, r)

			duration := time.Since(start)

			s.logger.Infoln(
				"uri", r.RequestURI,
				"method", r.Method,
				"status", logW.responseData.status,
				"duration", duration,
				"size", logW.responseData.size,
			)
		})
	}
}
