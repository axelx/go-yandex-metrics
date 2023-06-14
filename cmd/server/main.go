package main

import (
	"fmt"
	"github.com/axelx/go-yandex-metrics/internal/config"
	"github.com/axelx/go-yandex-metrics/internal/handlers"
	"github.com/axelx/go-yandex-metrics/internal/storage"
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
