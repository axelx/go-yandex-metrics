package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"html/template"
	"internal/service"
	"net/http"
)

type keeper interface {
	SetGauge(string, float64) error
	SetCounter(string, int64) error
	GetMetric(string, string) (string, error)
	GetGaugeMetric() map[string]float64
	GetCounterMetric() map[string]int64
}

type handler struct {
	memStorage keeper
}

func New(k keeper) handler {
	return handler{
		memStorage: k,
	}
}

func (h *handler) Router() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.GetAllMetrics())
	r.Post("/update/{typeM}/{nameM}/{valueM}", h.UpdatedMetric())
	r.Get("/value/{typeM}/{nameM}", h.GetMetric())
	return r
}

func (h *handler) UpdatedMetric() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		typeM := chi.URLParam(req, "typeM")
		nameM := chi.URLParam(req, "nameM")

		valueM := chi.URLParam(req, "valueM")

		if typeM == "" || nameM == "" || valueM == "" {
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
			err = h.memStorage.SetGauge(nameM, val)
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
			err = h.memStorage.SetCounter(nameM, i)
			if err != nil {
				http.Error(res, fmt.Sprint(err), http.StatusBadRequest)
				return
			}
		default:
			http.Error(res, "StatusBadRequest", http.StatusBadRequest)
			return
		}

		res.Write([]byte(valueM))

		res.WriteHeader(http.StatusOK)
	}
}

func (h *handler) GetMetric() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		typeM := chi.URLParam(req, "typeM")
		nameM := chi.URLParam(req, "nameM")

		if typeM == "" || nameM == "" {
			http.Error(res, "StatusNotFound", http.StatusNotFound)
			return
		}

		metric, err := h.memStorage.GetMetric(typeM, nameM)
		if err != nil {
			http.Error(res, "StatusNotFound", http.StatusNotFound)
			return
		}

		res.Write([]byte(metric))

		res.WriteHeader(http.StatusOK)
	}
}

func (h *handler) GetAllMetrics() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		tmpl := template.Must(template.ParseFiles("../../internal/server/handlers/layout.html"))
		tmpl.Execute(res, struct {
			Gauge   map[string]float64
			Counter map[string]int64
		}{
			Gauge:   h.memStorage.GetGaugeMetric(),
			Counter: h.memStorage.GetCounterMetric(),
		})
	}
}
