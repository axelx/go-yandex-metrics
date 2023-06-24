package handlers

import (
	"bytes"
	"fmt"
	"github.com/axelx/go-yandex-metrics/internal/logger"
	"github.com/axelx/go-yandex-metrics/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"html/template"
	"io"
	"net/http"
	"strconv"
)

type keeper interface {
	SetGauge(string, float64) error
	SetCounter(string, int64) error
	GetMetric(string, string) (string, error)
	GetTypeMetric(string) interface{}
}

type handler struct {
	memStorage keeper
}

func New(k keeper) handler {
	return handler{
		memStorage: k,
	}
}

func (h *handler) Router(flagLogLevel string) chi.Router {

	if err := logger.Initialize(flagLogLevel); err != nil {
		panic(err)
	}
	fmt.Println("conf.FlagLogLevel ", flagLogLevel)

	r := chi.NewRouter()

	r.Get("/", logger.RequestLogger(h.GetAllMetrics()))
	r.Post("/update/{typeM}/{nameM}/{valueM}", logger.RequestLogger(h.UpdatedMetric()))
	r.Get("/value/{typeM}/{nameM}", logger.RequestLogger(h.GetMetric()))
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

		size, _ := res.Write([]byte(valueM))

		res.WriteHeader(http.StatusOK)

		logger.Log.Info("sending HTTP response UpdatedMetric",
			zap.String("size", strconv.Itoa(size)),
			zap.String("status", strconv.Itoa(http.StatusOK)),
		)
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

		size, _ := res.Write([]byte(metric))

		res.WriteHeader(http.StatusOK)

		logger.Log.Info("sending HTTP response UpdatedMetric",
			zap.String("size", strconv.Itoa(size)),
			zap.String("status", strconv.Itoa(http.StatusOK)),
		)
	}
}

func (h *handler) GetAllMetrics() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		buf := bytes.NewBuffer(nil)
		ioWriter := io.MultiWriter(res, buf)

		tmpl := template.Must(template.ParseFiles("../../internal/handlers/layout.html"))

		tmpl.Execute(ioWriter, struct {
			Gauge   interface{}
			Counter interface{}
		}{
			Gauge:   h.memStorage.GetTypeMetric("gauge"),
			Counter: h.memStorage.GetTypeMetric("counter"),
		})
		tmplSize := len(buf.Bytes())

		logger.Log.Info("sending HTTP response UpdatedMetric",
			zap.String("size", strconv.Itoa(tmplSize)),
			zap.String("status", strconv.Itoa(http.StatusOK)),
		)
	}
}
