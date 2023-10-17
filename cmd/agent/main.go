// Agent для сбора и отправки метрик на сервер
package main

import (
	"fmt"
	"sync"

	"github.com/axelx/go-yandex-metrics/internal/config"
	"github.com/axelx/go-yandex-metrics/internal/logger"
	"github.com/axelx/go-yandex-metrics/internal/metrics"
	"github.com/axelx/go-yandex-metrics/internal/models"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version:= %s\n", buildVersion)
	fmt.Printf("Build date:= %s\n", buildDate)
	fmt.Printf("Build commit:= %s\n", buildCommit)

	// conf — принимаем конфигурацию модуля.
	conf := config.NewConfigAgent()

	if err := logger.Initialize("info"); err != nil {
		fmt.Println(err)
	}
	logger.Log.Info("Running server", "config"+conf.String())

	metric := metrics.New(conf)

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

	go func() {
		metric.Report(jobs)
	}()

	// создаем воркеры которые будут отправлять метрики из канала jobs
	for w := 1; w <= conf.RateLimit; w++ {
		go worker(w, jobs, conf)
	}
	wg.Wait()
}

func worker(id int, jobs <-chan models.Metrics, c config.ConfigAgent) {
	for job := range jobs {
		str := fmt.Sprintf("рабочий %d, Start запущена задача", id)
		logger.Log.Info("Worker", "worker"+str)
		metrics.SendRequestMetric(c, job)
	}
}
