package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/axelx/go-yandex-metrics/internal/logger"
	"github.com/axelx/go-yandex-metrics/internal/mgzip"
	"github.com/axelx/go-yandex-metrics/internal/models"
	"github.com/axelx/go-yandex-metrics/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type keeper interface {
	SetGauge(string, float64) error
	SetCounter(string, int64) error
	GetMetric(string, string) (string, error)

	SetJSONGauge(string, *float64) error
	SetJSONCounter(string, *int64) error
	GetJSONMetric(string, string) (models.Metrics, error)

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

	r.Get("/", logger.RequestLogger(mgzip.GzipHandle(h.GetAllMetrics())))
	//r.Get("/", logger.RequestLogger(h.GetAllMetrics()))
	r.Post("/update/", logger.RequestLogger(GzipMiddleware(h.UpdatedJSONMetric())))
	r.Post("/value/", logger.RequestLogger(GzipMiddleware(h.GetJSONMetric())))

	r.Post("/update/{typeM}/{nameM}/{valueM}", logger.RequestLogger(h.UpdatedMetric()))
	r.Get("/value/{typeM}/{nameM}", logger.RequestLogger(h.GetMetric()))

	return r
}

func GzipMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// по умолчанию устанавливаем оригинальный http.ResponseWriter как тот,
		// который будем передавать следующей функции
		ow := res

		// проверяем, что клиент умеет получать от сервера сжатые данные в формате mgzip
		acceptEncoding := req.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			cw := mgzip.NewCompressWriter(res)
			// меняем оригинальный http.ResponseWriter на новый
			ow = cw
			// не забываем отправить клиенту все сжатые данные после завершения middleware
			defer cw.Close()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате mgzip
		contentEncoding := req.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			cr, err := mgzip.NewCompressReader(req.Body)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
			// меняем тело запроса на новое
			req.Body = cr
			defer cr.Close()
		}

		// передаём управление хендлеру
		h.ServeHTTP(ow, req)
	}
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

func (h *handler) UpdatedJSONMetric() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		logger.Log.Debug("decoding request")
		var metrics models.Metrics
		dec := json.NewDecoder(req.Body)
		if err := dec.Decode(&metrics); err != nil {
			logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		iz := int64(0)
		fz := float64(0)
		if metrics.MType == "" || metrics.ID == "" || (metrics.Delta == &iz && metrics.Value == &fz) {
			http.Error(res, "StatusNotFound", http.StatusNotFound)
			return
		}
		switch metrics.MType {
		case "gauge":
			err := h.memStorage.SetJSONGauge(metrics.ID, metrics.Value)
			if err != nil {
				http.Error(res, fmt.Sprint(err), http.StatusBadRequest)
				return
			}
		case "counter":
			err := h.memStorage.SetJSONCounter(metrics.ID, metrics.Delta)
			if err != nil {
				http.Error(res, fmt.Sprint(err), http.StatusBadRequest)
				return
			}
		default:
			http.Error(res, "StatusBadRequest", http.StatusBadRequest)
			return
		}

		metricStorage, err := h.memStorage.GetJSONMetric(metrics.MType, metrics.ID)
		if err != nil {
			http.Error(res, "StatusNotFound", http.StatusNotFound)
			return
		}

		metricsJSON, err := json.Marshal(metricStorage)
		if err != nil {
			fmt.Println("Error json marshal metrics:", err)
		}
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		size, _ := res.Write(metricsJSON)

		logger.Log.Info("sending HTTP response UpdatedMetric",
			zap.String("size", strconv.Itoa(size)),
			zap.String("status", strconv.Itoa(http.StatusOK)),
		)
	}
}

func (h *handler) GetJSONMetric() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		var metrics models.Metrics
		dec := json.NewDecoder(req.Body)
		if err := dec.Decode(&metrics); err != nil {
			logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		if metrics.MType == "" || metrics.ID == "" {
			http.Error(res, "StatusNotFound", http.StatusNotFound)
			return
		}

		metric, err := h.memStorage.GetJSONMetric(metrics.MType, metrics.ID)
		if err != nil {
			http.Error(res, "StatusNotFound", http.StatusNotFound)
			return
		}

		metricsJSON, err := json.Marshal(metric)
		if err != nil {
			fmt.Println("Error json marshal metrics:", err)
		}
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		size, _ := res.Write(metricsJSON)

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

		//tmpl := template.Must(template.ParseFiles("../../internal/handlers/layout.html"))
		tmpl := template.Must(template.New("html-tmpl").Parse("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n    <meta charset=\"UTF-8\">\n    <title>Title</title>\n</head>\n<body>\n<h1>Метрики</h1>\n\n<h2>Gauge</h2>\n<ul>\n    {{range $name, $val := .Gauge}}\n    <li>{{$name}} - {{$val}}</li>`\n    {{end}}\n</ul>\n\n<h2>Counter</h2>\n<ul>\n    {{range $name, $val := .Counter}}\n    <li>{{$name}} - {{$val}}</li>`\n    {{end}}\n</ul>\n\n</body>\n</html>"))

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
