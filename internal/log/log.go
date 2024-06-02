// Модуль log содержит функции создания и упарвления логгером.
package log

import (
	"go.uber.org/zap"
)

type Logger struct {
	*zap.SugaredLogger
}

func New() (*Logger, error) {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	defer zapLogger.Sync()

	return &Logger{SugaredLogger: zapLogger.Sugar()}, nil
}

func (l *Logger) Close() error {
	return l.SugaredLogger.Sync()
}
