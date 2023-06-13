package main

import (
	"flag"
	"internal/config"
	"os"
	"strconv"
)

func parseFlags(c *config.ConfigAgentFlag) {

	flag.StringVar(&c.FlagServerAddr, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&c.FlagReportFrequency, "r", 10, "report frequency to run server")
	flag.IntVar(&c.FlagPollFrequency, "p", 2, "poll frequency")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		c.FlagServerAddr = envRunAddr

	}
	if envReportFrequency := os.Getenv("REPORT_INTERVAL"); envReportFrequency != "" {
		if v, err := strconv.Atoi(envReportFrequency); err == nil {
			c.FlagReportFrequency = v
		}
	}
	if envPollFrequency := os.Getenv("POLL_INTERVAL"); envPollFrequency != "" {
		if v, err := strconv.Atoi(envPollFrequency); err == nil {
			c.FlagPollFrequency = v
		}
	}
}

//ADDRESS=localhost:8082 REPORT_INTERVAL=4  POLL_INTERVAL=3 go run . -a localhost:8081 -p=1 -r=2
