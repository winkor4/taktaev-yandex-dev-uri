package server

import (
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/log"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/model"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/pkg/config"
	"net/http"

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
	return http.ListenAndServe(":8080", SrvRouter(s))
}

func SrvRouter(s *Server) *chi.Mux {
	r := chi.NewRouter()
	r.Use(gzipHandler, logHandler(s))

	r.Post("/", shortURL(s))
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
	r.Post("/", shortURL(s))
	r.Post("/batch", shortBatch(s))
	return r
}
