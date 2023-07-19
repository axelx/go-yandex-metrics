package main

import (
	"fmt"
	"github.com/axelx/go-yandex-metrics/internal/config"
	"github.com/axelx/go-yandex-metrics/internal/handlers"
	"github.com/axelx/go-yandex-metrics/internal/logger"
	"github.com/axelx/go-yandex-metrics/internal/pg"
	"github.com/axelx/go-yandex-metrics/internal/storage"
	"go.uber.org/zap"
	"net/http"
)

func main() {

	conf := config.NewConfigServer()
	metricStorage := storage.New(conf.FlagFileStoragePath, conf.FlagStoreInternal, conf.FlagRestore)
	lg := logger.Initialize("info")
	lg.Info("Running server", zap.String("config", conf.String()))

	newClient := pg.NewClient()

	if err := newClient.Open(conf.FlagDatabaseDSN); err != nil {
		fmt.Println("err not connect to db", err)
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

	hd := handlers.New(&metricStorage, "info", newClient)
	if err := http.ListenAndServe(conf.FlagRunAddr, hd.Router()); err != nil {
		panic(err)
	}
}
