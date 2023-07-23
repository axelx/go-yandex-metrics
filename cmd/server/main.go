package main

import (
	"github.com/axelx/go-yandex-metrics/internal/config"
	"github.com/axelx/go-yandex-metrics/internal/handlers"
	"github.com/axelx/go-yandex-metrics/internal/logger"
	"github.com/axelx/go-yandex-metrics/internal/pg"
	"github.com/axelx/go-yandex-metrics/internal/pg/db"
	"github.com/axelx/go-yandex-metrics/internal/storage"
	"go.uber.org/zap"
	"net/http"
)

func main() {

	conf := config.NewConfigServer()
	metricStorage := storage.New(conf.FlagFileStoragePath, conf.FlagStoreInternal, conf.FlagRestore)
	lg := logger.Initialize("info")
	lg.Info("Running server", zap.String("config", conf.String()))

	newClient := db.NewClient()
	NewDBStorage := pg.NewDBStorage(newClient)

	if err := newClient.Open(conf.FlagDatabaseDSN); err != nil {
		lg.Error("Error not connect to db",
			zap.String("about func", "main: newClient.Open"),
			zap.String("about ERR", err.Error()),
		)
	}
	// создаем таблицы при старте сервера
	if newClient.DB != nil {
		newClient.CreateTable()
	}

	defer func() {
		_ = newClient.Close()
	}()

	if conf.FlagFileStoragePath != "" {
		metricStorage.RestoreFromFile()
	}

	go metricStorage.UpdateFile()

	hd := handlers.New(&metricStorage, lg, newClient, NewDBStorage, conf.FlagHashKey)
	if err := http.ListenAndServe(conf.FlagRunAddr, hd.Router()); err != nil {
		panic(err)
	}
}
