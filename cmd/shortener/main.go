package main

// var buildVersion string
// var buildDate string
// var buildCommit string

// Build version: <buildVersion> (или "N/A" при отсутствии значения)
// Build date: <buildDate> (или "N/A" при отсутствии значения)
// Build commit: <buildCommit> (или "N/A" при отсутствии значения)

import (
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/pkg/app"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
