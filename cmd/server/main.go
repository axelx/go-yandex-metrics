// server для получения и отображения метрик которые отправляет агент
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"

	"github.com/axelx/go-yandex-metrics/internal/config"
	"github.com/axelx/go-yandex-metrics/internal/handlers"
	"github.com/axelx/go-yandex-metrics/internal/logger"
	"github.com/axelx/go-yandex-metrics/internal/pg"
	"github.com/axelx/go-yandex-metrics/internal/storage"
)

var errDB = errors.New("")

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version:= %s\n", buildVersion)
	fmt.Printf("Build date:= %s\n", buildDate)
	fmt.Printf("Build commit:= %s\n", buildCommit)

	conf := config.NewConfigServer()
	if err := logger.Initialize(conf.LogLevel); err != nil {
		fmt.Println(err)
	}
	logger.Log.Info("Running server", "address"+conf.String())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	metricStorage := storage.New(conf.FileStoragePath, conf.StoreInternal, conf.Restore)

	NewDBStorage := pg.NewDBStorage()

	//подключаемся к базе
	NewDBStorage.DB, errDB = sqlx.Connect("pgx", conf.DatabaseDSN)
	if errDB != nil {
		logger.Log.Error("Error not connect to db", "about ERR"+errDB.Error())
		go metricStorage.UpdateFile(ctx)
		if conf.FileStoragePath != "" {
			metricStorage.RestoreFromFile()
		}
	} else {
		NewDBStorage.DB.SetMaxOpenConns(10)

		defer func() {
			NewDBStorage.DB.Close()
		}()

		if NewDBStorage.DB != nil {
			NewDBStorage.CreateTable()
		}
	}

	hd := handlers.New(&metricStorage, NewDBStorage.DB, NewDBStorage, conf.HashKey, conf.CryptoKey)
	if err := http.ListenAndServe(conf.RunAddr, hd.Router()); err != nil {
		panic(err)
	}
}
