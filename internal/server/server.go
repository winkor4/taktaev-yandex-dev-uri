package server

import (
	"net/http"
	"strings"

	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/log"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/model"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/pkg/config"

	"github.com/go-chi/chi/v5"
)

type Config struct {
	URLRepo model.URLRepository
	Cfg     *config.Config
	Logger  *log.Logger
}

type Server struct {
	urlRepo model.URLRepository
	cfg     *config.Config
	logger  *log.Logger
}

func New(c Config) *Server {
	return &Server{
		urlRepo: c.URLRepo,
		cfg:     c.Cfg,
		logger:  c.Logger,
	}
}

func (s *Server) Run() error {
	s.logger.Logw(s.cfg.LogLevel, "Starting server", "SrvAdr", s.cfg.SrvAdr)
	return http.ListenAndServe(s.cfg.SrvAdr, SrvRouter(s))
}

func SrvRouter(s *Server) *chi.Mux {
	r := chi.NewRouter()
	r.Use(gzipHandler, logHandler(s))

	r.Post("/", checkContentTypeHandler(shortURL(s), "text/plain"))
	r.Get("/{id}", getURL(s))
	r.Get("/ping", pingDB(s))
	r.Mount("/api", apiRouter(s))

	return r
}

func apiRouter(s *Server) *chi.Mux {
	r := chi.NewRouter()
	r.Mount("/shorten", apiShortenRouter(s))
	return r
}

func apiShortenRouter(s *Server) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", checkContentTypeHandler(shortURL(s), "application/json"))
	r.Post("/batch", checkContentTypeHandler(shortBatch(s), "application/json"))
	return r
}

func checkContentTypeHandler(h http.HandlerFunc, exContentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if strings.Contains(contentType, "application/x-gzip") {
			r.Header.Set("Content-Type", exContentType)
			h(w, r)
			return
		}

		if !strings.Contains(contentType, exContentType) {
			http.Error(w, "unexpected Content-Type", http.StatusBadRequest)
			return
		}
		h(w, r)
	}
}
