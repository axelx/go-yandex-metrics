// Модуль metrics собирает метрики системы в рантайме и отправляет их по установленному урлу
package metrics

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/axelx/go-yandex-metrics/internal/config"
	"github.com/axelx/go-yandex-metrics/internal/crypto"
	"github.com/axelx/go-yandex-metrics/internal/hash"
	"github.com/axelx/go-yandex-metrics/internal/logger"
	"github.com/axelx/go-yandex-metrics/internal/models"
	"github.com/axelx/go-yandex-metrics/internal/service"
)

type Metric struct {
	data []models.Metrics
	conf config.ConfigAgent
}

// metrics.New новый экземпляр метрик с параметрами конфига и логированием
func New(conf config.ConfigAgent) Metric {
	return Metric{
		data: []models.Metrics{},
		conf: conf,
	}
}

var (
	ErrDialUp = errors.New("dial up connection")
)

// ReportBatch отправка группы метрик
func (m *Metric) ReportBatch(ctx context.Context) {
	for {
		for _, interval := range m.conf.RetryIntervals {
			select {
			case <-ctx.Done():
				return
			default:
				err := sendRequestSliceMetrics(m.conf, m.data)
				if err == nil {
					break
				}
				if errors.Is(err, ErrDialUp) {
					time.Sleep(interval)
				}
			}
		}
		time.Sleep(time.Duration(m.conf.ReportFrequency) * time.Second)
	}
}

// Report помещаем метрики в канал jobs
func (m *Metric) Report(ctx context.Context, jobs chan models.Metrics) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			for _, metrics := range m.data {
				jobs <- metrics
			}
		}
		time.Sleep(time.Duration(m.conf.ReportFrequency) * time.Second)
	}
}

// Poll опрос и сбор runtime метрик
func (m *Metric) Poll() {
	me := new(runtime.MemStats)

	mGopsutil, err := mem.VirtualMemory()
	if err != nil {
		logger.Error("Error Poll", "mGopsutil err: "+err.Error())
	}
	pcGopsutil, err := cpu.Percent(time.Duration(m.conf.PollFrequency)*time.Second, true)
	if err != nil {
		logger.Error("Error Poll", "pcGopsutil err: "+err.Error())
	}

	PollCount := 0
	maxCycle := m.conf.ReportFrequency / m.conf.PollFrequency

	for {
		runtime.ReadMemStats(me)
		m.data = make([]models.Metrics, 0, 33)

		m.getMetrics(me, PollCount, pcGopsutil[0], float64(mGopsutil.Total), float64(mGopsutil.Free))

		PollCount += 1

		if maxCycle <= PollCount {
			break
		}
		time.Sleep(time.Duration(m.conf.PollFrequency) * time.Second)
	}
}

func (m *Metric) getMetrics(me *runtime.MemStats, PollCount int, pcGopsutil, mGopsutilT, mGopsutilF float64) {
	m.data = append(m.data, models.Metrics{ID: "TotalMemory", MType: "gauge", Value: service.Float64ToPointerFloat64(mGopsutilT)})
	m.data = append(m.data, models.Metrics{ID: "FreeMemory", MType: "gauge", Value: service.Float64ToPointerFloat64(mGopsutilF)})
	m.data = append(m.data, models.Metrics{ID: "CPUutilization1", MType: "gauge", Value: service.Float64ToPointerFloat64(pcGopsutil)})

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
}

func sendRequestSliceMetrics(c config.ConfigAgent, metrics []models.Metrics) error {
	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	if c.CryptoKey != "" {
		metricsJSON, err = crypto.EncodeRSAAES(metricsJSON, c.CryptoKey)
		if err != nil {
			return err
		}
	}
	err = sendRequest("updates/", c, metricsJSON)
	if err != nil {
		logger.Error("Error sendRequest", "about err: "+err.Error())
		return err
	}
	return nil
}

// SendRequestMetric отправка метрик по указанному в кофиге урлу
func SendRequestMetric(c config.ConfigAgent, metric models.Metrics) error {
	metricJSON, err := json.Marshal(metric)
	if err != nil {
		logger.Error("Error SendRequestMetric", "about err: "+err.Error())
		return err
	}
	if c.CryptoKey != "" {
		metricJSON, err = crypto.EncodeRSAAES(metricJSON, c.CryptoKey)
		if err != nil {
			return err
		}
	}

	err = sendRequest("update/", c, metricJSON)
	if err != nil {
		return err
	}
	return nil
}

func sendRequest(url string, c config.ConfigAgent, metricsJSON []byte) error {
	buf := bytes.NewBuffer(nil)
	zb := gzip.NewWriter(buf)
	_, err := zb.Write([]byte(metricsJSON))
	if err != nil {
		logger.Error("Error zb.Write([]byte(metricsJSON)", "sendRequest; about err: "+err.Error())
		return err
	}
	zb.Close()

	req, err := http.NewRequest("POST", c.BaseURL+url, buf)
	if err != nil {
		logger.Error("Error create request", "sendRequest; about err: "+err.Error())
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	if c.HashKey != "" {
		req.Header.Set("HashSHA256", hash.GetHashSHA256Base64(metricsJSON, c.HashKey))
	}
	resp, err := c.Client.Do(req)
	resp.Body.Close()

	if err != nil {
		logger.Error("Error reporting metrics:", "metricJSON; about err: "+err.Error()+
			"about metricJSON: "+string(metricsJSON))
		return ErrDialUp
	}
	return nil
}
