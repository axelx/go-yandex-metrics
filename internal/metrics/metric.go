package metrics

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/axelx/go-yandex-metrics/internal/config"
	"github.com/axelx/go-yandex-metrics/internal/models"
	"github.com/axelx/go-yandex-metrics/internal/service"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

type Metric struct {
	data []models.Metrics
}

func New() Metric {
	return Metric{
		data: []models.Metrics{},
	}
}

func (m *Metric) Report(c config.ConfigAgent) {
	for {
		//производим опрос/обновление метрик
		m.poll(c)
		fmt.Println("metrics:///", m.data)
		metricsJSON, err := json.Marshal(m.data)
		if err != nil {
			fmt.Printf("Error metricsJSON: %s\n", err)
		}
		fmt.Println("[]metrics:///", string(metricsJSON))

		sendRequestSliceMetrics(c, m.data)
		for _, metrics := range m.data {
			sendRequestMetric(c, metrics)
		}

		time.Sleep(time.Duration(c.ReportFrequency) * time.Second)
	}
}

func (m *Metric) poll(c config.ConfigAgent) {
	mem := new(runtime.MemStats)
	PollCount := 0
	maxCycle := c.ReportFrequency / c.PollFrequency

	for {
		runtime.ReadMemStats(mem)
		m.data = []models.Metrics{}
		m.data = append(m.data, models.Metrics{ID: "RandomValue", MType: "gauge", Value: service.Float64ToPointerFloat64(rand.Float64())})
		m.data = append(m.data, models.Metrics{ID: "PollCount", MType: "counter", Delta: service.Int64ToPointerInt64(int64(PollCount))})
		m.data = append(m.data, models.Metrics{ID: "Alloc", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mem.Alloc))})
		m.data = append(m.data, models.Metrics{ID: "BuckHashSys", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mem.BuckHashSys))})
		m.data = append(m.data, models.Metrics{ID: "Frees", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mem.Frees))})
		m.data = append(m.data, models.Metrics{ID: "GCCPUFraction", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mem.GCCPUFraction))})
		m.data = append(m.data, models.Metrics{ID: "GCSys", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mem.GCSys))})
		m.data = append(m.data, models.Metrics{ID: "HeapAlloc", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mem.HeapAlloc))})
		m.data = append(m.data, models.Metrics{ID: "HeapIdle", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mem.HeapIdle))})
		m.data = append(m.data, models.Metrics{ID: "HeapInuse", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mem.HeapInuse))})
		m.data = append(m.data, models.Metrics{ID: "HeapObjects", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mem.HeapObjects))})
		m.data = append(m.data, models.Metrics{ID: "HeapReleased", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mem.HeapReleased))})
		m.data = append(m.data, models.Metrics{ID: "HeapSys", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mem.HeapSys))})
		m.data = append(m.data, models.Metrics{ID: "LastGC", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mem.LastGC))})
		m.data = append(m.data, models.Metrics{ID: "Lookups", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mem.Lookups))})
		m.data = append(m.data, models.Metrics{ID: "MCacheInuse", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mem.MCacheInuse))})
		m.data = append(m.data, models.Metrics{ID: "MCacheSys", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mem.MCacheSys))})
		m.data = append(m.data, models.Metrics{ID: "MSpanInuse", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mem.MSpanInuse))})
		m.data = append(m.data, models.Metrics{ID: "MSpanSys", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mem.MSpanSys))})
		m.data = append(m.data, models.Metrics{ID: "Mallocs", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mem.Mallocs))})
		m.data = append(m.data, models.Metrics{ID: "NextGC", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mem.NextGC))})
		m.data = append(m.data, models.Metrics{ID: "NumForcedGC", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mem.NumForcedGC))})
		m.data = append(m.data, models.Metrics{ID: "NumGC", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mem.NumGC))})
		m.data = append(m.data, models.Metrics{ID: "StackSys", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mem.StackSys))})
		m.data = append(m.data, models.Metrics{ID: "Sys", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mem.Sys))})
		m.data = append(m.data, models.Metrics{ID: "StackInuse", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mem.StackInuse))})
		m.data = append(m.data, models.Metrics{ID: "PauseTotalNs", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mem.PauseTotalNs))})
		m.data = append(m.data, models.Metrics{ID: "OtherSys", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mem.OtherSys))})
		m.data = append(m.data, models.Metrics{ID: "TotalAlloc", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(mem.TotalAlloc))})

		PollCount += 1

		if maxCycle <= PollCount {
			break
		}
		time.Sleep(time.Duration(c.PollFrequency) * time.Second)
	}
}

func sendRequestSliceMetrics(c config.ConfigAgent, metrics []models.Metrics) {
	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		fmt.Printf("Error metricsJSON: %s\n", err)
	}
	sendRequest("updates/", c, metricsJSON)
}
func sendRequestMetric(c config.ConfigAgent, metric models.Metrics) {
	metricsJSON, err := json.Marshal(metric)
	if err != nil {
		fmt.Printf("Error metricsJSON: %s\n", err)
	}
	sendRequest("update/", c, metricsJSON)
}

func sendRequest(url string, c config.ConfigAgent, metricsJSON []byte) {
	buf := bytes.NewBuffer(nil)
	zb := gzip.NewWriter(buf)
	_, err := zb.Write([]byte(metricsJSON))
	if err != nil {
		fmt.Println("Error", err)
	}
	zb.Close()

	req, err := http.NewRequest("POST", c.BaseURL+url, buf)
	if err != nil {
		fmt.Println("Error reporting metrics:", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	resp, err := c.Client.Do(req)

	if err != nil {
		fmt.Println("Error reporting metrics:", err, string(metricsJSON))
	} else {
		//body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		fmt.Printf("Metrics sent successfully! Send body: %s, Response body: \n", string(metricsJSON))
	}
}
