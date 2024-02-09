package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURLRouter(t *testing.T) {
	ts := httptest.NewServer(URLRouter())
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

	flagRunAddr = "localhost:8080"
	flagResultAddr = "http://localhost:8080"

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

			originalURL := UrlsID[shortKey]
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
