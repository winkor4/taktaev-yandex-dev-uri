// Модуль log содержит функции создания и упарвления логгером.
package log

import (
	"go.uber.org/zap"
)

// Logger описывает структуру логера
type Logger struct {
	*zap.SugaredLogger
}

// New возвращает новый логер
func New() (*Logger, error) {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = zapLogger.Sync()
	}()

	return &Logger{SugaredLogger: zapLogger.Sugar()}, nil
}

// Close закрывает логер
func (l *Logger) Close() error {
	return l.Sync()
}
