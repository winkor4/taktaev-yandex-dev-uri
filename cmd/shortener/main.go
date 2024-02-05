package main

import (
	"net/http"
)

var UrlsID = make(map[string]string)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, RootHandle)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
