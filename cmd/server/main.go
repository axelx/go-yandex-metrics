package main

import (
	"github.com/axelx/go-yandex-metrics/internal/config"
	"github.com/axelx/go-yandex-metrics/internal/handlers"
	"github.com/axelx/go-yandex-metrics/internal/logger"
	"github.com/axelx/go-yandex-metrics/internal/storage"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	metricStorage := storage.New()

	conf := config.NewConfigServer()
	logger.Log.Info("Running server", zap.String("address", conf.FlagRunAddr))

	hd := handlers.New(&metricStorage)
	if err := http.ListenAndServe(conf.FlagRunAddr, hd.Router(conf.FlagLogLevel)); err != nil {
		panic(err)
	}
}
