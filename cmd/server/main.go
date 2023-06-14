package main

import (
	"fmt"
	"internal/config"
	"internal/handlers"
	"internal/storage"
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
