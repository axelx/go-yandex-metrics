package config

import (
	"flag"
	"os"
)

type ConfigServerFlag struct {
	FlagRunAddr  string
	FlagLogLevel string
}

func NewConfigServer() ConfigServerFlag {
	conf := ConfigServerFlag{
		FlagRunAddr:  "",
		FlagLogLevel: "",
	}
	parseFlagsServer(&conf)

	return conf
}

func parseFlagsServer(c *ConfigServerFlag) {
	flag.StringVar(&c.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&c.FlagLogLevel, "l", "info", "log level")

	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		c.FlagRunAddr = envRunAddr
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		c.FlagLogLevel = envLogLevel
	}
}
