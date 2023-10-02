package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var Log = Logger{}

type Logger struct {
	Lg       *zap.Logger
	Title    string
	Comment  string
	Comments []string
}

func (l *Logger) Info(t, c string) {
	l.Lg.Info(t,
		zap.String("info", c),
	)
}
func (l *Logger) Error(t, c string) {
	l.Lg.Error(t,
		zap.String("error", c),
	)
}
func (l *Logger) Debug(t, c string) {
	l.Lg.Debug(t,
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
	Log.Lg = zl
	return nil
}

// RequestLogger — middleware-логер для входящих HTTP-запросов.
func RequestLogger() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			h.ServeHTTP(w, r)
			duration := time.Since(start)
			Log.Lg.Info("got incoming HTTP request (middleware)",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("duration", duration.String()),
			)
		})
	}
}
