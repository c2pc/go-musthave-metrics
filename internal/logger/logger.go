package logger

import (
	"go.uber.org/zap"
)

var Log = Logger{zap.NewNop()}

type Logger struct {
	logger *zap.Logger
}

func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()

	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = Logger{zl}

	return nil
}

func (log *Logger) Sync() error {
	return log.logger.Sync()
}

func (log *Logger) Debug(msg string, fields ...Field) {
	log.logger.Debug(msg, convertFields(fields...)...)
}

func (log *Logger) Info(msg string, fields ...Field) {
	log.logger.Info(msg, convertFields(fields...)...)
}

func (log *Logger) Warn(msg string, fields ...Field) {
	log.logger.Warn(msg, convertFields(fields...)...)
}

func (log *Logger) Error(msg string, fields ...Field) {
	log.logger.Error(msg, convertFields(fields...)...)
}

func (log *Logger) Panic(msg string, fields ...Field) {
	log.logger.Panic(msg, convertFields(fields...)...)
}

func (log *Logger) Fatal(msg string, fields ...Field) {
	log.logger.Fatal(msg, convertFields(fields...)...)
}
