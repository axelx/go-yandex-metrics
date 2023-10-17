package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"github.com/axelx/go-yandex-metrics/internal/crypto"
	"github.com/axelx/go-yandex-metrics/internal/hash"
	"github.com/axelx/go-yandex-metrics/internal/logger"
	"github.com/axelx/go-yandex-metrics/internal/mgzip"
	"github.com/axelx/go-yandex-metrics/internal/models"
	"github.com/axelx/go-yandex-metrics/internal/mtemplate"
	"github.com/axelx/go-yandex-metrics/internal/pg"
	"github.com/axelx/go-yandex-metrics/internal/service"
)

type keeper interface {
	SetGauge(string, float64) error
	SetCounter(string, int64) error
	GetMetric(models.MetricType, string) (string, error)

	SetJSONGauge(string, *float64) error
	SetJSONCounter(string, *int64) error
	GetJSONMetric(models.MetricType, string) (models.Metrics, error)

	GetTypeMetric(models.MetricType) interface{}
}

type handler struct {
	memStorage keeper
	DB         *sqlx.DB
	DBPostgres *pg.PgStorage
	HashKey    string
	CryptoKey  string
}

// handler.New создаем новый обработчик
func New(k keeper, db *sqlx.DB, NewDBStorage *pg.PgStorage, hashKey, cryptoKey string) handler {
	return handler{
		memStorage: k,
		DB:         db,
		DBPostgres: NewDBStorage,
		HashKey:    hashKey,
		CryptoKey:  cryptoKey,
	}
}

func (h *handler) Router() chi.Router {

	r := chi.NewRouter()
	r.Use(logger.RequestLogger())

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
		if h.DB != nil {
			res.WriteHeader(http.StatusOK)
			_, err := res.Write([]byte(""))
			if err != nil {
				logger.Log.Error("Error res.Write(metricsJSON)", "DbConnect err: "+err.Error())
			}
		} else {
			res.WriteHeader(http.StatusInternalServerError)
			_, err := res.Write([]byte(""))
			if err != nil {
				logger.Log.Error("Error res.Write(metricsJSON)", "DbConnect err: "+err.Error())
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
		tM := chi.URLParam(req, "typeM")
		typeM := models.MetricType(tM)
		nameM := chi.URLParam(req, "nameM")

		valueM := chi.URLParam(req, "valueM")

		if typeM == "" || nameM == "" || valueM == "" {
			http.Error(res, "StatusNotFound", http.StatusNotFound)
			return
		}
		switch typeM {
		case models.MetricGauge:
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
		case models.MetricCounter:
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
			logger.Log.Error("Error res.Write(metricsJSON)", "UpdatedMetric err: "+err.Error())
		}

		res.WriteHeader(http.StatusOK)

		logger.Log.Info("sending HTTP response UpdateMetrics",
			"status: 200"+"size: "+strconv.Itoa(size))
	}
}

func (h *handler) GetMetric() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		tM := chi.URLParam(req, "typeM")
		typeM := models.MetricType(tM)

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
			logger.Log.Error("Error res.Write(metricsJSON)", "GetMetric err: "+err.Error())
		}

		res.WriteHeader(http.StatusOK)

		logger.Log.Info("sending HTTP response UpdatedMetric",
			"status: 200"+"size: "+strconv.Itoa(size))
	}
}

func (h *handler) UpdatedJSONMetric() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var metrics models.Metrics

		if h.CryptoKey == "" {
			dec := json.NewDecoder(req.Body)
			err := dec.Decode(&metrics)
			if err != nil {
				logger.Log.Error("cannot decode request JSON body", err.Error())
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			dataDecode := crypto.Decode(service.StreamToByte(req.Body), h.CryptoKey)
			err := json.Unmarshal(dataDecode, &metrics)
			if err != nil {
				logger.Log.Error("Cannot decode request JSON body with crypto private key", err.Error())
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		var deltaDefault int64
		var valueDefault float64
		if metrics.MType == "" || metrics.ID == "" || (metrics.Delta == &deltaDefault && metrics.Value == &valueDefault) {
			http.Error(res, "StatusNotFound", http.StatusNotFound)
			return
		}

		err := h.setJSONorDBmetric(h.memStorage, metrics.MType, metrics.ID, metrics.Value, metrics.Delta, h.DB, h.DBPostgres)
		if err != nil {
			http.Error(res, fmt.Sprint(err), http.StatusBadRequest)
			return
		}

		metricStorage, err := h.getJSONorDBmetrics(h.memStorage, metrics.MType, metrics.ID, h.DB, h.DBPostgres)
		if err != nil {
			http.Error(res, "StatusNotFound", http.StatusNotFound)
			return
		}

		metricsJSON, err := json.Marshal(metricStorage)
		if err != nil {
			logger.Log.Error("Error json.Marshal(metricStorage)", "UpdatedJSONMetric err: "+err.Error())
		}
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		size, err := res.Write(metricsJSON)
		if err != nil {
			logger.Log.Error("Error res.Write(metricsJSON)", "UpdatedJSONMetric err: "+err.Error())
		}

		logger.Log.Info("sending HTTP response UpdatedMetric",
			"status: 200"+"size: "+strconv.Itoa(size))
	}
}
func (h *handler) UpdatedJSONMetrics() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var metrics []models.Metrics

		if h.CryptoKey == "" {
			dec := json.NewDecoder(req.Body)
			err := dec.Decode(&metrics)
			if err != nil {
				logger.Log.Error("UpdatedJSONMetrics: Cannot decode request JSON body", err.Error())
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			dataDecode := crypto.Decode(service.StreamToByte(req.Body), h.CryptoKey)
			err := json.Unmarshal(dataDecode, &metrics)
			if err != nil {
				logger.Log.Error("UpdatedJSONMetrics: Cannot decode request JSON body with crypto private key", err.Error())
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		if h.HashKey != "" {
			hashHeader := req.Header.Get("HashSHA256")
			metricsJSON, _ := json.Marshal(metrics)
			if !h.checkHash(h.HashKey, hashHeader, metricsJSON) {
				http.Error(res, "StatusNotFound", http.StatusNotFound)
				return
			}
		}

		if len(metrics) == 0 {
			http.Error(res, "StatusNotFound", http.StatusNotFound)
			return
		}

		err := h.DBPostgres.SetBatchMetrics(metrics)
		if err != nil {
			http.Error(res, fmt.Sprint(err), http.StatusBadRequest)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		if h.HashKey != "" {
			req.Header.Set("HashSHA256", hash.GetHashSHA256Base64([]byte("{}"), h.HashKey))
		}
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("{}"))

		logger.Log.Info("sending HTTP response UpdatedMetric",
			"status: 200")
	}
}

func (h *handler) GetJSONMetric() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		var metrics models.Metrics
		dec := json.NewDecoder(req.Body)
		if err := dec.Decode(&metrics); err != nil {
			logger.Log.Debug("cannot decode request JSON body", err.Error())
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		if metrics.MType == "" || metrics.ID == "" {
			http.Error(res, "StatusNotFound", http.StatusNotFound)
			return
		}

		metric, err := h.getJSONorDBmetrics(h.memStorage, metrics.MType, metrics.ID, h.DB, h.DBPostgres)
		if err != nil {
			http.Error(res, "StatusNotFound", http.StatusNotFound)
			return
		}

		metricsJSON, err := json.Marshal(metric)
		if err != nil {
			logger.Log.Error("Error json marshal metrics", "GetJSONMetric err: "+err.Error())
		}
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		size, _ := res.Write(metricsJSON)

		logger.Log.Info("sending HTTP response UpdatedMetric",
			"status: 200"+"size: "+strconv.Itoa(size))
		fmt.Println(size)
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
			Gauge:   h.getMetrics(h.memStorage, models.MetricGauge, h.DB, h.DBPostgres),
			Counter: h.getMetrics(h.memStorage, models.MetricCounter, h.DB, h.DBPostgres),
		})
		tmplSize := len(buf.Bytes())

		logger.Log.Info("sending HTTP response GetAllMetrics",
			"status: 200"+"size: "+strconv.Itoa(tmplSize))
	}
}

func (h *handler) getJSONorDBmetrics(m keeper, MType models.MetricType, ID string, DB *sqlx.DB, DBPostgres *pg.PgStorage) (models.Metrics, error) {
	metricStorage := models.Metrics{}
	var err error = nil
	if DB == nil {
		metricStorage, err = m.GetJSONMetric(MType, ID)
	} else {
		metricStorage, err = DBPostgres.GetDBMetric(MType, ID)
	}
	if err != nil {
		return models.Metrics{}, err
	}
	return metricStorage, nil
}

func (h *handler) setJSONorDBmetric(m keeper, MType models.MetricType, ID string, value *float64, delta *int64, DB *sqlx.DB, DBPostgres *pg.PgStorage) error {
	var err error = nil

	deltaDefault := int64(0)
	valueDefault := float64(0)

	switch MType {
	case models.MetricGauge:
		if DB == nil {
			err = m.SetJSONGauge(ID, value)
		} else {
			err = DBPostgres.SetDBMetric(MType, ID, value, &deltaDefault)
		}
		if err != nil {
			return err
		}
	case models.MetricCounter:
		if DB == nil {
			err = m.SetJSONCounter(ID, delta)
		} else {
			err = DBPostgres.SetDBMetric(MType, ID, &valueDefault, delta)
		}
		if err != nil {
			return err
		}
	default:
		return err
	}
	return nil
}

func (h *handler) getMetrics(m keeper, MType models.MetricType, DB *sqlx.DB, DBPostgres *pg.PgStorage) interface{} {
	if DB == nil {
		return m.GetTypeMetric(MType)
	} else {
		return DBPostgres.GetDBMetrics(MType)
	}
}

func (h *handler) checkHash(key, hashHeader string, data []byte) bool {
	ha := hash.GetHashSHA256Base64(data, key)

	logger.Log.Info("checkHash, hashHeader",
		"calculated hash: "+ha+"key: "+key)

	return hashHeader == ha
}
