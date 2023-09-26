// server для получения и отображения метрик которые отправляет агент
package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"net/http"

	"github.com/axelx/go-yandex-metrics/internal/config"
	"github.com/axelx/go-yandex-metrics/internal/handlers"
	"github.com/axelx/go-yandex-metrics/internal/logger"
	"github.com/axelx/go-yandex-metrics/internal/pg"
	"github.com/axelx/go-yandex-metrics/internal/storage"
)

var err = errors.New("")

func main() {

	conf := config.NewConfigServer()
	if err := logger.Initialize(conf.FlagLogLevel); err != nil {
		fmt.Println(err)
	}
	logger.Log.Info("Running server", zap.String("address", conf.String()))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	metricStorage := storage.New(conf.FlagFileStoragePath, conf.FlagStoreInternal, conf.FlagRestore)

	NewDBStorage := pg.NewDBStorage()

	//подключаемся к базе
	NewDBStorage.DB, err = sqlx.Connect("pgx", conf.FlagDatabaseDSN)
	if err != nil {
		logger.Log.Error("Error not connect to db", zap.String("about ERR", err.Error()))
	} else {
		cancel()
	}
	NewDBStorage.DB.SetMaxOpenConns(10)

	defer func() {
		NewDBStorage.DB.Close()
	}()

	if NewDBStorage.DB != nil {
		NewDBStorage.CreateTable()
	}

	if conf.FlagFileStoragePath != "" {
		metricStorage.RestoreFromFile()
	}

	go metricStorage.UpdateFile(ctx)

	hd := handlers.New(&metricStorage, NewDBStorage.DB, NewDBStorage, conf.FlagHashKey)
	if err := http.ListenAndServe(conf.FlagRunAddr, hd.Router()); err != nil {
		panic(err)
	}
}
