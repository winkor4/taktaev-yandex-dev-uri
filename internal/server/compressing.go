package server

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// gzipResponseWriter описывает ResponseWriter с использованием gzip
type gzipResponseWriter struct {
	http.ResponseWriter
	gzipW *gzip.Writer
}

// newGzipResponseWriter возвращает новый gzipResponseWriter
func newGzipResponseWriter(w http.ResponseWriter) *gzipResponseWriter {
	return &gzipResponseWriter{
		ResponseWriter: w,
		gzipW:          gzip.NewWriter(w),
	}
}

// Header команда соответствия интерфейсу
func (gw *gzipResponseWriter) Header() http.Header {
	return gw.ResponseWriter.Header()
}

// Write команда соответствия интерфейсу
func (gw *gzipResponseWriter) Write(p []byte) (int, error) {
	return gw.gzipW.Write(p)
}

// WriteHeader команда соответствия интерфейсу
func (gw *gzipResponseWriter) WriteHeader(statusCode int) {
	if statusCode < 300 || statusCode == http.StatusConflict {
		gw.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	}
	gw.ResponseWriter.WriteHeader(statusCode)
}

// Close команда соответствия интерфейсу
func (gw *gzipResponseWriter) Close() error {
	return gw.gzipW.Close()
}

// gzipReader описывает ReadCloser для gzip
type gzipReader struct {
	ioR   io.ReadCloser
	gzipR *gzip.Reader
}

// newGzipReader возвращает новый gzipReader
func newGzipReader(ioR io.ReadCloser) (*gzipReader, error) {
	gzipR, err := gzip.NewReader(ioR)
	if err != nil {
		return nil, err
	}

	return &gzipReader{
		ioR:   ioR,
		gzipR: gzipR,
	}, nil
}

// Read команда соответствия интерфейсу
func (gzipR *gzipReader) Read(p []byte) (n int, err error) {
	return gzipR.gzipR.Read(p)
}

// Close команда соответствия интерфейсу
func (gzipR *gzipReader) Close() error {
	if err := gzipR.ioR.Close(); err != nil {
		return err
	}
	return gzipR.gzipR.Close()
}

// gzipMiddleware обработчик сжатия/распаковки данных
func gzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := w

		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gw := newGzipResponseWriter(w)
			rw = gw
			defer func() {
				_ = gw.Close()
			}()
		}

		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzipR, err := newGzipReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = gzipR
			defer func() {
				_ = gzipR.Close()
			}()
		}

		h.ServeHTTP(rw, r)
	})
}
