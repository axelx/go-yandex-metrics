package metrics

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/axelx/go-yandex-metrics/internal/config"
	"github.com/axelx/go-yandex-metrics/internal/hash"
	"github.com/axelx/go-yandex-metrics/internal/models"
	"github.com/axelx/go-yandex-metrics/internal/service"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

type Metric struct {
	data []models.Metrics
	conf config.ConfigAgent
}

func New(conf config.ConfigAgent) Metric {
	return Metric{
		data: []models.Metrics{},
		conf: conf,
	}
}

var (
	ErrDialUp = errors.New("dial up connection")
)

func (m *Metric) ReportBatch() {
	for {

		for _, interval := range m.conf.RetryIntervals {
			err := sendRequestSliceMetrics(m.conf, m.data)
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
func (m *Metric) Report(jobs chan models.Metrics) {
	for {
		for _, metrics := range m.data {
			jobs <- metrics
		}

		time.Sleep(time.Duration(m.conf.ReportFrequency) * time.Second)
	}
}

func (m *Metric) CreateJobs(jobs chan models.Metrics) {
	for _, metrics := range m.data {
		jobs <- metrics
	}
}

func (m *Metric) Poll() {
	me := new(runtime.MemStats)

	m_gopsutil, err := mem.VirtualMemory()
	if err != nil {
		fmt.Println(err)
	}
	pc_gopsutil, err := cpu.Percent(time.Duration(m.conf.PollFrequency)*time.Second, true)
	if err != nil {
		fmt.Println(err)
	}

	PollCount := 0
	maxCycle := m.conf.ReportFrequency / m.conf.PollFrequency

	for {
		runtime.ReadMemStats(me)
		m.data = []models.Metrics{}
		m.data = append(m.data, models.Metrics{ID: "TotalMemory", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(m_gopsutil.Total))})
		m.data = append(m.data, models.Metrics{ID: "FreeMemory", MType: "gauge", Value: service.Float64ToPointerFloat64(float64(m_gopsutil.Free))})
		m.data = append(m.data, models.Metrics{ID: "CPUutilization1", MType: "gauge", Value: service.Float64ToPointerFloat64(pc_gopsutil[0])})

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

func sendRequestSliceMetrics(c config.ConfigAgent, metrics []models.Metrics) error {
	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	err = sendRequest("updates/", c, metricsJSON)
	if err != nil {
		fmt.Printf("Error sendRequest: %s\n", err)
		return err
	}
	return nil
}
func SendRequestMetric(c config.ConfigAgent, metric models.Metrics) error {
	metricsJSON, err := json.Marshal(metric)
	if err != nil {
		fmt.Printf("Error metricsJSON: %s\n", err)
		return err
	}
	err = sendRequest("update/", c, metricsJSON)
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
		fmt.Println("Error zb.Write([]byte(metricsJSON)): ", err)
		return err
	}
	zb.Close()

	req, err := http.NewRequest("POST", c.BaseURL+url, buf)
	if err != nil {
		fmt.Println("Error create request:", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	if c.FlagHashKey != "" {
		req.Header.Set("HashSHA256", hash.GetHashSHA256Base64(metricsJSON, c.FlagHashKey))
	}
	resp, err := c.Client.Do(req)

	if err != nil {
		fmt.Println("Error reporting metrics:", err, string(metricsJSON))
		return ErrDialUp
	} else {
		resp.Body.Close()

		fmt.Printf("Metrics sent successfully! Send body: %s, Response body: \n", string(metricsJSON))
	}
	return nil
}
