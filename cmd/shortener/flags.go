package main

import (
	"flag"
)

var (
	flagRunAddr    string
	flagResultAddr string
)

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&flagResultAddr, "b", "http://localhost:8080", "address and port to run server")
	flag.Parse()
}
