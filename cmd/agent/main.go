package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

var (
	client          = &http.Client{}
	baseURL         = ""
	reportFrequency = 10
	pollFrequency   = 2
)

func main() {
	// обрабатываем аргументы командной строки
	parseFlags()
	baseURL = "http://localhost" + flagRunAddr + "/update/"
	reportFrequency = flagReportFrequency
	pollFrequency = flagPollFrequency

	fmt.Println("Running server on", baseURL, reportFrequency, pollFrequency)

	var metrics []string

	go pollMetrics(&metrics)
	go reportMetrics(&metrics)
	select {}
}

func reportMetrics(metrics *[]string) {
	for {
		for _, metric := range *metrics {
			resp, err := client.Post(metric, "text/plain", nil)
			if err != nil {
				fmt.Println("Error reporting metrics:", err)
			} else {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				fmt.Printf("Metrics sent successfully! Response body: %s\n", body)
			}
		}

		fmt.Println("readMetrics ", *metrics)
		time.Sleep(time.Duration(reportFrequency) * time.Second)
	}
}

func pollMetrics(metrics *[]string) {
	mem := new(runtime.MemStats)
	PollCount := 0.00

	for {
		runtime.ReadMemStats(mem)
		*metrics = []string{}

		*metrics = append(*metrics, urlReportMetric("gauge", "Alloc", float64(mem.Alloc)))
		*metrics = append(*metrics, urlReportMetric("gauge", "BuckHashSys", float64(mem.BuckHashSys)))
		*metrics = append(*metrics, urlReportMetric("gauge", "Frees", float64(mem.Frees)))
		*metrics = append(*metrics, urlReportMetric("gauge", "GCCPUFraction", float64(mem.GCCPUFraction)))
		*metrics = append(*metrics, urlReportMetric("gauge", "GCSys", float64(mem.GCSys)))
		*metrics = append(*metrics, urlReportMetric("gauge", "HeapAlloc", float64(mem.HeapAlloc)))
		*metrics = append(*metrics, urlReportMetric("gauge", "HeapIdle", float64(mem.HeapIdle)))
		*metrics = append(*metrics, urlReportMetric("gauge", "HeapInuse", float64(mem.HeapInuse)))
		*metrics = append(*metrics, urlReportMetric("gauge", "HeapObjects", float64(mem.HeapObjects)))
		*metrics = append(*metrics, urlReportMetric("gauge", "HeapReleased", float64(mem.HeapReleased)))
		*metrics = append(*metrics, urlReportMetric("gauge", "HeapSys", float64(mem.HeapSys)))
		*metrics = append(*metrics, urlReportMetric("gauge", "LastGC", float64(mem.LastGC)))
		*metrics = append(*metrics, urlReportMetric("gauge", "Lookups", float64(mem.Lookups)))
		*metrics = append(*metrics, urlReportMetric("gauge", "MCacheInuse", float64(mem.MCacheInuse)))
		*metrics = append(*metrics, urlReportMetric("gauge", "MCacheSys", float64(mem.MCacheSys)))
		*metrics = append(*metrics, urlReportMetric("gauge", "MSpanInuse", float64(mem.MSpanInuse)))
		*metrics = append(*metrics, urlReportMetric("gauge", "MSpanSys", float64(mem.MSpanSys)))
		*metrics = append(*metrics, urlReportMetric("gauge", "Mallocs", float64(mem.Mallocs)))
		*metrics = append(*metrics, urlReportMetric("gauge", "NextGC", float64(mem.NextGC)))
		*metrics = append(*metrics, urlReportMetric("gauge", "NumForcedGC", float64(mem.NumForcedGC)))
		*metrics = append(*metrics, urlReportMetric("gauge", "NumGC", float64(mem.NumGC)))

		*metrics = append(*metrics, urlReportMetric("gauge", "RandomValue", rand.Float64()))
		*metrics = append(*metrics, urlReportMetric("counter", "PollCount", PollCount))

		PollCount += 1
		time.Sleep(time.Duration(pollFrequency) * time.Second)
	}
}

func urlReportMetric(metricType string, metricName string, metricValue float64) string {
	if metricType == "gauge" {
		return fmt.Sprintf("%s%s/%s/%f", baseURL, metricType, metricName, metricValue)
	} else if metricType == "counter" {
		i := int(metricValue)
		return fmt.Sprintf("%s%s/%s/%v", baseURL, metricType, metricName, i)
	} else {
		return "ошибка в создании отчета по метрики"
	}
}
