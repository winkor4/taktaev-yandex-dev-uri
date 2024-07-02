package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/model"

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

		user := s.user

		var ourl string
		switch {
		case strings.Contains(contentType, "application/json"):
			var schema urlSchema
			if err = json.Unmarshal(body, &schema); err != nil {
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
		urls[0].UserID = user

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

		switch contentType {
		case "application/json":
			var resSchema responseSchema
			resSchema.Result = result
			if err = json.NewEncoder(w).Encode(resSchema); err != nil {
				http.Error(w, "Can't encode response", http.StatusInternalServerError)
				return
			}
		case "text/plain":
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
		if err == model.ErrIsDeleted {
			w.WriteHeader(http.StatusGone)
			return
		}
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

		user := s.user

		urls := make([]model.URL, 0)
		if err = json.Unmarshal(body, &urls); err != nil {
			http.Error(w, "Can't unmarshal body", http.StatusBadRequest)
			return
		}

		data := make([]batchResponseSchema, 0, len(urls))
		for i, url := range urls {
			urls[i].Key = model.ShortKey(url.OriginalURL)
			urls[i].UserID = user
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

func getUsersURL(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var user string
		user, err := parseUser(r, false)
		if err != nil {
			if err == http.ErrNoCookie {
				http.Error(w, "unauthorized user", http.StatusUnauthorized)
				return
			}
			http.Error(w, "can't parse cookie", http.StatusBadRequest)
			return
		}

		urls, err := s.urlRepo.GetUsersURL(user)
		if err != nil {
			http.Error(w, "can't get user's urls", http.StatusInternalServerError)
			return
		}
		if len(urls) == 0 {
			http.Error(w, "no content", http.StatusNoContent)
		}

		for i := range urls {
			urls[i].Key = fmt.Sprintf(s.cfg.ResSrvAdr+"/%s", urls[i].Key)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(urls); err != nil {
			http.Error(w, "Can't encode response", http.StatusInternalServerError)
			return
		}
	}
}

func deleteURL(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var user string
		user, err := parseUser(r, false)
		if err != nil {
			if err == http.ErrNoCookie {
				http.Error(w, "unauthorized user", http.StatusUnauthorized)
				return
			}
			http.Error(w, "can't parse cookie", http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Can't read body", http.StatusBadRequest)
			return
		}

		keys := make([]string, 0)
		if err := json.Unmarshal(body, &keys); err != nil {
			http.Error(w, "Can't unmarshal body", http.StatusBadRequest)
			return
		}

		var data delURL
		data.user = user
		data.keys = keys
		go putDelURL(s, data)

		w.WriteHeader(http.StatusAccepted)
	}
}

func putDelURL(s *Server, data delURL) {
	s.deleteCh <- data
}

func delWorker(s *Server, wg *sync.WaitGroup) {
	for data := range s.deleteCh {
		s.urlRepo.DeleteURL(data.user, data.keys)
	}
	wg.Done()
}
