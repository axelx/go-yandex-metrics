package main

import (
	"fmt"
	"github.com/axelx/go-yandex-metrics/internal/config"
	"github.com/axelx/go-yandex-metrics/internal/logger"
	"github.com/axelx/go-yandex-metrics/internal/metrics"
	"github.com/axelx/go-yandex-metrics/internal/models"
	"go.uber.org/zap"
	"sync"
)

func main() {
	conf := config.NewConfigAgent()

	lg := logger.Initialize("info")
	lg.Info("Running server", zap.String("config", conf.String()))

	metric := metrics.New(conf, lg)

	var wg sync.WaitGroup
	//производим опрос/обновление метрик
	wg.Add(1)
	go func() {
		for {
			metric.Poll()
		}
	}()
	wg.Add(1)
	//производим отправку пачкой метрики
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
		go worker(w, jobs, conf, lg)
	}
	wg.Wait()
}

func worker(id int, jobs <-chan models.Metrics, c config.ConfigAgent, log *zap.Logger) {
	for job := range jobs {
		str := fmt.Sprintf("рабочий %d, Start запущена задача %s", id, job)
		log.Info("Worker", zap.String("worker", str))
		metrics.SendRequestMetric(c, job, log)
	}
}
