package main

import (
	"fmt"
	"github.com/axelx/go-yandex-metrics/internal/config"
	"github.com/axelx/go-yandex-metrics/internal/metrics"
	"github.com/axelx/go-yandex-metrics/internal/models"
	"sync"
)

func main() {
	conf := config.NewConfigAgent()

	fmt.Println("Running server on", conf.BaseURL, conf.ReportFrequency, conf.PollFrequency)

	metric := metrics.New(conf)

	var wg sync.WaitGroup
	//производим опрос/обновление метрик
	wg.Add(1)
	go func() {
		for {
			metric.Poll()
		}
	}()
	//производим отправку пачкой метрики
	wg.Add(1)
	go func() {
		for {
			metric.ReportBatch()
		}
	}()

	jobs := make(chan models.Metrics, 30)

	//добавляем метрики для отправки в канал
	go func() {
		metric.Report(jobs)
	}()

	// создаем воркеры которые будут отправлять метрики из канала jobs
	for w := 1; w <= conf.FlagRateLimit; w++ {
		go worker(w, jobs, conf)
	}
	wg.Wait()
}

func worker(id int, jobs <-chan models.Metrics, c config.ConfigAgent) {
	for job := range jobs {
		fmt.Println("рабочий", id, "Start запущен задача", job)
		metrics.SendRequestMetric(c, job)
	}
}
