package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"syscall"

	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/adapter/psql"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/log"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/pkg/config"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/storage"
	spsql "github.com/winkor4/taktaev-yandex-dev-uri.git/internal/storage/psql"
)

// Проверяем работу сервера
func Example() {
	srv, err := newTestServer()
	if err != nil {
		panic(err)
	}
	defer srv.Close()
	err = testAPIBatch(srv, "dsn")
	if err != nil {
		panic(err)
	}
}

// Создаем сервер для теста
func newTestServer() (*httptest.Server, error) {
	cfg, err := config.Parse()
	if err != nil {
		return nil, err
	}

	var repo storage.Repository

	db, err := spsql.New(cfg.DSN)
	if err != nil {
		return nil, err
	}
	sqlRepo := psql.NewRepository(db)
	err = sqlRepo.DeleteTable()
	if err != nil {
		return nil, err
	}
	repo = sqlRepo

	logger, err := log.New()
	if err != nil {
		return nil, err
	}

	srv := New(Config{
		URLRepo: repo,
		Cfg:     cfg,
		Logger:  logger,
	})

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	srv.Workers(sigs)

	return httptest.NewServer(SrvRouter(srv)), nil
}

// Проверяем АПИ
func testAPIBatch(srv *httptest.Server, dbName string) error {

	type want struct {
		contentType string
		statusCode  int
		response    []byte
	}

	type testData struct {
		name string
		body []byte
		want want
	}

	type (
		reqSchema struct {
			CorrelationID string `json:"correlation_id"`
			OriginalURL   string `json:"original_url"`
		}
		resSchema struct {
			CorrelationID string `json:"correlation_id"`
			ShortURL      string `json:"short_url"`
		}
	)

	reqData := []reqSchema{
		{
			CorrelationID: "1111",
			OriginalURL:   "https://www.youtube.com/watch?v=R6_3OchvW_c",
		},
		{
			CorrelationID: "2222",
			OriginalURL:   "https://tarkov-market.com/maps/lighthouse?quest=01a7662b-33c6-4236-991e-2f417a55ac0b",
		},
		{
			CorrelationID: "3333",
			OriginalURL:   "https://tarkov.help/ru/map/ground-zero/",
		},
	}

	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return err
	}

	resData := []resSchema{
		{
			CorrelationID: "1111",
			ShortURL:      "http://localhost:8080/f623e4d83928fb684b01a5972aba8346",
		},
		{
			CorrelationID: "2222",
			ShortURL:      "http://localhost:8080/f3ed5bff46a78b51cb66659582054012",
		},
		{
			CorrelationID: "3333",
			ShortURL:      "http://localhost:8080/6be5bdb4593aa12fed172b6764730a52",
		},
	}

	resBody, err := json.Marshal(resData)
	if err != nil {
		return err
	}

	testTable := []testData{
		{
			name: dbName + " Выполнить Post /api/shorten/bacth",
			body: reqBody,
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusCreated,
				response:    resBody,
			},
		},
	}

	for _, testData := range testTable {
		body := bytes.NewReader(testData.body)
		request, err := http.NewRequest(http.MethodPost, srv.URL+"/api/shorten/batch", body)
		if err != nil {
			return err
		}
		request.Header.Set("Content-Type", "application/json")

		client := srv.Client()
		r, err := client.Do(request)
		if err != nil {
			return err
		}

		rBody, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		err = r.Body.Close()
		if err != nil {
			return err
		}
		fmt.Println(string(rBody))

	}
	return nil
}
