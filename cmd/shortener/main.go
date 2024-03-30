package main

import (
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/pkg/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
