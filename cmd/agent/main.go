// Agent для сбора и отправки метрик на сервер
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

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

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// conf — принимаем конфигурацию модуля.
	conf := config.NewConfigAgent()

	if err := logger.Initialize("info"); err != nil {
		fmt.Println(err)
	}
	logger.Info("Running server", "config"+conf.String())

	metric := metrics.New(conf)

	var wg sync.WaitGroup
	wg.Add(1)
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				wg.Done()
				return
			default:
				metric.Poll()
			}
		}
	}(ctx)

	wg.Add(1)
	//производим отправку метрик пачкой
	go func(ctx context.Context) {
		metric.ReportBatch(ctx)
		wg.Done()
	}(ctx)

	jobs := make(chan models.Metrics, 30)

	wg.Add(1)
	go func(ctx context.Context) {
		metric.Report(ctx, jobs)
		wg.Done()
	}(ctx)

	// создаем воркеры которые будут отправлять метрики из канала jobs
	for w := 1; w <= conf.RateLimit; w++ {
		go worker(w, jobs, conf)
	}

	<-sigint
	cancel()
	wg.Wait()

	fmt.Println("Agent Shutdown gracefully")
}

func worker(id int, jobs <-chan models.Metrics, c config.ConfigAgent) {
	for job := range jobs {
		str := fmt.Sprintf("рабочий %d, Start запущена задача", id)
		logger.Info("Worker", "worker"+str)
		metrics.SendRequestMetric(c, job)
	}
}
