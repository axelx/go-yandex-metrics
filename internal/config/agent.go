package config

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

type ConfigAgentFlag struct {
	FlagServerAddr      string
	FlagReportFrequency int
	FlagPollFrequency   int
	FlagHashKey         string
	FlagRateLimit       int
}

type ConfigAgent struct {
	Client          *http.Client
	BaseURL         string
	ReportFrequency int
	PollFrequency   int
	RetryIntervals  []time.Duration
	FlagHashKey     string
	FlagRateLimit   int
}

func (c *ConfigAgent) String() string {
	return fmt.Sprintf("Client: , BaseURL: %s, ReportFrequency: %v, PollFrequency: %d, RetryIntervals: %v, FlagHashKey: %s, FlagRateLimit: %d",
		c.BaseURL, c.ReportFrequency, c.PollFrequency, c.RetryIntervals, c.FlagHashKey, c.FlagRateLimit)
}

// NewConfigAgent создаём конфигурацию агента для сбора и отправки метрик
func NewConfigAgent() ConfigAgent {

	cf := ConfigAgentFlag{
		FlagServerAddr:      "",
		FlagReportFrequency: 1,
		FlagPollFrequency:   1,
		FlagHashKey:         "",
		FlagRateLimit:       1,
	}
	parseFlagsAgent(&cf)

	var rt = http.DefaultTransport

	confDefault := ConfigAgent{
		Client: &http.Client{
			Transport: rt,
		},
		BaseURL:         "http://localhost:8080/",
		ReportFrequency: 10,
		PollFrequency:   2,
		RetryIntervals:  []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
		FlagHashKey:     "",
		FlagRateLimit:   1,
	}

	if cf.FlagServerAddr != "" {
		confDefault.BaseURL = "http://" + cf.FlagServerAddr + "/"
	}
	if cf.FlagReportFrequency != 0 {
		confDefault.ReportFrequency = cf.FlagReportFrequency
	}
	if cf.FlagPollFrequency != 0 {
		confDefault.PollFrequency = cf.FlagPollFrequency
	}
	if cf.FlagHashKey != "" {
		confDefault.FlagHashKey = cf.FlagHashKey
	}
	if cf.FlagRateLimit != 0 {
		confDefault.FlagRateLimit = cf.FlagRateLimit
	}
	return confDefault
}

func parseFlagsAgent(c *ConfigAgentFlag) {
	flag.StringVar(&c.FlagServerAddr, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&c.FlagReportFrequency, "r", 10, "report frequency to run server")
	flag.IntVar(&c.FlagPollFrequency, "p", 2, "poll frequency")
	flag.StringVar(&c.FlagHashKey, "k", "", "hash key")
	flag.IntVar(&c.FlagRateLimit, "l", 1, "simultaneous outgoing requests to the server")
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
	if envKey := os.Getenv("KEY"); envKey != "" {
		c.FlagHashKey = envKey
	}
	if envFlagRateLimit := os.Getenv("RATE_LIMIT"); envFlagRateLimit != "" {
		if v, err := strconv.Atoi(envFlagRateLimit); err == nil {
			c.FlagRateLimit = v
		}
	}
}
