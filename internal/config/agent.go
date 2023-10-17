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
	ServerAddr      string
	ReportFrequency int
	PollFrequency   int
	HashKey         string
	RateLimit       int
	CryptoKey       string
}

type ConfigAgent struct {
	Client          *http.Client
	BaseURL         string
	ReportFrequency int
	PollFrequency   int
	RetryIntervals  []time.Duration
	HashKey         string
	RateLimit       int
	CryptoKey       string
}

func (c *ConfigAgent) String() string {
	return fmt.Sprintf("Client: , BaseURL: %s, ReportFrequency: %v, PollFrequency: %d, RetryIntervals: %v, RateLimit: %d, HashKey: %s, CryptoKey: %s",
		c.BaseURL, c.ReportFrequency, c.PollFrequency, c.RetryIntervals, c.RateLimit, c.HashKey, c.CryptoKey)
}

// NewConfigAgent создаём конфигурацию агента для сбора и отправки метрик
func NewConfigAgent() ConfigAgent {

	cf := ConfigAgentFlag{
		ServerAddr:      "",
		ReportFrequency: 1,
		PollFrequency:   1,
		HashKey:         "",
		RateLimit:       1,
		CryptoKey:       "",
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
		HashKey:         "",
		RateLimit:       1,
		CryptoKey:       "",
	}

	if cf.ServerAddr != "" {
		confDefault.BaseURL = "http://" + cf.ServerAddr + "/"
	}
	if cf.ReportFrequency != 0 {
		confDefault.ReportFrequency = cf.ReportFrequency
	}
	if cf.PollFrequency != 0 {
		confDefault.PollFrequency = cf.PollFrequency
	}
	if cf.HashKey != "" {
		confDefault.HashKey = cf.HashKey
	}
	if cf.RateLimit != 0 {
		confDefault.RateLimit = cf.RateLimit
	}
	if cf.CryptoKey != "" {
		confDefault.CryptoKey = cf.CryptoKey
	}
	return confDefault
}

func parseFlagsAgent(c *ConfigAgentFlag) {
	flag.StringVar(&c.ServerAddr, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&c.ReportFrequency, "r", 10, "report frequency to run server")
	flag.IntVar(&c.PollFrequency, "p", 2, "poll frequency")
	flag.StringVar(&c.HashKey, "k", "", "hash key")
	flag.IntVar(&c.RateLimit, "l", 1, "simultaneous outgoing requests to the server")
	flag.StringVar(&c.CryptoKey, "crypto-key", "", "location public key")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		c.ServerAddr = envRunAddr
	}
	if envReportFrequency := os.Getenv("REPORT_INTERVAL"); envReportFrequency != "" {
		if v, err := strconv.Atoi(envReportFrequency); err == nil {
			c.ReportFrequency = v
		}
	}
	if envPollFrequency := os.Getenv("POLL_INTERVAL"); envPollFrequency != "" {
		if v, err := strconv.Atoi(envPollFrequency); err == nil {
			c.PollFrequency = v
		}
	}
	if envKey := os.Getenv("KEY"); envKey != "" {
		c.HashKey = envKey
	}
	if envRateLimit := os.Getenv("RATE_LIMIT"); envRateLimit != "" {
		if v, err := strconv.Atoi(envRateLimit); err == nil {
			c.RateLimit = v
		}
	}
	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		c.CryptoKey = envCryptoKey
	}
}
