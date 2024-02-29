package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/config"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/handlers"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/logger"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/models"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/storage"
)

func TestURLRouter(t *testing.T) {
	hd := hd(t)
	ts := httptest.NewServer(hd.URLRouter())
	defer ts.Close()

	type want struct {
		contentType    string
		statusCodePost int
		statusCodeGet  int
	}
	testTable := []struct {
		name string
		url  string
		want want
	}{
		{
			name: "post and get url",
			url:  "https://reqbin.com/post-online",
			want: want{
				contentType:    "text/plain",
				statusCodePost: http.StatusCreated,
				statusCodeGet:  http.StatusTemporaryRedirect,
			},
		},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(tt.url)
			request, err := http.NewRequest(http.MethodPost, ts.URL+"/", body)
			require.NoError(t, err)
			request.Header.Set("Content-Type", "text/plain")
			request.Host = "localhost:8080"

			res, err := ts.Client().Do(request)
			require.NoError(t, err)

			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.statusCodePost, res.StatusCode)

			//Читаем тело ответа
			url, err := io.ReadAll(res.Body)
			//Проверяем, что смогли прочитать тело, иначе тест остановится
			require.NoError(t, err)
			err = res.Body.Close()
			require.NoError(t, err)

			shortURL := string(url)
			require.NotEmpty(t, shortURL)

			shortKey := strings.ReplaceAll(shortURL, "http://localhost:8080/", "")
			require.NotEmpty(t, shortKey)

			originalURL := hd.SM.GetURL(shortKey)
			require.NotEmpty(t, originalURL)

			request, err = http.NewRequest(http.MethodGet, ts.URL+"/"+shortKey, nil)
			require.NoError(t, err)
			request.Host = "localhost:8080"

			client := ts.Client()
			client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}
			res, err = client.Do(request)
			require.NoError(t, err)
			assert.Equal(t, originalURL, res.Header.Get("Location"))
			assert.Equal(t, tt.want.statusCodeGet, res.StatusCode)
			err = res.Body.Close()
			require.NoError(t, err)
		})
	}

}

func TestApiShorten(t *testing.T) {
	hd := hd(t)
	ts := httptest.NewServer(hd.URLRouter())
	defer ts.Close()

	type want struct {
		contentType    string
		statusCodePost int
	}
	testTable := []struct {
		name string
		json models.ShortenRequest
		want want
	}{
		{
			name: "shorten with js request",
			json: models.ShortenRequest{
				URL: "https://reqbin.com/post-online",
			},
			want: want{
				contentType:    "application/json",
				statusCodePost: http.StatusCreated,
			},
		},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			js, err := json.Marshal(tt.json)
			require.NoError(t, err)
			body := bytes.NewReader(js)
			request, err := http.NewRequest(http.MethodPost, ts.URL+"/api/shorten", body)
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")
			request.Host = "localhost:8080"

			res, err := ts.Client().Do(request)
			require.NoError(t, err)

			err = res.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.statusCodePost, res.StatusCode)

			var sres models.ShortenResponse
			err = json.NewDecoder(res.Body).Decode(&sres)
			require.NoError(t, err)

		})
	}

}

func TestGzip(t *testing.T) {
	hd := hd(t)
	ts := httptest.NewServer(hd.URLRouter())
	defer ts.Close()

	type want struct {
		contentType    string
		statusCodePost int
		statusCodeGet  int
	}
	testTable := []struct {
		name string
		url  string
		want want
		json models.ShortenRequest
	}{
		{
			name: "gzip post and get url",
			url:  "https://reqbin.com/post-online",
			want: want{
				contentType:    "text/plain",
				statusCodePost: http.StatusCreated,
				statusCodeGet:  http.StatusTemporaryRedirect,
			},
		},
		{
			name: "gzip shorten with js request",
			json: models.ShortenRequest{
				URL: "https://reqbin.com/post-online",
			},
			want: want{
				contentType:    "application/json",
				statusCodePost: http.StatusCreated,
			},
		},
	}
	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			if strings.Contains(tt.name, "gzip post and get") {
				var buf bytes.Buffer
				gw := gzip.NewWriter(&buf)
				_, err := gw.Write([]byte(tt.url))
				require.NoError(t, err)
				err = gw.Close()
				require.NoError(t, err)

				data := buf.Bytes()

				// gr, err := gzip.NewReader(bytes.NewReader(data))
				// require.NoError(t, err)
				// body, err := io.ReadAll(gr)
				// require.NoError(t, err)
				// str := string(body)
				// assert.Equal(t, tt.url, str)

				body := bytes.NewReader(data)
				request, err := http.NewRequest(http.MethodPost, ts.URL+"/", body)
				require.NoError(t, err)
				request.Header.Set("Content-Type", "application/x-gzip")
				request.Header.Set("Content-Encoding", "gzip")
				request.Host = "localhost:8080"

				res, err := ts.Client().Do(request)
				require.NoError(t, err)

				assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
				assert.Equal(t, tt.want.statusCodePost, res.StatusCode)

				//Читаем тело ответа
				url, err := io.ReadAll(res.Body)
				//Проверяем, что смогли прочитать тело, иначе тест остановится
				require.NoError(t, err)
				err = res.Body.Close()
				require.NoError(t, err)

				shortURL := string(url)
				require.NotEmpty(t, shortURL)
			}
		})
	}

}

func hd(t *testing.T) handlers.HandlerData {
	cfg, err := config.Parse()
	require.NoError(t, err)
	sm := storage.NewStorageMap()
	l, err := logger.NewLogZap()
	require.NoError(t, err)
	hd := handlers.HandlerData{
		SM:  sm,
		Cfg: cfg,
		L:   l,
	}
	return hd
}
