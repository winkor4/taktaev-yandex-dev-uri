package main

import (
	"testing"

	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/model"
)

// Чтобы запустить бенчмарк, воспользуемся утилитой go test:
// go test -bench .

// Флаг -bench указывает утилите,
// что вместе с тестами в пакете надо выполнить бенчмарки,
// удовлетворяющие переданному регулярному выражению.

func BenchmarkAPI(b *testing.B) {

	str := "https://www.youtube.com"

	for i := 0; i < b.N; i++ {
		model.ShortKey(str)
	}

}
