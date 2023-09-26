package main_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
)

func ExampleUpdateJSON() {

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

	// Output:
	// 200
	// {"id": "RandomValue", "type": "gauge", "value": 0.815}
}
func ExamplePing() {

	req, err := http.NewRequest("POST", "/ping", nil)
	if err != nil {
		fmt.Println(err)
	}
	// Создаем ResponseRecorder (реализация интерфейса http.ResponseWriter)
	rr := httptest.NewRecorder()

	// Вызываем обработчик соответствующего эндпоинта
	handler := http.HandlerFunc(pingHandler)
	handler.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String())

	//Output:
	// 200
	//
}
func ExampleUpdateMetric() {

	req, err := http.NewRequest("POST", "/update/gauge/someMetric421/33.337", nil)
	if err != nil {
		fmt.Println(err)
	}
	// Создаем ResponseRecorder (реализация интерфейса http.ResponseWriter)
	rr := httptest.NewRecorder()

	// Вызываем обработчик соответствующего эндпоинта
	handler := http.HandlerFunc(updateMetricHandler)
	handler.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String())

	//Output:
	// 200
	// 33.337
}

func ExampleValue() {
	// Создаем тело запроса в виде строки
	requestBody := []byte(`{"id": "RandomValue", "type": "gauge"`)

	req, err := http.NewRequest("POST", "/value/", bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println(err)
	}

	// Создаем ResponseRecorder (реализация интерфейса http.ResponseWriter)
	rr := httptest.NewRecorder()

	// Вызываем обработчик соответствующего эндпоинта
	handler := http.HandlerFunc(valueHandler)
	handler.ServeHTTP(rr, req)

	// Проверяем код ответа
	fmt.Println(rr.Code)

	// Проверяем тело ответа
	fmt.Println(rr.Body.String())

	// Output:
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
