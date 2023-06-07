package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type (
	MemStorage struct {
		Gauge   map[string]float64 //новое значение должно замещать предыдущее.
		Counter map[string]int64   //новое значение должно добавляться к предыдущему, если какое-то значение уже было известно серверу.
	}
)

type Keeper interface {
	gauge(string, string) error
	count(string, string) error
}

func (m *MemStorage) gauge(nameMetric, data string) error {
	_, b := m.Gauge[nameMetric]
	if f, err := strconv.ParseFloat(data, 64); err == nil && b {
		m.Gauge[nameMetric] = f
		return nil
	}
	return errors.New("ошибка обработки параметра gauge")
}

func (m *MemStorage) count(nameMetric, data string) error {
	_, b := m.Counter[nameMetric]

	if i, err := strconv.ParseInt(data, 10, 64); err == nil && b {
		m.Counter[nameMetric] += i
		return nil
	}
	return errors.New("ошибка обработки параметра counter " + nameMetric)
}

func updMem(m Keeper) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		s := strings.Split(req.URL.String(), "/")

		// s[2] <ТИП_МЕТРИКИ>  s[3] <ИМЯ_МЕТРИКИ> s[4] <ЗНАЧЕНИЕ_МЕТРИКИ>
		if cap(s) != 5 {
			http.Error(res, "StatusNotFound", http.StatusNotFound)
			return
		}

		if s[2] == "gauge" {
			err := m.gauge(s[3], s[4])
			if err != nil {
				http.Error(res, fmt.Sprint(err), http.StatusNotFound)
				return
			}
		} else if s[2] == "counter" {
			err := m.count(s[3], s[4])
			if err != nil {
				http.Error(res, fmt.Sprint(err), http.StatusNotFound)
				return
			}
		} else {
			http.Error(res, "StatusBadRequest", http.StatusBadRequest)
			return
		}
		fmt.Println(m)

		body := fmt.Sprintf("Метрика тип %s название %s обновлена %v\r\n", s[2], s[3], s[4])
		res.Write([]byte(body))

		res.WriteHeader(http.StatusOK)
	}
}

func (m MemStorage) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	data := []byte("Привет! MemStorage")
	res.Write(data)
}

func main() {
	m := MemStorage{map[string]float64{
		"someMetric":  10.00,
		"Alloc":       10,
		"BuckHashSys": 10,
		"TotalAlloc":  1,
	}, map[string]int64{
		"someMetric": 0,
		"PollCount":  0,
	}}

	fmt.Println(m)
	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, updMem(&m))

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

}
