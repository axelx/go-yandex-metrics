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

	// обрабатываем аргументы командной строки
	parseFlags(&conf)

	fmt.Println("Running server on", conf.FlagRunAddr)

	err := http.ListenAndServe(conf.FlagRunAddr, handlers.Router(metricStorage))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}
