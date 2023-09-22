package main

import (
	"errors"
	"net/http"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/axelx/go-yandex-metrics/internal/config"
	"github.com/axelx/go-yandex-metrics/internal/handlers"
	"github.com/axelx/go-yandex-metrics/internal/logger"
	"github.com/axelx/go-yandex-metrics/internal/pg"
	"github.com/axelx/go-yandex-metrics/internal/storage"
)

var err = errors.New("")

func main() {

	conf := config.NewConfigServer()
	lg := logger.Initialize("info")
	lg.Info("Running server", zap.String("config", conf.String()))
	metricStorage := storage.New(conf.FlagFileStoragePath, conf.FlagStoreInternal, conf.FlagRestore, lg)

	NewDBStorage := pg.NewDBStorage(lg)

	//подключаемся к базе
	NewDBStorage.DB, err = sqlx.Connect("pgx", conf.FlagDatabaseDSN)
	if err != nil {
		lg.Error("Error not connect to db", zap.String("about ERR", err.Error()))
	}
	NewDBStorage.DB.SetMaxOpenConns(10)

	defer func() {
		NewDBStorage.DB.Close()
	}()

	// создаем таблицы при старте сервера
	if NewDBStorage.DB != nil {
		NewDBStorage.CreateTable()
	}

	if conf.FlagFileStoragePath != "" {
		metricStorage.RestoreFromFile()
	}

	go metricStorage.UpdateFile()

	hd := handlers.New(&metricStorage, lg, NewDBStorage.DB, NewDBStorage, conf.FlagHashKey)
	if err := http.ListenAndServe(conf.FlagRunAddr, hd.Router()); err != nil {
		panic(err)
	}
}
