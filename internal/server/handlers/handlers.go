package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"html/template"
	"net/http"
)

type Keeper interface {
	SetGauge(string, string) error
	SetCounter(string, string) error
	GetMetric(string, string) (string, error)
}

func UpdatedMetric(m Keeper) http.HandlerFunc {
	fmt.Println("UpdatedMetric for test   111")
	return func(res http.ResponseWriter, req *http.Request) {
		typeM := chi.URLParam(req, "typeM")
		nameM := chi.URLParam(req, "nameM")

		valueM := chi.URLParam(req, "valueM")

		fmt.Println("===", typeM, nameM, valueM)

		if typeM == "" || nameM == "" || valueM == "" {
			fmt.Println("=====")
			http.Error(res, "StatusNotFound", http.StatusNotFound)
			return
		}
		switch typeM {
		case "gauge":
			err := m.SetGauge(nameM, valueM)
			if err != nil {
				http.Error(res, fmt.Sprint(err), http.StatusBadRequest)
				return
			}
		case "counter":
			err := m.SetCounter(nameM, valueM)
			if err != nil {
				http.Error(res, fmt.Sprint(err), http.StatusBadRequest)
				return
			}
		default:
			http.Error(res, "StatusBadRequest", http.StatusBadRequest)
			return
		}
		fmt.Println(m)

		res.Write([]byte(valueM))

		res.WriteHeader(http.StatusOK)
	}
}

func GetMetric(m Keeper) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		typeM := chi.URLParam(req, "typeM")
		nameM := chi.URLParam(req, "nameM")

		if typeM == "" || nameM == "" {
			http.Error(res, "StatusNotFound", http.StatusNotFound)
			return
		}

		metric, err := m.GetMetric(typeM, nameM)
		if err != nil {
			http.Error(res, "StatusNotFound", http.StatusNotFound)
			return
		}

		body := fmt.Sprintf("Метрика тип %s название %s равна %v\r\n", typeM, nameM, metric)
		res.Write([]byte(body))

		res.WriteHeader(http.StatusOK)
	}
}

func GetAllMetrics(m Keeper) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		tmpl := template.Must(template.ParseFiles("../../internal/server/handlers/layout.html"))
		tmpl.Execute(res, m)
	}
}
