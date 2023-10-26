package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/axelx/go-yandex-metrics/internal/logger"
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

	agentConfigDefaultValue(&cf)

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

var configAgentJSON string

func parseFlagsAgent(c *ConfigAgentFlag) {
	flag.StringVar(&c.ServerAddr, "a", "", "address and port to run server")
	flag.IntVar(&c.ReportFrequency, "r", 0, "report frequency to run server")
	flag.IntVar(&c.PollFrequency, "p", 0, "poll frequency")
	flag.StringVar(&c.HashKey, "k", "", "hash key")
	flag.IntVar(&c.RateLimit, "l", 0, "simultaneous outgoing requests to the server")
	flag.StringVar(&c.CryptoKey, "crypto-key", "", "location public key")
	flag.StringVar(&configAgentJSON, "c", "", "config json")
	flag.StringVar(&configAgentJSON, "config", "", "config json")
	flag.Parse()

	fmt.Println("configAgentJSON", configAgentJSON)

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
	if envConfigJSON := os.Getenv("CONFIG"); envConfigJSON != "" {
		configAgentJSON = envConfigJSON
	}

	parseConfigAgentJSON(configAgentJSON, c)
}

func parseConfigAgentJSON(configJSON string, c *ConfigAgentFlag) *ConfigAgentFlag {
	if configJSON == "" {
		return c
	}

	type ConfigAgentJSON struct {
		RunAddr         string `json:"address"`
		ReportFrequency string `json:"report_interval"`
		PollFrequency   string `json:"poll_interval"`
		CryptoKey       string `json:"crypto_key"`
	}
	f, err := os.Open(configJSON)
	if err != nil {
		return c
	}
	defer f.Close()

	bs, err := io.ReadAll(f)
	if err != nil {
		logger.Error("config agent", "io.ReadAll: "+err.Error())
		return c
	}
	fmt.Println("io.ReadAll(f)", string(bs))

	cs := ConfigAgentJSON{}
	err = json.Unmarshal(bs, &cs)
	if err != nil {
		logger.Error("config agent", "Ошибка UnmarshalЖ"+err.Error())
		return c
	}

	rf := strings.Split(cs.ReportFrequency, "s")
	rfi, err := strconv.Atoi(rf[0])
	if err != nil {
		logger.Error("config agent", "Ошибка конвертации  Atoi rf"+err.Error())
		return c
	}
	pf := strings.Split(cs.PollFrequency, "s")
	pfi, err := strconv.Atoi(pf[0])
	if err != nil {
		logger.Error("config agent", "Ошибка конвертации  Atoi pf"+err.Error())
		return c
	}

	if c.ServerAddr == "" {
		c.ServerAddr = cs.RunAddr
	}
	if c.ReportFrequency == 0 {
		c.ReportFrequency = rfi
	}
	if c.PollFrequency == 0 {
		c.PollFrequency = pfi
	}
	if c.CryptoKey == "" {
		c.CryptoKey = cs.CryptoKey
	}

	return c
}

func agentConfigDefaultValue(c *ConfigAgentFlag) {
	if c.ServerAddr == "" {
		c.ServerAddr = "localhost:8080"
	}
	if c.ReportFrequency == 0 {
		c.ReportFrequency = 10
	}
	if c.PollFrequency == 0 {
		c.PollFrequency = 5
	}
	if c.RateLimit == 0 {
		c.RateLimit = 1
	}
}
