package main

import (
	"fmt"
	"internal/config"
	"internal/metrics"
)

func main() {

	confF := config.NewConfigAgentFlag()
	parseFlags(&confF)
	conf := config.NewConfigAgent(confF)

	fmt.Println("Running server on", conf.BaseURL, conf.ReportFrequency, conf.PollFrequency)

	metric := metrics.New()
	metric.Report(conf)
}
