package main

import (
	"net/http"
)

var UrlsID = make(map[string]string)

func main() {
	err := http.ListenAndServe(`:8080`, URLRouter())
	if err != nil {
		panic(err)
	}
}
