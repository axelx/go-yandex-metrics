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
	"time"
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

	go updateMemstorage(metricStorage)

	hd := handlers.New(&metricStorage)
	if err := http.ListenAndServe(conf.FlagRunAddr, hd.Router(lg, newClient)); err != nil {
		panic(err)
	}
}

func updateMemstorage(metricStorage storage.MemStorage) {
	if metricStorage.UpdateInterval == 0 {
		return
	}
	for {
		metricStorage.SaveMetricToFile()
		fmt.Println("updateMemstorage from file ")
		time.Sleep(time.Duration(metricStorage.UpdateInterval) * time.Second)
	}
}
