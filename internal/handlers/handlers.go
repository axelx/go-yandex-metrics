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
	"github.com/axelx/go-yandex-metrics/internal/pg/db"
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
	Logger     *zap.Logger
	ClientDB   *db.DB
	DbPostgres *pg.PgStorage
}

func New(k keeper, logLevel string, newClient *db.DB, NewDBStorage *pg.PgStorage) handler {
	return handler{
		memStorage: k,
		Logger:     logger.Initialize(logLevel),
		ClientDB:   newClient,
		DbPostgres: NewDBStorage,
	}
}

func (h *handler) Router() chi.Router {

	r := chi.NewRouter()
	r.Use(logger.RequestLogger(h.Logger))

	r.Post("/update/{typeM}/{nameM}/{valueM}", h.UpdatedMetric())
	r.Get("/value/{typeM}/{nameM}", h.GetMetric())
	r.Get("/", mgzip.GzipHandle(h.GetAllMetrics()))
	r.Post("/update/", GzipMiddleware(h.UpdatedJSONMetric()))
	r.Post("/value/", GzipMiddleware(h.GetJSONMetric()))
	r.Get("/ping", h.DBConnect())
	r.Post("/updates/", GzipMiddleware(h.UpdatedJSONMetrics()))

	return r
}

func (h *handler) DBConnect() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if h.ClientDB.DB != nil {
			res.WriteHeader(http.StatusOK)
			_, err := res.Write([]byte(""))
			if err != nil {
				h.Logger.Error("Error res.Write(metricsJSON)",
					zap.String("about func", "DbConnect"),
					zap.String("about ERR", err.Error()),
				)
			}
		} else {
			res.WriteHeader(http.StatusInternalServerError)
			_, err := res.Write([]byte(""))
			if err != nil {
				h.Logger.Error("Error res.Write(metricsJSON)",
					zap.String("about func", "DbConnect"),
					zap.String("about ERR", err.Error()),
				)
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

		size, err := res.Write([]byte(valueM))
		if err != nil {
			h.Logger.Error("Error res.Write(metricsJSON)",
				zap.String("about func", "UpdatedMetric"),
				zap.String("about ERR", err.Error()),
			)
		}

		res.WriteHeader(http.StatusOK)

		h.Logger.Info("sending HTTP response UpdatedMetric",
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

		size, err := res.Write([]byte(metric))
		if err != nil {
			h.Logger.Error("Error res.Write(metricsJSON)",
				zap.String("about func", "GetMetric"),
				zap.String("about ERR", err.Error()),
			)
		}

		res.WriteHeader(http.StatusOK)

		h.Logger.Info("sending HTTP response UpdatedMetric",
			zap.String("size", strconv.Itoa(size)),
			zap.String("status", strconv.Itoa(http.StatusOK)),
		)
	}
}

func (h *handler) UpdatedJSONMetric() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		h.Logger.Debug("decoding request")
		var metrics models.Metrics
		dec := json.NewDecoder(req.Body)
		if err := dec.Decode(&metrics); err != nil {
			h.Logger.Debug("cannot decode request JSON body", zap.Error(err))
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		deltaDefault := int64(0)
		valueDefault := float64(0)
		if metrics.MType == "" || metrics.ID == "" || (metrics.Delta == &deltaDefault && metrics.Value == &valueDefault) {
			http.Error(res, "StatusNotFound", http.StatusNotFound)
			return
		}

		err := setJSONorDBmetric(h.memStorage, metrics.MType, metrics.ID, metrics.Value, metrics.Delta, h.ClientDB, h.DbPostgres)
		if err != nil {
			http.Error(res, fmt.Sprint(err), http.StatusBadRequest)
			return
		}

		metricStorage, err := getJSONorDBmetrics(h.memStorage, metrics.MType, metrics.ID, h.ClientDB, h.DbPostgres)
		if err != nil {
			http.Error(res, "StatusNotFound", http.StatusNotFound)
			return
		}

		metricsJSON, err := json.Marshal(metricStorage)
		if err != nil {
			h.Logger.Error("Error json.Marshal(metricStorage)",
				zap.String("about func", "UpdatedJSONMetric"),
				zap.String("about ERR", err.Error()),
			)
		}
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		size, err := res.Write(metricsJSON)
		if err != nil {
			h.Logger.Error("Error res.Write(metricsJSON)",
				zap.String("about func", "UpdatedJSONMetric"),
				zap.String("about ERR", err.Error()),
			)
		}

		h.Logger.Info("sending HTTP response UpdatedMetric",
			zap.String("size", strconv.Itoa(size)),
			zap.String("status", strconv.Itoa(http.StatusOK)),
		)
	}
}
func (h *handler) UpdatedJSONMetrics() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		var metrics []models.Metrics

		dec := json.NewDecoder(req.Body)
		if err := dec.Decode(&metrics); err != nil {
			h.Logger.Debug("cannot decode request JSON body", zap.Error(err))
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(metrics) == 0 {
			http.Error(res, "StatusNotFound", http.StatusNotFound)
			return
		}

		err := h.DbPostgres.SetBatchMetrics(metrics)
		if err != nil {
			http.Error(res, fmt.Sprint(err), http.StatusBadRequest)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("{}"))

		h.Logger.Info("sending HTTP response UpdatedMetric",
			zap.String("status", strconv.Itoa(http.StatusOK)),
		)
	}
}

func (h *handler) GetJSONMetric() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		var metrics models.Metrics
		dec := json.NewDecoder(req.Body)
		if err := dec.Decode(&metrics); err != nil {
			h.Logger.Debug("cannot decode request JSON body", zap.Error(err))
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		if metrics.MType == "" || metrics.ID == "" {
			http.Error(res, "StatusNotFound", http.StatusNotFound)
			return
		}

		metric, err := getJSONorDBmetrics(h.memStorage, metrics.MType, metrics.ID, h.ClientDB, h.DbPostgres)
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

		h.Logger.Info("sending HTTP response UpdatedMetric",
			zap.String("size", strconv.Itoa(size)),
			zap.String("status", strconv.Itoa(http.StatusOK)),
		)
	}
}

func (h *handler) GetAllMetrics() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		buf := bytes.NewBuffer(nil)
		ioWriter := io.MultiWriter(res, buf)
		res.Header().Set("Content-Type", "text/html")

		tmpl := mtemplate.MainTemplate()
		tmpl.Execute(ioWriter, struct {
			Gauge   interface{}
			Counter interface{}
		}{
			Gauge:   getMetrics(h.memStorage, "gauge", h.ClientDB, h.DbPostgres),
			Counter: getMetrics(h.memStorage, "counter", h.ClientDB, h.DbPostgres),
		})
		tmplSize := len(buf.Bytes())

		h.Logger.Info("sending HTTP response UpdatedMetric",
			zap.String("size", strconv.Itoa(tmplSize)),
			zap.String("status", strconv.Itoa(http.StatusOK)),
			zap.String("about", "GetAllMetrics"),
		)
	}
}

func getJSONorDBmetrics(m keeper, MType, ID string, client *db.DB, DbPostgres *pg.PgStorage) (models.Metrics, error) {
	metricStorage := models.Metrics{}
	var err error = nil
	if client.DB == nil {
		metricStorage, err = m.GetJSONMetric(MType, ID)
	} else {
		fmt.Println(client, MType, ID)
		metricStorage, err = DbPostgres.GetDBMetric(MType, ID)
	}
	if err != nil {
		return models.Metrics{}, err
	}
	return metricStorage, nil
}

func setJSONorDBmetric(m keeper, MType, ID string, value *float64, delta *int64, client *db.DB, DbPostgres *pg.PgStorage) error {
	var err error = nil

	deltaDefault := int64(0)
	valueDefault := float64(0)

	switch MType {
	case "gauge":
		if client.DB == nil {
			err = m.SetJSONGauge(ID, value)
		} else {
			err = DbPostgres.SetDBMetric(MType, ID, value, &deltaDefault)
		}
		if err != nil {
			return err
		}
	case "counter":
		if client.DB == nil {
			err = m.SetJSONCounter(ID, delta)
		} else {
			err = DbPostgres.SetDBMetric(MType, ID, &valueDefault, delta)
		}
		if err != nil {
			return err
		}
	default:
		return err
	}
	return nil
}

func getMetrics(m keeper, MType string, client *db.DB, DbPostgres *pg.PgStorage) interface{} {
	if client.DB == nil {
		return m.GetTypeMetric(MType)
	} else {
		return DbPostgres.GetDBMetrics(MType)
	}
}
