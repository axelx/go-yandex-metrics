package main_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
)

func Example() {
	// Endpoing UpdateJSON
	// Создаем тело запроса в виде строки
	requestBody := []byte(`{"id": "RandomValue", "type": "gauge", "value": 0.815`)
	req, err := http.NewRequest("POST", "/update/", bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println(err)
	}
	// Создаем ResponseRecorder (реализация интерфейса http.ResponseWriter)
	rr := httptest.NewRecorder()
	// Вызываем обработчик соответствующего эндпоинта
	handler := http.HandlerFunc(valueHandler)
	handler.ServeHTTP(rr, req)
	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String())

	// Endpoing PING
	req, err = http.NewRequest("POST", "/ping", nil)
	if err != nil {
		fmt.Println(err)
	}
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(pingHandler)
	handler.ServeHTTP(rr, req)
	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String())

	// Endpoing UpdateMetric
	req, err = http.NewRequest("POST", "/update/gauge/someMetric/33.337", nil)
	if err != nil {
		fmt.Println(err)
	}
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(updateMetricHandler)
	handler.ServeHTTP(rr, req)
	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String())

	// Endpoing GetValue
	requestBody = []byte(`{"id": "RandomValue", "type": "gauge"`)
	req, err = http.NewRequest("POST", "/value/", bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println(err)
	}
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(valueHandler)
	handler.ServeHTTP(rr, req)
	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String())

	// Output:
	// 200
	// {"id": "RandomValue", "type": "gauge", "value": 0.815}
	// 200
	//
	// 200
	// 33.337
	// 200
	// {"id": "RandomValue", "type": "gauge", "value": 0.815}
}

// Обработчик эндпоинта /value/ И эндпоинт /update/
func valueHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"id": "RandomValue", "type": "gauge", "value": 0.815}`))
}

// Обработчик эндпоинта /value/
func updateMetricHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`33.337`))
}

// Обработчик эндпоинта /ping/
func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(``))
}
