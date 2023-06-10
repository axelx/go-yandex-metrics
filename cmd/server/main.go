package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"internal/handlers"
	"internal/storage"
	"net/http"
)

func main() {
	storage.StorageTest()
	m := storage.MemStorage{Gauge: map[string]float64{}, Counter: map[string]int64{}}

	err := http.ListenAndServe(`:8080`, router(m))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

}

func router(m storage.MemStorage) chi.Router {
	r := chi.NewRouter()

	r.Get("/", handlers.GetAllMetrics(&m))
	r.Get("/value/{typeM}/{nameM}", handlers.GetMetric(&m))
	r.Post("/update/{typeM}/{nameM}/{valueM}", handlers.UpdatedMetric(&m))
	return r
}
