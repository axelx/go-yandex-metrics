package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"html/template"
	"internal/service"
	"internal/storage"
	"net/http"
)

type Keeper interface {
	SetGauge(string, float64) error
	SetCounter(string, int64) error
	GetMetric(string, string) (string, error)
}

func Router(m storage.MemStorage) chi.Router {
	r := chi.NewRouter()

	r.Get("/", GetAllMetrics(&m))
	r.Get("/value/{typeM}/{nameM}", GetMetric(&m))
	r.Post("/update/{typeM}/{nameM}/{valueM}", UpdatedMetric(&m))
	return r
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
			val, err := service.PrepareFloat64Data(valueM)
			if err != nil {
				http.Error(res, fmt.Sprint(err), http.StatusBadRequest)
				return
			}
			err = m.SetGauge(nameM, val)
			if err != nil {
				http.Error(res, fmt.Sprint(err), http.StatusBadRequest)
				return
			}
		case "counter":
			i, err := service.PrepareInt64Data(valueM)
			if err != nil {
				http.Error(res, fmt.Sprint(err), http.StatusBadRequest)
				return
			}
			err = m.SetCounter(nameM, i)
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

		res.Write([]byte(metric))

		res.WriteHeader(http.StatusOK)
	}
}

func GetAllMetrics(m Keeper) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		tmpl := template.Must(template.ParseFiles("../../internal/server/handlers/layout.html"))
		tmpl.Execute(res, m)
	}
}
