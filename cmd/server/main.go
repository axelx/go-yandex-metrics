package main

import (
	"fmt"
	"github.com/axelx/go-yandex-metrics/internal/config"
	"github.com/axelx/go-yandex-metrics/internal/handlers"
	"github.com/axelx/go-yandex-metrics/internal/logger"
	"github.com/axelx/go-yandex-metrics/internal/storage"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func main() {

	conf := config.NewConfigServer()
	metricStorage := storage.New(conf.FlagFileStoragePath, conf.FlagStoreInternal, conf.FlagRestore)
	metricStorage.RestoreFromFile()
	//go updateMemstorage(metricStorage)

	lg := logger.Initialize("info")
	lg.Info("Running server", zap.String("address", conf.FlagRunAddr))
	fmt.Println("conf", conf)

	hd := handlers.New(&metricStorage)
	if err := http.ListenAndServe(conf.FlagRunAddr, hd.Router(lg)); err != nil {
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
