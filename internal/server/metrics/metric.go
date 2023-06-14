package metrics

import (
	"fmt"
	"github.com/axelx/go-yandex-metrics/internal/server/config"
	"io"
	"math/rand"
	"runtime"
	"time"
)

type Metric struct {
	data []string
}

func New() Metric {
	return Metric{
		data: []string{},
	}
}

func (m *Metric) Report(c config.ConfigAgent) {
	//производим опрос/обновление метрик
	m.Poll(c)
	for {
		for _, metrics := range m.data {

			resp, err := c.Client.Post(metrics, "text/plain", nil)
			if err != nil {
				fmt.Println("Error reporting metrics:", err)
			} else {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				fmt.Printf("Metrics sent successfully! Response body: %s\n", body)
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

		m.data = append(m.data, urlReportMetric("gauge", "Alloc", float64(mem.Alloc), c.BaseURL))
		m.data = append(m.data, urlReportMetric("gauge", "BuckHashSys", float64(mem.BuckHashSys), c.BaseURL))
		m.data = append(m.data, urlReportMetric("gauge", "Frees", float64(mem.Frees), c.BaseURL))
		m.data = append(m.data, urlReportMetric("gauge", "GCCPUFraction", float64(mem.GCCPUFraction), c.BaseURL))
		m.data = append(m.data, urlReportMetric("gauge", "GCSys", float64(mem.GCSys), c.BaseURL))
		m.data = append(m.data, urlReportMetric("gauge", "HeapAlloc", float64(mem.HeapAlloc), c.BaseURL))
		m.data = append(m.data, urlReportMetric("gauge", "HeapIdle", float64(mem.HeapIdle), c.BaseURL))
		m.data = append(m.data, urlReportMetric("gauge", "HeapInuse", float64(mem.HeapInuse), c.BaseURL))
		m.data = append(m.data, urlReportMetric("gauge", "HeapObjects", float64(mem.HeapObjects), c.BaseURL))
		m.data = append(m.data, urlReportMetric("gauge", "HeapReleased", float64(mem.HeapReleased), c.BaseURL))
		m.data = append(m.data, urlReportMetric("gauge", "HeapSys", float64(mem.HeapSys), c.BaseURL))
		m.data = append(m.data, urlReportMetric("gauge", "LastGC", float64(mem.LastGC), c.BaseURL))
		m.data = append(m.data, urlReportMetric("gauge", "Lookups", float64(mem.Lookups), c.BaseURL))
		m.data = append(m.data, urlReportMetric("gauge", "MCacheInuse", float64(mem.MCacheInuse), c.BaseURL))
		m.data = append(m.data, urlReportMetric("gauge", "MCacheSys", float64(mem.MCacheSys), c.BaseURL))
		m.data = append(m.data, urlReportMetric("gauge", "MSpanInuse", float64(mem.MSpanInuse), c.BaseURL))
		m.data = append(m.data, urlReportMetric("gauge", "MSpanSys", float64(mem.MSpanSys), c.BaseURL))
		m.data = append(m.data, urlReportMetric("gauge", "Mallocs", float64(mem.Mallocs), c.BaseURL))
		m.data = append(m.data, urlReportMetric("gauge", "NextGC", float64(mem.NextGC), c.BaseURL))
		m.data = append(m.data, urlReportMetric("gauge", "NumForcedGC", float64(mem.NumForcedGC), c.BaseURL))
		m.data = append(m.data, urlReportMetric("gauge", "NumGC", float64(mem.NumGC), c.BaseURL))

		m.data = append(m.data, urlReportMetric("gauge", "RandomValue", rand.Float64(), c.BaseURL))
		m.data = append(m.data, urlReportMetric("counter", "PollCount", float64(PollCount), c.BaseURL))

		PollCount += 1

		if maxCycle <= PollCount {
			break
		}
		time.Sleep(time.Duration(c.PollFrequency) * time.Second)
	}
}
func urlReportMetric(metricType string, metricName string, metricValue float64, baseURL string) string {
	if metricType == "gauge" {
		return fmt.Sprintf("%s%s/%s/%f", baseURL, metricType, metricName, metricValue)
	} else if metricType == "counter" {
		i := int(metricValue)
		return fmt.Sprintf("%s%s/%s/%v", baseURL, metricType, metricName, i)
	} else {
		return "ошибка в создании отчета по метрики"
	}
}
