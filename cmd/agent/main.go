package main

import (
	"fmt"
	"github.com/axelx/go-yandex-metrics/internal/config"
	"github.com/axelx/go-yandex-metrics/internal/metrics"
)

func main() {
	conf := config.NewConfigAgent()

	fmt.Println("Running server on", conf.BaseURL, conf.ReportFrequency, conf.PollFrequency)

	metric := metrics.New()
	metric.Report(conf)
}
