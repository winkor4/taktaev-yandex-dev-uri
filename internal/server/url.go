package server

import (
	"context"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/model"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

func shortURL(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Can't read body", http.StatusBadRequest)
			return
		}

		contentType := r.Header.Get("Content-Type")

		var ourl string
		switch {
		case strings.Contains(contentType, "application/json"):
			var schema urlSchema
			if err := json.Unmarshal(body, &schema); err != nil {
				http.Error(w, "Can't unmarshal body", http.StatusBadRequest)
				return
			}
			ourl = schema.URL
			contentType = "application/json"
		case strings.Contains(contentType, "text/plain"):
			ourl = string(body)
			contentType = "text/plain"
		}

		if ourl == "" {
			http.Error(w, "URL parameter is missing", http.StatusBadRequest)
			return
		}

		urls := make([]model.URL, 1)
		urls[0].Key = model.ShortKey(ourl)
		urls[0].OriginalURL = ourl

		err = s.urlRepo.SaveURL(urls)
		if err != nil {
			http.Error(w, "Can't save data", http.StatusInternalServerError)
		}

		result := fmt.Sprintf(s.cfg.ResSrvAdr+"/%s", urls[0].Key)
		w.Header().Set("Content-Type", contentType)

		if urls[0].Conflict {
			w.WriteHeader(http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusCreated)
		}

		switch {
		case contentType == "application/json":
			var resSchema responseSchema
			resSchema.Result = result
			if err := json.NewEncoder(w).Encode(resSchema); err != nil {
				http.Error(w, "Can't encode response", http.StatusInternalServerError)
				return
			}
		case contentType == "text/plain":
			data := []byte(result)
			_, err = w.Write(data)
			if err != nil {
				http.Error(w, "Can't write response", http.StatusInternalServerError)
				return
			}
		}
	}
}

func getURL(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := chi.URLParam(r, "id")
		url, err := s.urlRepo.GetURL(key)
		if err != nil {
			http.Error(w, "Not found", http.StatusBadRequest)
		}
		http.Redirect(w, r, url.OriginalURL, http.StatusTemporaryRedirect)
	}
}

func pingDB(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		if err := s.urlRepo.PingDB(ctx); err != nil {
			http.Error(w, "connection could't be established", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func shortBatch(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Can't read body", http.StatusBadRequest)
			return
		}

		urls := make([]model.URL, 0)
		if err := json.Unmarshal(body, &urls); err != nil {
			http.Error(w, "Can't unmarshal body", http.StatusBadRequest)
			return
		}

		data := make([]batchResponseSchema, 0, len(urls))
		for i, url := range urls {
			urls[i].Key = model.ShortKey(url.OriginalURL)
			data = append(data, batchResponseSchema{
				CorrelationID: url.CorrelationID,
				ShortURL:      fmt.Sprintf(s.cfg.ResSrvAdr+"/%s", urls[i].Key),
			})
		}

		err = s.urlRepo.SaveURL(urls)
		if err != nil {
			http.Error(w, "Can't save data", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, "Can't encode response", http.StatusInternalServerError)
			return
		}
	}
}
