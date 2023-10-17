package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type ConfigServer struct {
	RunAddr         string
	LogLevel        string
	StoreInternal   int
	FileStoragePath string
	Restore         bool
	DatabaseDSN     string
	HashKey         string
	CryptoKey       string
}

// String нужен для чтения всех параметров конфига
func (c *ConfigServer) String() string {
	return fmt.Sprintf("RunAddr: %s, LogLevel: %s, StoreInternal: %v, FileStoragePath: %s, Restore: %v, DatabaseDSN: %s, HashKey: %s, CryptoKey: %s",
		c.RunAddr, c.LogLevel, c.StoreInternal, c.FileStoragePath, c.Restore, c.DatabaseDSN, c.HashKey, c.CryptoKey)
}

// NewConfigServer создаём конфигурацию сервера для получения и сохранения метрик
func NewConfigServer() *ConfigServer {
	conf := ConfigServer{
		RunAddr:         "",
		LogLevel:        "",
		StoreInternal:   0,
		FileStoragePath: "",
		Restore:         true,
		DatabaseDSN:     "",
		HashKey:         "",
		CryptoKey:       "",
	}
	parseFlagsServer(&conf)

	return &conf
}

func parseFlagsServer(c *ConfigServer) {
	flag.StringVar(&c.RunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&c.LogLevel, "l", "info", "log level")
	flag.IntVar(&c.StoreInternal, "i", 300, "STORE_INTERVAL int")
	flag.StringVar(&c.FileStoragePath, "f", "/tmp/metrics-db.json", "FILE_STORAGE_PATH string")
	flag.BoolVar(&c.Restore, "r", true, "RESTORE true/false")
	flag.StringVar(&c.DatabaseDSN, "d", "", "DATABASE_DSN string")
	flag.StringVar(&c.HashKey, "k", "", "hash key")
	flag.StringVar(&c.CryptoKey, "crypto-key", "", "location privat key")

	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		c.RunAddr = envRunAddr
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		c.LogLevel = envLogLevel
	}
	if envStoreInternal := os.Getenv("STORE_INTERVAL"); envStoreInternal != "" {
		if v, err := strconv.Atoi(envStoreInternal); err == nil {
			c.StoreInternal = v
		}
	}
	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		c.FileStoragePath = envFileStoragePath
	}
	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		if v, err := strconv.ParseBool(envRestore); err == nil {
			c.Restore = v
		}
	}
	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		c.DatabaseDSN = envDatabaseDSN
	}
	if envKey := os.Getenv("KEY"); envKey != "" {
		c.HashKey = envKey
	}
	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		c.CryptoKey = envCryptoKey
	}
}
