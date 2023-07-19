package logger

import (
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// Initialize инициализирует логер с необходимым уровнем логирования.
func Initialize(level string) *zap.Logger {
	// преобразуем текстовый уровень логирования в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		fmt.Println("Ошибка инициализации уровня логирования")
	}
	// создаём новую конфигурацию логера
	cfg := zap.NewProductionConfig()
	// устанавливаем уровень
	cfg.Level = lvl
	// создаём логер на основе конфигурации
	zl, err := cfg.Build()
	if err != nil {
		fmt.Println("Ошибка инициализации лога")
	}
	return zl
}

// RequestLogger — middleware-логер для входящих HTTP-запросов.
func RequestLogger(log *zap.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			h.ServeHTTP(w, r)
			duration := time.Since(start)

			log.Info("got incoming HTTP request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("duration", duration.String()),
				zap.String("РУДДЩ", "0000000"),
			)
		})
	}
}
