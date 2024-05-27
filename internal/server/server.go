package server

import (
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/google/uuid"
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

type delURL struct {
	user string
	keys []string
}

type Server struct {
	urlRepo  model.URLRepository
	cfg      *config.Config
	logger   *log.Logger
	user     string
	deleteCh chan delURL
}

func New(c Config) *Server {
	deleteCh := make(chan delURL)
	return &Server{
		urlRepo:  c.URLRepo,
		cfg:      c.Cfg,
		logger:   c.Logger,
		deleteCh: deleteCh,
	}
}

func (s *Server) Run() error {
	go s.Workers()
	s.logger.Logw(s.cfg.LogLevel, "Starting server", "SrvAdr", s.cfg.SrvAdr)
	return http.ListenAndServe(s.cfg.SrvAdr, SrvRouter(s))
}

func (s *Server) Workers() {
	go delWorker(s)
}

func SrvRouter(s *Server) *chi.Mux {
	r := chi.NewRouter()
	r.Use(authorizationMiddleware(s) /*gzipMiddleware,*/, logMiddleware(s))

	r.Post("/", checkContentTypeMiddleware(shortURL(s), "text/plain"))
	r.Get("/{id}", getURL(s))
	r.Get("/ping", pingDB(s))
	r.Mount("/api", apiRouter(s))

	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return r
}

func apiRouter(s *Server) *chi.Mux {
	r := chi.NewRouter()
	r.Mount("/shorten", apiShortenRouter(s))
	r.Mount("/user", apiUserRouter(s))
	return r
}

func apiShortenRouter(s *Server) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", checkContentTypeMiddleware(shortURL(s), "application/json"))
	r.Post("/batch", checkContentTypeMiddleware(shortBatch(s), "application/json"))
	return r
}

func apiUserRouter(s *Server) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/urls", getUsersURL(s))
	r.Delete("/urls", checkContentTypeMiddleware(deleteURL(s), "application/json"))
	return r
}

func checkContentTypeMiddleware(h http.HandlerFunc, exContentType string) http.HandlerFunc {
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

func authorizationMiddleware(s *Server) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			user, err := parseUser(r, true)
			if err != nil {
				http.Error(w, "can't get cookie", http.StatusBadRequest)
			}

			http.SetCookie(w, &http.Cookie{
				Name:  "auth_token",
				Value: user,
			})

			s.user = user

			h.ServeHTTP(w, r)
		})
	}
}

func parseUser(r *http.Request, createNew bool) (string, error) {

	var user string

	auth, err := r.Cookie("auth_token")
	if err != nil && err != http.ErrNoCookie {
		return "", err
	}

	switch {
	case err == http.ErrNoCookie && createNew:
		user = uuid.New().String()
	case err == http.ErrNoCookie && !createNew:
		return "", err
	default:
		user = auth.Value
	}

	return user, nil
}
