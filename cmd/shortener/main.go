package main

import (
	"net/http"
)

var UrlsID = make(map[string]string)

func main() {
	parseFlags()

	err := http.ListenAndServe(flagRunAddr, URLRouter())
	if err != nil {
		panic(err)
	}
}
