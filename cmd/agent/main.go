// Agent для сбора и отправки метрик на сервер
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	logger.Log.Info("Running server", "config"+conf.String())

	metric := metrics.New(conf)

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				metric.Poll()
			}
		}
	}(ctx)

	//производим отправку пачкой метрики
	go func(ctx context.Context) {
		metric.ReportBatch(ctx)
	}(ctx)

	jobs := make(chan models.Metrics, 30)

	go func(ctx context.Context) {
		metric.Report(ctx, jobs)
	}(ctx)

	// создаем воркеры которые будут отправлять метрики из канала jobs
	for w := 1; w <= conf.RateLimit; w++ {
		go worker(w, jobs, conf)
	}

	<-sigint
	cancel()

	time.Sleep(time.Second * 2)

	fmt.Println("Agent Shutdown gracefully")
}

func worker(id int, jobs <-chan models.Metrics, c config.ConfigAgent) {
	for job := range jobs {
		str := fmt.Sprintf("рабочий %d, Start запущена задача", id)
		logger.Log.Info("Worker", "worker"+str)
		metrics.SendRequestMetric(c, job)
	}
}
