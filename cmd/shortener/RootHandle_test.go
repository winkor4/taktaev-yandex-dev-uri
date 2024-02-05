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

func TestRootHandle(t *testing.T) {
	type want struct {
		contentType    string
		statusCodePost int
		statusCodeGet  int
	}
	tests := []struct {
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(tt.url)
			request := httptest.NewRequest(http.MethodPost, "/", body)
			request.Header.Set("Content-Type", "text/plain")
			request.Host = "localhost:8080"
			w := httptest.NewRecorder()
			h := http.HandlerFunc(RootHandle)
			h(w, request)

			res := w.Result()
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.statusCodePost, res.StatusCode)

			//Читаем тело ответа
			url, err := io.ReadAll(res.Body)
			//Проверяем, что смогли прочитать тело, иначе тест остановится
			require.NoError(t, err)
			//Закрываем тело
			err = res.Body.Close()
			//Проверяем, что смогли закрыть тело, иначе тест остановится
			require.NoError(t, err)

			shortURL := string(url)
			require.NotEmpty(t, shortURL)

			shortKey := strings.ReplaceAll(shortURL, "http://localhost:8080/", "")
			require.NotEmpty(t, shortKey)

			originalURL := UrlsID[shortKey]
			require.NotEmpty(t, originalURL)

			request = httptest.NewRequest(http.MethodGet, "/"+shortKey, nil)
			request.Header.Set("Content-Type", "text/plain")
			request.Host = "localhost:8080"
			w = httptest.NewRecorder()
			h = http.HandlerFunc(RootHandle)
			h(w, request)

			res = w.Result()
			assert.Equal(t, originalURL, res.Header.Get("Location"))
			assert.Equal(t, tt.want.statusCodeGet, res.StatusCode)
		})
	}
}
