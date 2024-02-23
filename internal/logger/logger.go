package logger

import "go.uber.org/zap"

func NewLogZap() (*zap.SugaredLogger, error) {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	defer zapLogger.Sync()

	sugar := *zapLogger.Sugar()
	return &sugar, err
}
