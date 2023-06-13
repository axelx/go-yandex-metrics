package main

import (
	"fmt"
	"internal/config"
	"internal/handlers"
	"internal/storage"
	"net/http"
)

func main() {
	metricStorage := storage.NewStorage()
	conf := config.NewConfigServerFlag()
	parseFlags(&conf)

	fmt.Println("Running server on", conf.FlagRunAddr)

	hd := handlers.New(&metricStorage)
	err := http.ListenAndServe(conf.FlagRunAddr, hd.Router())

	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}
