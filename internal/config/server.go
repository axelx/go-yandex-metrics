package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type ConfigServer struct {
	FlagRunAddr         string
	FlagLogLevel        string
	FlagStoreInternal   int
	FlagFileStoragePath string
	FlagRestore         bool
}

func (c *ConfigServer) String() string {
	return fmt.Sprintf("FlagRunAddr: %s, FlagLogLevel: %s, FlagStoreInternal: %v, FlagFileStoragePath: %s, FlagRestore: %v",
		c.FlagRunAddr, c.FlagLogLevel, c.FlagStoreInternal, c.FlagFileStoragePath, c.FlagRestore)
}

func NewConfigServer() *ConfigServer {
	conf := ConfigServer{
		FlagRunAddr:         "",
		FlagLogLevel:        "",
		FlagStoreInternal:   0,
		FlagFileStoragePath: "",
		FlagRestore:         true,
	}
	parseFlagsServer(&conf)

	return &conf
}

func parseFlagsServer(c *ConfigServer) {
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
	if envStoreInternal := os.Getenv("STORE_INTERVAL"); envStoreInternal != "" {
		if v, err := strconv.Atoi(envStoreInternal); err == nil {
			c.FlagStoreInternal = v
		}
	}
	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		c.FlagFileStoragePath = envFileStoragePath
	}
	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		if v, err := strconv.ParseBool(envRestore); err == nil {
			c.FlagRestore = v
		}
	}
}
