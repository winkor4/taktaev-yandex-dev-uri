package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/adapter/file"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/adapter/memory"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/adapter/psql"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/log"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/pkg/config"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/server"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/storage"
	sfile "github.com/winkor4/taktaev-yandex-dev-uri.git/internal/storage/file"
	smemory "github.com/winkor4/taktaev-yandex-dev-uri.git/internal/storage/memory"
	spsql "github.com/winkor4/taktaev-yandex-dev-uri.git/internal/storage/psql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApp(t *testing.T) {
	storages := make([]string, 0, 3)
	// storages = append(storages, "")
	// storages = append(storages, "file")
	storages = append(storages, "dsn")

	for _, dbName := range storages {
		srv := newTestServer(t, dbName)
		if srv == nil {
			continue
		}
		testAPI(t, srv, dbName)
		testAPIBatch(t, srv, dbName)
		testAPIDelete(t, srv, dbName)
		srv.Close()
	}
}

func newTestServer(t *testing.T, dbName string) *httptest.Server {
	cfg, err := config.Parse()
	require.NoError(t, err)

	var repo storage.Repository

	switch {
	case dbName == "dsn":
		if cfg.DSN == "" {
			return nil
		}
		db, err := spsql.New(cfg.DSN)
		require.NoError(t, err)
		sqlRepo := psql.NewRepository(db)
		err = sqlRepo.DeleteTable()
		require.NoError(t, err)
		repo = sqlRepo
	case dbName == "file":
		err := os.Remove("tmp/short-url-db-test.json")
		require.NoError(t, err)
		db, err := sfile.New("tmp/short-url-db-test.json")
		require.NoError(t, err)
		repo = file.NewRepository(db)
	default:
		db := smemory.New()
		repo = memory.NewRepository(db)
	}

	logger, err := log.New()
	require.NoError(t, err)

	srv := server.New(server.Config{
		URLRepo: repo,
		Cfg:     cfg,
		Logger:  logger,
	})

	return httptest.NewServer(server.SrvRouter(srv))
}

func testAPI(t *testing.T, srv *httptest.Server, dbName string) {

	type reqData struct {
		shortenedURL string
		originalURL  string
		key          string
	}

	cache := make(map[string]reqData, 0)

	type want struct {
		contentType  string
		statusCode   int
		originalURL  string
		key          string
		shortenedURL string
		body         []byte
	}

	type testData struct {
		name              string
		id                string
		postID            string
		method            string
		path              string
		contentType       string
		body              []byte
		originalURL       string
		withAuthorization bool
		checkRedirect     bool
		checkContentType  bool
		checkBody         bool
		want              want
	}

	type (
		urlSchema struct {
			URL string `json:"url"`
		}

		responseSchema struct {
			Result string `json:"result"`
		}

		responseUserURLS struct {
			ShortURL    string `json:"short_url"`
			OriginalURL string `json:"original_url"`
		}
	)

	resUserURLS := []responseUserURLS{
		{
			ShortURL:    "http://localhost:8080/d245406cb6c9f36be9064c92c34e12e1",
			OriginalURL: "https://www.youtube.com",
		},
		{
			ShortURL:    "http://localhost:8080/05db7cd93a2ace1c51bc40e23a1fab87",
			OriginalURL: "https://www.youtube.com/watch?v=etAIpkdhU9Q&list=RD09839DpTctU&index=30",
		},
	}

	resBody, err := json.Marshal(resUserURLS)
	require.NoError(t, err)

	testURL, err := json.Marshal(urlSchema{
		URL: "https://www.youtube.com/watch?v=etAIpkdhU9Q&list=RD09839DpTctU&index=30",
	})
	require.NoError(t, err)

	pingStatus := make(map[string]int)
	pingStatus[""] = http.StatusInternalServerError
	pingStatus["file"] = http.StatusInternalServerError
	pingStatus["dsn"] = http.StatusOK

	testTable := []testData{
		{
			name:              dbName + " Выполнить Post запрос /",
			id:                "post /",
			postID:            "",
			method:            http.MethodPost,
			path:              "/",
			contentType:       "text/plain",
			body:              []byte("https://www.youtube.com"),
			originalURL:       "https://www.youtube.com",
			withAuthorization: true,
			checkRedirect:     false,
			checkContentType:  true,
			checkBody:         false,
			want: want{
				contentType:  "text/plain",
				statusCode:   http.StatusCreated,
				originalURL:  "",
				key:          "d245406cb6c9f36be9064c92c34e12e1",
				shortenedURL: "http://localhost:8080/d245406cb6c9f36be9064c92c34e12e1",
				body:         []byte(""),
			},
		},
		{
			name:              dbName + " Выполнить повторно Post запрос /",
			id:                "second post /",
			postID:            "",
			method:            http.MethodPost,
			path:              "/",
			contentType:       "text/plain",
			body:              []byte("https://www.youtube.com"),
			originalURL:       "https://www.youtube.com",
			withAuthorization: false,
			checkRedirect:     false,
			checkContentType:  true,
			checkBody:         false,
			want: want{
				contentType:  "text/plain",
				statusCode:   http.StatusConflict,
				originalURL:  "",
				key:          "d245406cb6c9f36be9064c92c34e12e1",
				shortenedURL: "http://localhost:8080/d245406cb6c9f36be9064c92c34e12e1",
				body:         []byte(""),
			},
		},
		{
			name:              dbName + " Выполнить Get запрос /{id}",
			id:                "get /{id}",
			postID:            "post /",
			method:            http.MethodGet,
			path:              "/",
			contentType:       "",
			body:              []byte(""),
			originalURL:       "",
			withAuthorization: false,
			checkRedirect:     true,
			checkContentType:  false,
			checkBody:         false,
			want: want{
				contentType:  "",
				statusCode:   http.StatusTemporaryRedirect,
				originalURL:  "https://www.youtube.com",
				key:          "",
				shortenedURL: "",
				body:         []byte(""),
			},
		},
		{
			name:              dbName + " Выполнить Get запрос /ping",
			id:                "get /ping",
			postID:            "",
			method:            http.MethodGet,
			path:              "/ping",
			contentType:       "",
			body:              []byte(""),
			originalURL:       "",
			withAuthorization: false,
			checkRedirect:     false,
			checkContentType:  false,
			checkBody:         false,
			want: want{
				contentType:  "",
				statusCode:   pingStatus[dbName],
				originalURL:  "",
				key:          "",
				shortenedURL: "",
				body:         []byte(""),
			},
		},
		{
			name:              dbName + " Выполнить Post запрос /api/shorten",
			id:                "post /api/shorten",
			postID:            "",
			method:            http.MethodPost,
			path:              "/api/shorten",
			contentType:       "application/json",
			body:              testURL,
			originalURL:       "https://www.youtube.com/watch?v=etAIpkdhU9Q&list=RD09839DpTctU&index=30",
			withAuthorization: true,
			checkRedirect:     false,
			checkContentType:  true,
			checkBody:         false,
			want: want{
				contentType:  "application/json",
				statusCode:   http.StatusCreated,
				originalURL:  "",
				key:          "05db7cd93a2ace1c51bc40e23a1fab87",
				shortenedURL: "http://localhost:8080/05db7cd93a2ace1c51bc40e23a1fab87",
				body:         []byte(""),
			},
		},
		{
			name:              dbName + " Выполнить повторно Post запрос /api/shorten",
			id:                "second post /api/shorten",
			postID:            "",
			method:            http.MethodPost,
			path:              "/api/shorten",
			contentType:       "application/json",
			body:              testURL,
			originalURL:       "https://www.youtube.com/watch?v=etAIpkdhU9Q&list=RD09839DpTctU&index=30",
			withAuthorization: false,
			checkRedirect:     false,
			checkContentType:  true,
			checkBody:         false,
			want: want{
				contentType:  "application/json",
				statusCode:   http.StatusConflict,
				originalURL:  "",
				key:          "05db7cd93a2ace1c51bc40e23a1fab87",
				shortenedURL: "http://localhost:8080/05db7cd93a2ace1c51bc40e23a1fab87",
				body:         []byte(""),
			},
		},
		{
			name:              dbName + " Выполнить Get запрос /api/user/urls",
			id:                "get /api/user/urls",
			postID:            "",
			method:            http.MethodGet,
			path:              "/api/user/urls",
			contentType:       "",
			body:              []byte(""),
			originalURL:       "",
			withAuthorization: true,
			checkRedirect:     false,
			checkContentType:  true,
			checkBody:         true,
			want: want{
				contentType:  "application/json",
				statusCode:   http.StatusOK,
				originalURL:  "",
				key:          "",
				shortenedURL: "",
				body:         resBody,
			},
		},
		{
			name:              dbName + " Выполнить Get запрос без авторизации /api/user/urls",
			id:                "get unauth /api/user/urls",
			postID:            "",
			method:            http.MethodGet,
			path:              "/api/user/urls",
			contentType:       "",
			body:              []byte(""),
			originalURL:       "",
			withAuthorization: false,
			checkRedirect:     false,
			checkContentType:  false,
			checkBody:         false,
			want: want{
				contentType:  "application/json",
				statusCode:   http.StatusUnauthorized,
				originalURL:  "",
				key:          "",
				shortenedURL: "",
				body:         []byte(""),
			},
		},
	}

	var user string

	for _, testData := range testTable {
		t.Run(testData.name, func(t *testing.T) {

			if testData.postID != "" {
				testData.path = testData.path + cache[testData.postID].key
			}

			body := bytes.NewReader(testData.body)
			request, err := http.NewRequest(testData.method, srv.URL+testData.path, body)
			require.NoError(t, err)
			request.Header.Set("Content-Type", testData.contentType)

			client := srv.Client()
			if testData.checkRedirect {
				client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				}
			}

			if user != "" && testData.withAuthorization {
				request.AddCookie(&http.Cookie{
					Name:  "auth_token",
					Value: user,
				})
			}

			r, err := client.Do(request)
			require.NoError(t, err)

			if testData.withAuthorization {
				for _, c := range r.Cookies() {
					if c.Name == "auth_token" {
						user = c.Value
					}
				}
			}

			assert.Equal(t, testData.want.statusCode, r.StatusCode)
			if testData.checkContentType {
				assert.Equal(t, testData.want.contentType, r.Header.Get("Content-Type"))
			}
			if testData.checkRedirect {
				assert.Equal(t, testData.want.originalURL, r.Header.Get("Location"))
			}

			if testData.method == http.MethodGet && !testData.checkBody {
				return
			}

			rBody, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			err = r.Body.Close()
			require.NoError(t, err)

			var shortenedURL string
			switch {
			case testData.checkBody:
				assert.JSONEq(t, string(testData.want.body), string(rBody))
				return
			case testData.contentType == "text/plain":
				shortenedURL = string(rBody)
			case testData.contentType == "application/json":
				var schema responseSchema
				err = json.Unmarshal(rBody, &schema)
				require.NoError(t, err)
				shortenedURL = schema.Result
			}

			require.NotEmpty(t, shortenedURL)
			key := strings.ReplaceAll(shortenedURL, "http://localhost:8080/", "")

			assert.Equal(t, testData.want.shortenedURL, shortenedURL)
			assert.Equal(t, testData.want.key, key)

			require.NotEmpty(t, key)
			cache[testData.id] = reqData{
				shortenedURL: shortenedURL,
				originalURL:  testData.originalURL,
				key:          key,
			}

		})
	}
}

func testAPIBatch(t *testing.T, srv *httptest.Server, dbName string) {

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
	require.NoError(t, err)

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
	require.NoError(t, err)

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
		t.Run(testData.name, func(t *testing.T) {

			body := bytes.NewReader(testData.body)
			request, err := http.NewRequest(http.MethodPost, srv.URL+"/api/shorten/batch", body)
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")

			client := srv.Client()
			r, err := client.Do(request)
			require.NoError(t, err)

			assert.Equal(t, testData.want.statusCode, r.StatusCode)
			assert.Equal(t, testData.want.contentType, r.Header.Get("Content-Type"))

			rBody, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			err = r.Body.Close()
			require.NoError(t, err)

			require.NotEmpty(t, rBody)
			assert.JSONEq(t, string(testData.want.response), string(rBody))

		})
	}
}

func testAPIDelete(t *testing.T, srv *httptest.Server, dbName string) {

	type want struct {
		statusCode int
	}

	type testData struct {
		name  string
		bodys [][]byte
		want  want
	}

	testBodys := make([][]byte, 0, 10)

	for i := 0; i < 10; i++ {
		testBodys = append(testBodys, []byte(fmt.Sprintf("https://www.youtube.com/%d", i)))
	}

	testTable := []testData{
		{
			name:  dbName + " Выполнить Delete /api/user/urls",
			bodys: testBodys,
			want: want{
				statusCode: http.StatusAccepted,
			},
		},
	}

	var user string

	for _, testData := range testTable {
		t.Run(testData.name, func(t *testing.T) {

			shortenURLS := make([]string, 0, 10)

			for _, tbody := range testData.bodys {
				body := bytes.NewReader(tbody)
				request, err := http.NewRequest(http.MethodPost, srv.URL+"/", body)
				require.NoError(t, err)
				request.Header.Set("Content-Type", "text/plain")

				if user != "" {
					request.AddCookie(&http.Cookie{
						Name:  "auth_token",
						Value: user,
					})
				}

				client := srv.Client()
				r, err := client.Do(request)
				require.NoError(t, err)
				assert.Equal(t, http.StatusCreated, r.StatusCode)

				if user == "" {
					for _, c := range r.Cookies() {
						if c.Name == "auth_token" {
							user = c.Value
						}
					}
				}

				rBody, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				err = r.Body.Close()
				require.NoError(t, err)

				shortenedURL := string(rBody)
				key := strings.ReplaceAll(shortenedURL, "http://localhost:8080/", "")
				shortenURLS = append(shortenURLS, key)
			}

			dBody, err := json.Marshal(shortenURLS)
			require.NoError(t, err)

			body := bytes.NewReader(dBody)
			request, err := http.NewRequest(http.MethodDelete, srv.URL+"/api/user/urls", body)
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(&http.Cookie{
				Name:  "auth_token",
				Value: user,
			})

			client := srv.Client()
			r, err := client.Do(request)
			require.NoError(t, err)
			err = r.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, testData.want.statusCode, r.StatusCode)

			time.Sleep(time.Second * 1)

			request, err = http.NewRequest(http.MethodGet, srv.URL+"/"+shortenURLS[0], nil)
			require.NoError(t, err)

			client = srv.Client()
			client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}

			r, err = client.Do(request)
			require.NoError(t, err)
			err = r.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, http.StatusGone, r.StatusCode)

		})
	}
}
