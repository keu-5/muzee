package infrastructure

import (
	"go.uber.org/zap"
)

type Logger struct {
	*zap.SugaredLogger
}

func NewLogger() *Logger {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

	logger, err := config.Build()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	return &Logger{
		SugaredLogger: logger.Sugar(),
	}
}

func NewDevelopmentLogger() *Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("failed to initialize development logger: " + err.Error())
	}

	return &Logger{
		SugaredLogger: logger.Sugar(),
	}
}

func (l *Logger) Sync() {
	_ = l.SugaredLogger.Sync()
}
