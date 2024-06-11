package main

import "os"

// .\multichecker.exe .\testdata\pkg1\main.go
// main.go:6:5: it is forbidden to use exit in the main package in the main function
func main() {
	os.Exit(1)
}

func main2() {
	os.Exit(1)
}
