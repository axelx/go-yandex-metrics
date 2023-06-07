package main

import (
	"fmt"
	"internal/handlers"
	"internal/storage"
	"net/http"
)

func main() {
	storage.StorageTest()
	m := storage.MemStorage{map[string]float64{}, map[string]int64{}}

	fmt.Println(m)
	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, handlers.UpdatedMem(&m))

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

}
