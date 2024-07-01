// storage возвращает интерфейс хранения данных на сервере.
package storage

import "github.com/winkor4/taktaev-yandex-dev-uri.git/internal/model"

// Repository описание интерфейса хранилища данных
type Repository interface {
	model.URLRepository
}
