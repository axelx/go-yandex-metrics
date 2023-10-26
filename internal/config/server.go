package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/axelx/go-yandex-metrics/internal/logger"
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
	serverConfigDefaultValue(&conf)

	return &conf
}

var configServerJSON string

func parseFlagsServer(c *ConfigServer) {
	flag.StringVar(&c.RunAddr, "a", "", "address and port to run server")
	flag.StringVar(&c.LogLevel, "l", "", "log level")
	flag.IntVar(&c.StoreInternal, "i", 0, "STORE_INTERVAL int")
	flag.StringVar(&c.FileStoragePath, "f", "", "FILE_STORAGE_PATH string")
	flag.BoolVar(&c.Restore, "r", false, "RESTORE true/false")
	flag.StringVar(&c.DatabaseDSN, "d", "", "DATABASE_DSN string")
	flag.StringVar(&c.HashKey, "k", "", "hash key")
	flag.StringVar(&c.CryptoKey, "crypto-key", "", "location privat key")
	flag.StringVar(&configServerJSON, "c", "", "config json")
	flag.StringVar(&configServerJSON, "config", "", "config json")

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
	if envConfigJSON := os.Getenv("CONFIG"); envConfigJSON != "" {
		configServerJSON = envConfigJSON
	}

	parseConfigJSON(configServerJSON, c)

}

func parseConfigJSON(configServerJSON string, c *ConfigServer) *ConfigServer {
	if configServerJSON == "" {
		return c
	}

	type ConfigServerJSON struct {
		StoreInternalJSON string `json:"store_interval"`
		RunAddr           string `json:"address"`
		FileStoragePath   string `json:"store_file"`
		Restore           bool   `json:"restore"`
		DatabaseDSN       string `json:"database_dsn"`
		CryptoKey         string `json:"crypto_key"`
	}
	f, err := os.Open(configServerJSON)
	if err != nil {
		return c
	}
	defer f.Close()

	bs, err := io.ReadAll(f)
	if err != nil {
		logger.Error("config server", "io.ReadAll: "+err.Error())
		return c
	}
	cs := ConfigServerJSON{}
	err = json.Unmarshal(bs, &cs)
	if err != nil {
		logger.Error("config server", "Ошибка Unmarshal"+err.Error())
		return c
	}

	ss := strings.Split(cs.StoreInternalJSON, "s")

	si, err := strconv.Atoi(ss[0])
	if err != nil {
		logger.Error("config server", "Ошибка конвертации  Atoi"+err.Error())
		return c
	}

	if c.StoreInternal == 0 {
		c.StoreInternal = si
	}
	if c.RunAddr == "" {
		c.RunAddr = cs.RunAddr
	}
	if c.FileStoragePath == "" {
		c.FileStoragePath = cs.FileStoragePath
	}
	if !c.Restore {
		c.Restore = cs.Restore
	}
	if c.DatabaseDSN == "" {
		c.DatabaseDSN = cs.DatabaseDSN
	}
	if c.CryptoKey == "" {
		c.CryptoKey = cs.CryptoKey
	}

	return c
}

func serverConfigDefaultValue(c *ConfigServer) {
	if c.StoreInternal == 0 {
		c.StoreInternal = 300
	}
	if c.RunAddr == "" {
		c.RunAddr = "localhost:8080"
	}
	if c.FileStoragePath == "" {
		c.FileStoragePath = "/tmp/metrics-db.json"
	}
	if c.LogLevel == "" {
		c.LogLevel = "info"
	}
	if !c.Restore {
		c.Restore = true
	}
}
