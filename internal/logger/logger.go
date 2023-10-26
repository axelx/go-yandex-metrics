package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var log = Logger{}

type Logger struct {
	lg       *zap.Logger
	Title    string
	Comment  string
	Comments []string
}

func Info(t, c string) {
	if log.lg == nil {
		panic("The logger is not initialized")
	}
	log.lg.Info(t,
		zap.String("info", c),
	)
}
func Error(t, c string) {
	if log.lg == nil {
		panic("The logger is not initialized")
	}
	log.lg.Error(t,
		zap.String("error", c),
	)
}
func Debug(t, c string) {
	if log.lg == nil {
		panic("The logger is not initialized")
	}
	log.lg.Debug(t,
		zap.String("debug", c),
	)
}

// Initialize инициализирует синглтон логера с необходимым уровнем логирования.
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
	log.lg = zl
	return nil
}

// RequestLogger — middleware-логер для входящих HTTP-запросов.
func RequestLogger() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			h.ServeHTTP(w, r)
			duration := time.Since(start)
			log.lg.Info("got incoming HTTP request (middleware)",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("duration", duration.String()),
			)
		})
	}
}
