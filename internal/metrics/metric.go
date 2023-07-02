package metrics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/axelx/go-yandex-metrics/internal/config"
	"github.com/axelx/go-yandex-metrics/internal/models"
	"io"
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
		m.Poll(c)
		for _, metrics := range m.data {

			metricsJSON, err := json.Marshal(metrics)
			if err != nil {
				fmt.Printf("Error metricsJSON: %s\n", err)
			}

			buf := bytes.NewBuffer(metricsJSON)
			//buf := bytes.NewBuffer(nil)
			//zb := gzip.NewWriter(buf)
			//_, err = zb.Write([]byte(metricsJSON))
			//zb.Close()

			req, _ := http.NewRequest("POST", c.BaseURL, buf)
			req.Header.Set("Content-Type", "application/json")
			//req.Header.Set("Content-Encoding", "gzip")
			resp, _ := c.Client.Do(req)

			if err != nil {
				fmt.Println("Error reporting metrics:", err, string(metricsJSON))
			} else {
				body, _ := io.ReadAll(resp.Body)
				defer resp.Body.Close()

				fmt.Printf("Metrics sent successfully! Send body: %s, Response body: %s\n", string(metricsJSON), body)
			}
		}

		time.Sleep(time.Duration(c.ReportFrequency) * time.Second)
	}
}

func (m *Metric) Poll(c config.ConfigAgent) {
	mem := new(runtime.MemStats)
	PollCount := 0
	maxCycle := c.ReportFrequency / c.PollFrequency

	for {
		runtime.ReadMemStats(mem)
		m.data = []models.Metrics{}
		r := rand.Float64()
		m.data = append(m.data, models.Metrics{ID: "RandomValue", MType: "gauge", Value: &r})
		i := int64(PollCount)
		m.data = append(m.data, models.Metrics{ID: "PollCount", MType: "counter", Delta: &i})
		mf1 := float64(mem.Alloc)
		m.data = append(m.data, models.Metrics{ID: "Alloc", MType: "gauge", Value: &mf1})
		mf2 := float64(mem.BuckHashSys)
		m.data = append(m.data, models.Metrics{ID: "BuckHashSys", MType: "gauge", Value: &mf2})
		mf3 := float64(mem.Frees)
		m.data = append(m.data, models.Metrics{ID: "Frees", MType: "gauge", Value: &mf3})
		mf4 := float64(mem.GCCPUFraction)
		m.data = append(m.data, models.Metrics{ID: "GCCPUFraction", MType: "gauge", Value: &mf4})
		mf5 := float64(mem.GCSys)
		m.data = append(m.data, models.Metrics{ID: "GCSys", MType: "gauge", Value: &mf5})
		mf6 := float64(mem.HeapAlloc)
		m.data = append(m.data, models.Metrics{ID: "HeapAlloc", MType: "gauge", Value: &mf6})
		mf7 := float64(mem.HeapIdle)
		m.data = append(m.data, models.Metrics{ID: "HeapIdle", MType: "gauge", Value: &mf7})
		mf8 := float64(mem.HeapInuse)
		m.data = append(m.data, models.Metrics{ID: "HeapInuse", MType: "gauge", Value: &mf8})
		mf9 := float64(mem.HeapObjects)
		m.data = append(m.data, models.Metrics{ID: "HeapObjects", MType: "gauge", Value: &mf9})
		mf10 := float64(mem.HeapReleased)
		m.data = append(m.data, models.Metrics{ID: "HeapReleased", MType: "gauge", Value: &mf10})
		mf11 := float64(mem.HeapSys)
		m.data = append(m.data, models.Metrics{ID: "HeapSys", MType: "gauge", Value: &mf11})
		mf12 := float64(mem.LastGC)
		m.data = append(m.data, models.Metrics{ID: "LastGC", MType: "gauge", Value: &mf12})
		mf13 := float64(mem.Lookups)
		m.data = append(m.data, models.Metrics{ID: "Lookups", MType: "gauge", Value: &mf13})
		mf14 := float64(mem.MCacheInuse)
		m.data = append(m.data, models.Metrics{ID: "MCacheInuse", MType: "gauge", Value: &mf14})
		mf15 := float64(mem.MCacheSys)
		m.data = append(m.data, models.Metrics{ID: "MCacheSys", MType: "gauge", Value: &mf15})
		mf16 := float64(mem.MSpanInuse)
		m.data = append(m.data, models.Metrics{ID: "MSpanInuse", MType: "gauge", Value: &mf16})
		mf17 := float64(mem.MSpanSys)
		m.data = append(m.data, models.Metrics{ID: "MSpanSys", MType: "gauge", Value: &mf17})
		mf18 := float64(mem.Mallocs)
		m.data = append(m.data, models.Metrics{ID: "Mallocs", MType: "gauge", Value: &mf18})
		mf19 := float64(mem.NextGC)
		m.data = append(m.data, models.Metrics{ID: "NextGC", MType: "gauge", Value: &mf19})
		mf20 := float64(mem.NumForcedGC)
		m.data = append(m.data, models.Metrics{ID: "NumForcedGC", MType: "gauge", Value: &mf20})
		mf21 := float64(mem.NumGC)
		m.data = append(m.data, models.Metrics{ID: "NumGC", MType: "gauge", Value: &mf21})
		mf22 := float64(mem.StackSys)
		m.data = append(m.data, models.Metrics{ID: "StackSys", MType: "gauge", Value: &mf22})
		mf23 := float64(mem.Sys)
		m.data = append(m.data, models.Metrics{ID: "Sys", MType: "gauge", Value: &mf23})
		mf24 := float64(mem.StackInuse)
		m.data = append(m.data, models.Metrics{ID: "StackInuse", MType: "gauge", Value: &mf24})
		mf25 := float64(mem.PauseTotalNs)
		m.data = append(m.data, models.Metrics{ID: "PauseTotalNs", MType: "gauge", Value: &mf25})
		mf26 := float64(mem.OtherSys)
		m.data = append(m.data, models.Metrics{ID: "OtherSys", MType: "gauge", Value: &mf26})
		mf27 := float64(mem.TotalAlloc)
		m.data = append(m.data, models.Metrics{ID: "TotalAlloc", MType: "gauge", Value: &mf27})

		PollCount += 1

		if maxCycle <= PollCount {
			break
		}
		time.Sleep(time.Duration(c.PollFrequency) * time.Second)
	}
}
