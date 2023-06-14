package main

import (
	"fmt"
	"github.com/axelx/go-yandex-metrics/internal/server/config"
	"github.com/axelx/go-yandex-metrics/internal/server/handlers"
	"github.com/axelx/go-yandex-metrics/internal/server/storage"
	"net/http"
)

func main() {
	metricStorage := storage.New()

	conf := config.NewConfigServer()

	fmt.Println("Running server on", conf.FlagRunAddr)

	hd := handlers.New(&metricStorage)
	err := http.ListenAndServe(conf.FlagRunAddr, hd.Router())

	if err != nil {
		panic(err)
	}
}
