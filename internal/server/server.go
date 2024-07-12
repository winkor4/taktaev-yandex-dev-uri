// Функции создания и запуска сервера.
package server

import (
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/google/uuid"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/log"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/model"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/pkg/config"
	"google.golang.org/grpc"

	// импортируем пакет со сгенерированными protobuf-файлами
	pb "github.com/winkor4/taktaev-yandex-dev-uri.git/proto"

	"github.com/go-chi/chi/v5"
)

// ServerGRPC поддерживает все необходимые методы gRPC сервера.
type ServerGRPC struct {
	pb.UnimplementedURLShortenerServer
	cfg      *config.Config
	user     string
	deleteCh chan delURL
	urlRepo  model.URLRepository
}

// Config хранит параметры для создания нового сервера.
type Config struct {
	URLRepo model.URLRepository // URLRepo - интерфейс хранилища.
	Cfg     *config.Config      // Cfg - параметры создания сервера.
	Logger  *log.Logger         // Logger - логгер сервера.
}

type delURL struct {
	user string
	keys []string
}

// Server содержит данные для запуска и работы сервера.
type Server struct {
	urlRepo  model.URLRepository
	cfg      *config.Config
	logger   *log.Logger
	user     string
	deleteCh chan delURL
}

// New создает и возвращает новый сервер.
func New(c Config) *Server {
	deleteCh := make(chan delURL)
	return &Server{
		urlRepo:  c.URLRepo,
		cfg:      c.Cfg,
		logger:   c.Logger,
		deleteCh: deleteCh,
	}
}

// Run запускает сервер.
func (s *Server) Run() error {
	var err error
	// Добавил sync.WaitGroup чтобы быть уверенным, что очередь с удалением будет очищена
	var wg sync.WaitGroup
	wg.Add(1)

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		sg := <-sigint
		println(sg)
		s.shutdown()
		wg.Wait()
		os.Exit(0)
	}()

	go runGRPC(s)

	s.Workers(&wg)
	s.logger.Logw(s.cfg.LogLevel, "Starting server", "SrvAdr", s.cfg.SrvAdr)
	if s.cfg.EnableHTTPS {
		err = http.ListenAndServeTLS(s.cfg.SrvAdr, "cert.pem", "key.pem", SrvRouter(s))
	} else {
		err = http.ListenAndServe(s.cfg.SrvAdr, SrvRouter(s))
	}

	return err
}

func runGRPC(s *Server) {
	// определяем порт для сервера
	listen, err := net.Listen("tcp", ":3200")
	if err != nil {
		return
	}

	serverGRPC := new(ServerGRPC)
	serverGRPC.cfg = s.cfg
	serverGRPC.deleteCh = s.deleteCh
	serverGRPC.urlRepo = s.urlRepo

	// создаём gRPC-сервер без зарегистрированной службы
	server := grpc.NewServer(grpc.UnaryInterceptor(serverGRPC.authorizationInterceptor))
	// регистрируем сервис
	pb.RegisterURLShortenerServer(server, serverGRPC)
	s.logger.Logw(s.cfg.LogLevel, "Сервер gRPC начал работу")
	// получаем запрос gRPC
	if err := server.Serve(listen); err != nil {
		return
	}
}

// Workers запускает фоновые обработчики.
func (s *Server) Workers(wg *sync.WaitGroup) {
	go delWorker(s, wg)
}

// SrvRouter возвращает описание (handler) сервера для запуска
func SrvRouter(s *Server) *chi.Mux {
	r := chi.NewRouter()
	r.Use(authorizationMiddleware(s), gzipMiddleware, logMiddleware(s))

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
	r.Get("/internal/stats", getStats(s))
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

func (s *Server) shutdown() {
	close(s.deleteCh)
}
