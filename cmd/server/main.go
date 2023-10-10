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
	if err := logger.Initialize(conf.FlagLogLevel); err != nil {
		fmt.Println(err)
	}
	logger.Log.Info("Running server", "address"+conf.String())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	metricStorage := storage.New(conf.FlagFileStoragePath, conf.FlagStoreInternal, conf.FlagRestore)

	NewDBStorage := pg.NewDBStorage()

	//подключаемся к базе
	NewDBStorage.DB, errDB = sqlx.Connect("pgx", conf.FlagDatabaseDSN)
	if errDB != nil {
		logger.Log.Error("Error not connect to db", "about ERR"+errDB.Error())
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
