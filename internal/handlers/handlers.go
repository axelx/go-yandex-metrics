package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/axelx/go-yandex-metrics/internal/logger"
	"github.com/axelx/go-yandex-metrics/internal/mgzip"
	"github.com/axelx/go-yandex-metrics/internal/models"
	"github.com/axelx/go-yandex-metrics/internal/mtemplate"
	"github.com/axelx/go-yandex-metrics/internal/pg"
	"github.com/axelx/go-yandex-metrics/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
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

func (h *handler) Router(log *zap.Logger, client *pg.PgStorage) chi.Router {

	r := chi.NewRouter()
	r.Use(logger.RequestLogger(log))

	r.Post("/update/{typeM}/{nameM}/{valueM}", h.UpdatedMetric(log))
	r.Get("/value/{typeM}/{nameM}", h.GetMetric(log))
	r.Get("/", mgzip.GzipHandle(h.GetAllMetrics(log, client)))
	r.Post("/update/", GzipMiddleware(h.UpdatedJSONMetric(log, client)))
	r.Post("/value/", GzipMiddleware(h.GetJSONMetric(log, client)))
	r.Get("/ping", h.DBConnect(client))
	r.Post("/updates/", GzipMiddleware(h.UpdatedJSONMetrics(log, client)))

	return r
}

func (h *handler) DBConnect(client *pg.PgStorage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if client.DB != nil {
			res.WriteHeader(http.StatusOK)
			_, err := res.Write([]byte(""))
			if err != nil {
				fmt.Println("в DbConnect что-то пошло не так 1", err)
			}

		} else {
			res.WriteHeader(http.StatusInternalServerError)
			_, err := res.Write([]byte(""))
			if err != nil {
				fmt.Println("в DbConnect что-то пошло не так 2", err)
			}
		}
	}
}

func GzipMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		acceptEncoding := req.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := mgzip.NewCompressWriter(res)
			res = cw
			defer cw.Close()
		}

		if req.Header.Get("Content-Encoding") == "gzip" {
			cr, err := mgzip.NewCompressReader(req.Body)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
			req.Body = cr
			defer cr.Close()
		}
		h.ServeHTTP(res, req)
	}
}

func (h *handler) UpdatedMetric(log *zap.Logger) http.HandlerFunc {
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

		size, err := res.Write([]byte(valueM))
		if err != nil {
			fmt.Println("в UpdatedMetric что-то пошло не так", err)
		}

		res.WriteHeader(http.StatusOK)

		log.Info("sending HTTP response UpdatedMetric",
			zap.String("size", strconv.Itoa(size)),
			zap.String("status", strconv.Itoa(http.StatusOK)),
		)
	}
}

func (h *handler) GetMetric(log *zap.Logger) http.HandlerFunc {
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

		log.Info("sending HTTP response UpdatedMetric",
			zap.String("size", strconv.Itoa(size)),
			zap.String("status", strconv.Itoa(http.StatusOK)),
		)
	}
}

func (h *handler) UpdatedJSONMetric(log *zap.Logger, client *pg.PgStorage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		log.Debug("decoding request")
		var metrics models.Metrics
		dec := json.NewDecoder(req.Body)
		if err := dec.Decode(&metrics); err != nil {
			log.Debug("cannot decode request JSON body", zap.Error(err))
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		deltaDefault := int64(0)
		valueDefault := float64(0)
		if metrics.MType == "" || metrics.ID == "" || (metrics.Delta == &deltaDefault && metrics.Value == &valueDefault) {
			http.Error(res, "StatusNotFound", http.StatusNotFound)
			return
		}

		err := setJSONorDBmetric(h.memStorage, metrics.MType, metrics.ID, metrics.Value, metrics.Delta, client)
		if err != nil {
			http.Error(res, fmt.Sprint(err), http.StatusBadRequest)
			return
		}

		metricStorage, err := getJSONorDBmetrics(h.memStorage, metrics.MType, metrics.ID, client)
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

		log.Info("sending HTTP response UpdatedMetric",
			zap.String("size", strconv.Itoa(size)),
			zap.String("status", strconv.Itoa(http.StatusOK)),
		)
	}
}
func (h *handler) UpdatedJSONMetrics(log *zap.Logger, client *pg.PgStorage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		var metrics []models.Metrics

		dec := json.NewDecoder(req.Body)
		if err := dec.Decode(&metrics); err != nil {
			log.Debug("cannot decode request JSON body", zap.Error(err))
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(metrics) == 0 {
			http.Error(res, "StatusNotFound", http.StatusNotFound)
			return
		}

		err := client.SetBatchMetrics(metrics)
		if err != nil {
			http.Error(res, fmt.Sprint(err), http.StatusBadRequest)
			return
		}
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)

		log.Info("sending HTTP response UpdatedMetric",
			zap.String("status", strconv.Itoa(http.StatusOK)),
		)
	}
}

func (h *handler) GetJSONMetric(log *zap.Logger, client *pg.PgStorage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		var metrics models.Metrics
		dec := json.NewDecoder(req.Body)
		if err := dec.Decode(&metrics); err != nil {
			log.Debug("cannot decode request JSON body", zap.Error(err))
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		if metrics.MType == "" || metrics.ID == "" {
			http.Error(res, "StatusNotFound", http.StatusNotFound)
			return
		}

		metric, err := getJSONorDBmetrics(h.memStorage, metrics.MType, metrics.ID, client)
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

		log.Info("sending HTTP response UpdatedMetric",
			zap.String("size", strconv.Itoa(size)),
			zap.String("status", strconv.Itoa(http.StatusOK)),
		)
	}
}

func (h *handler) GetAllMetrics(log *zap.Logger, client *pg.PgStorage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		buf := bytes.NewBuffer(nil)
		ioWriter := io.MultiWriter(res, buf)
		res.Header().Set("Content-Type", "text/html")

		tmpl := mtemplate.MainTemplate()
		tmpl.Execute(ioWriter, struct {
			Gauge   interface{}
			Counter interface{}
		}{
			Gauge:   getMetrics(h.memStorage, "gauge", client),
			Counter: getMetrics(h.memStorage, "counter", client),
		})
		tmplSize := len(buf.Bytes())

		log.Info("sending HTTP response UpdatedMetric",
			zap.String("size", strconv.Itoa(tmplSize)),
			zap.String("status", strconv.Itoa(http.StatusOK)),
			zap.String("about", "GetAllMetrics"),
		)
	}
}

func getJSONorDBmetrics(m keeper, MType, ID string, client *pg.PgStorage) (models.Metrics, error) {
	metricStorage := models.Metrics{}
	var err error = nil
	if client.DB == nil {
		metricStorage, err = m.GetJSONMetric(MType, ID)
	} else {
		fmt.Println(client, MType, ID)
		metricStorage, err = client.GetDBMetric(MType, ID)
	}
	if err != nil {
		return models.Metrics{}, err
	}
	return metricStorage, nil
}

func setJSONorDBmetric(m keeper, MType, ID string, value *float64, delta *int64, client *pg.PgStorage) error {
	var err error = nil

	deltaDefault := int64(0)
	valueDefault := float64(0)

	switch MType {
	case "gauge":
		if client.DB == nil {
			err = m.SetJSONGauge(ID, value)
		} else {
			err = client.SetDBMetric(MType, ID, value, &deltaDefault)
		}
		if err != nil {
			return err
		}
	case "counter":
		if client.DB == nil {
			err = m.SetJSONCounter(ID, delta)
		} else {
			err = client.SetDBMetric(MType, ID, &valueDefault, delta)
		}
		if err != nil {
			return err
		}
	default:
		return err
	}
	return nil
}

func getMetrics(m keeper, MType string, client *pg.PgStorage) interface{} {
	if client.DB == nil {
		return m.GetTypeMetric(MType)
	} else {
		return client.GetDBMetrics(MType)
	}
}
