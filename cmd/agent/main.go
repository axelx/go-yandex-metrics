package main

import (
	"fmt"
	"github.com/axelx/go-yandex-metrics/internal/server/config"
	"github.com/axelx/go-yandex-metrics/internal/server/metrics"
)

func main() {
	conf := config.NewConfigAgent()

	fmt.Println("Running server on", conf.BaseURL, conf.ReportFrequency, conf.PollFrequency)

	metric := metrics.New()
	metric.Report(conf)
}
