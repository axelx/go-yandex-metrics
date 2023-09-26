// Модуль metrics собирает метрики системы в рантайме и отправляет их по установленному урлу
package metrics

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"go.uber.org/zap"

	"github.com/axelx/go-yandex-metrics/internal/config"
	"github.com/axelx/go-yandex-metrics/internal/hash"
	"github.com/axelx/go-yandex-metrics/internal/models"
	"github.com/axelx/go-yandex-metrics/internal/service"
)

type Metric struct {
	data   []models.Metrics
	conf   config.ConfigAgent
	Logger *zap.Logger
}

// metrics.New новый экземпляр метрик с параметрами конфига и логированием
func New(conf config.ConfigAgent, logger *zap.Logger) Metric {
	return Metric{
		data:   []models.Metrics{},
		conf:   conf,
		Logger: logger,
	}
}

var (
	ErrDialUp = errors.New("dial up connection")
)

// ReportBatch отправка группы метрик
func (m *Metric) ReportBatch() {
	for {

		for _, interval := range m.conf.RetryIntervals {
			err := sendRequestSliceMetrics(m.conf, m.data, m.Logger)
			if err == nil {
				break
			}
			if errors.Is(err, ErrDialUp) {
				time.Sleep(interval)
			}
		}
		time.Sleep(time.Duration(m.conf.ReportFrequency) * time.Second)
	}
}

// Report помещаем метрики в канал jobs
func (m *Metric) Report(jobs chan models.Metrics) {
	for {
		for _, metrics := range m.data {
			jobs <- metrics
		}

		time.Sleep(time.Duration(m.conf.ReportFrequency) * time.Second)
	}
}

// Poll опрос и сбор runtime метрик
func (m *Metric) Poll() {
	me := new(runtime.MemStats)

	mGopsutil, err := mem.VirtualMemory()
	if err != nil {
		m.Logger.Error("Error Poll",
			zap.String("about func", "mGopsutil"),
			zap.String("about ERR", err.Error()),
		)
	}
	pcGopsutil, err := cpu.Percent(time.Duration(m.conf.PollFrequency)*time.Second, true)
	if err != nil {
		m.Logger.Error("Erro Poll",
			zap.String("about func", "pcGopsutil"),
			zap.String("about ERR", err.Error()),
		)
	}

	PollCount := 0
	maxCycle := m.conf.ReportFrequency / m.conf.PollFrequency

	for {
		runtime.ReadMemStats(me)
		m.data = make([]models.Metrics, 0, 33)

		m.data = append(m.data, models.Metrics{ID: "TotalMemory", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mGopsutil.Total))})
		m.data = append(m.data, models.Metrics{ID: "FreeMemory", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mGopsutil.Free))})
		m.data = append(m.data, models.Metrics{ID: "CPUutilization1", MType: "gauge", Value: service.Float64ToPointerFloat64(pcGopsutil[0])})

		m.data = append(m.data, models.Metrics{ID: "RandomValue", MType: "gauge", Value: service.Float64ToPointerFloat64(rand.Float64())})
		m.data = append(m.data, models.Metrics{ID: "PollCount", MType: "counter", Delta: service.Int64ToPointerInt64(int64(PollCount))})
		m.data = append(m.data, models.Metrics{ID: "PollCount2", MType: "counter", Delta: service.Int64ToPointerInt64(int64(PollCount))})
		m.data = append(m.data, models.Metrics{ID: "Alloc", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(me.Alloc))})
		m.data = append(m.data, models.Metrics{ID: "BuckHashSys", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(me.BuckHashSys))})
		m.data = append(m.data, models.Metrics{ID: "Frees", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(me.Frees))})
		m.data = append(m.data, models.Metrics{ID: "GCCPUFraction", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(me.GCCPUFraction))})
		m.data = append(m.data, models.Metrics{ID: "GCSys", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(me.GCSys))})
		m.data = append(m.data, models.Metrics{ID: "HeapAlloc", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(me.HeapAlloc))})
		m.data = append(m.data, models.Metrics{ID: "HeapIdle", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(me.HeapIdle))})
		m.data = append(m.data, models.Metrics{ID: "HeapInuse", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(me.HeapInuse))})
		m.data = append(m.data, models.Metrics{ID: "HeapObjects", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(me.HeapObjects))})
		m.data = append(m.data, models.Metrics{ID: "HeapReleased", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(me.HeapReleased))})
		m.data = append(m.data, models.Metrics{ID: "HeapSys", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(me.HeapSys))})
		m.data = append(m.data, models.Metrics{ID: "LastGC", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(me.LastGC))})
		m.data = append(m.data, models.Metrics{ID: "Lookups", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(me.Lookups))})
		m.data = append(m.data, models.Metrics{ID: "MCacheInuse", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(me.MCacheInuse))})
		m.data = append(m.data, models.Metrics{ID: "MCacheSys", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(me.MCacheSys))})
		m.data = append(m.data, models.Metrics{ID: "MSpanInuse", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(me.MSpanInuse))})
		m.data = append(m.data, models.Metrics{ID: "MSpanSys", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(me.MSpanSys))})
		m.data = append(m.data, models.Metrics{ID: "Mallocs", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(me.Mallocs))})
		m.data = append(m.data, models.Metrics{ID: "NextGC", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(me.NextGC))})
		m.data = append(m.data, models.Metrics{ID: "NumForcedGC", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(me.NumForcedGC))})
		m.data = append(m.data, models.Metrics{ID: "NumGC", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(me.NumGC))})
		m.data = append(m.data, models.Metrics{ID: "StackSys", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(me.StackSys))})
		m.data = append(m.data, models.Metrics{ID: "Sys", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(me.Sys))})
		m.data = append(m.data, models.Metrics{ID: "StackInuse", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(me.StackInuse))})
		m.data = append(m.data, models.Metrics{ID: "PauseTotalNs", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(me.PauseTotalNs))})
		m.data = append(m.data, models.Metrics{ID: "OtherSys", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(me.OtherSys))})
		m.data = append(m.data, models.Metrics{ID: "TotalAlloc", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(me.TotalAlloc))})

		PollCount += 1

		if maxCycle <= PollCount {
			break
		}
		time.Sleep(time.Duration(m.conf.PollFrequency) * time.Second)
	}
}

func sendRequestSliceMetrics(c config.ConfigAgent, metrics []models.Metrics, log *zap.Logger) error {
	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	err = sendRequest("updates/", c, metricsJSON, log)
	if err != nil {
		log.Error("Error sendRequest", zap.String("about ERR", err.Error()))
		return err
	}
	return nil
}

// SendRequestMetric отправка метрик по указанному в кофиге урлу
func SendRequestMetric(c config.ConfigAgent, metric models.Metrics, log *zap.Logger) error {
	metricsJSON, err := json.Marshal(metric)
	if err != nil {
		log.Error("Error SendRequestMetric", zap.String("about ERR", err.Error()))
		return err
	}
	err = sendRequest("update/", c, metricsJSON, log)
	if err != nil {
		return err
	}
	return nil
}

func sendRequest(url string, c config.ConfigAgent, metricsJSON []byte, log *zap.Logger) error {
	buf := bytes.NewBuffer(nil)
	zb := gzip.NewWriter(buf)
	_, err := zb.Write([]byte(metricsJSON))
	if err != nil {
		log.Error("Error zb.Write([]byte(metricsJSON))",
			zap.String("about func", "sendRequest"),
			zap.String("about ERR", err.Error()),
		)
		return err
	}
	zb.Close()

	req, err := http.NewRequest("POST", c.BaseURL+url, buf)
	if err != nil {
		log.Error("Error create request",
			zap.String("about func", "sendRequest"),
			zap.String("about ERR", err.Error()),
		)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	if c.FlagHashKey != "" {
		req.Header.Set("HashSHA256", hash.GetHashSHA256Base64(metricsJSON, c.FlagHashKey))
	}
	resp, err := c.Client.Do(req)

	if err != nil {
		log.Error("Error reporting metrics:",
			zap.String("about func", "sendRequest"),
			zap.String("about metricJSON", string(metricsJSON)),
			zap.String("about ERR", err.Error()),
		)
		return ErrDialUp
	} else {
		resp.Body.Close()
		log.Info("Metrics sent successfully! Send body: %s, Response body: " + string(metricsJSON)) //zap.String("about func", "sendRequest"),
	}
	return nil
}
