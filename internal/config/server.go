package config

import (
	"flag"
	"os"
	"strconv"
)

type ConfigServerFlag struct {
	FlagRunAddr         string
	FlagLogLevel        string
	FlagStoreInternal   int
	FlagFileStoragePath string
	FlagRestore         bool
}

func NewConfigServer() ConfigServerFlag {
	conf := ConfigServerFlag{
		FlagRunAddr:         "",
		FlagLogLevel:        "",
		FlagStoreInternal:   0,
		FlagFileStoragePath: "",
		FlagRestore:         false,
	}
	parseFlagsServer(&conf)

	return conf
}

func parseFlagsServer(c *ConfigServerFlag) {
	flag.StringVar(&c.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&c.FlagLogLevel, "l", "info", "log level")
	flag.IntVar(&c.FlagStoreInternal, "i", 300, "STORE_INTERVAL int")
	flag.StringVar(&c.FlagFileStoragePath, "f", "/tmp/metrics-db.json", "FILE_STORAGE_PATH string")
	flag.BoolVar(&c.FlagRestore, "r", true, "RESTORE true/false")

	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		c.FlagRunAddr = envRunAddr
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		c.FlagLogLevel = envLogLevel
	}
	if envStoreInternal := os.Getenv("LOG_LEVEL"); envStoreInternal != "" {
		if v, err := strconv.Atoi(envStoreInternal); err == nil {
			c.FlagStoreInternal = v
		}
	}
	if envFileStoragePath := os.Getenv("LOG_LEVEL"); envFileStoragePath != "" {
		c.FlagFileStoragePath = envFileStoragePath
	}
	if envRestore := os.Getenv("LOG_LEVEL"); envRestore != "" {
		if v, err := strconv.ParseBool(envRestore); err == nil {
			c.FlagRestore = v
		}
	}
}
