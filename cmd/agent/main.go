package main

import (
	"fmt"
	"internal/config"
	"internal/metrics"
)

func main() {
	conf := config.NewConfigAgent()

	fmt.Println("Running server on", conf.BaseURL, conf.ReportFrequency, conf.PollFrequency)

	metric := metrics.New()
	metric.Report(conf)
}
