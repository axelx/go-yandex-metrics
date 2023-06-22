package config

import (
	"flag"
	"os"
)

type ConfigServerFlag struct {
	FlagRunAddr string
}

func NewConfigServer() ConfigServerFlag {
	conf := ConfigServerFlag{
		FlagRunAddr: "",
	}
	parseFlagsServer(&conf)

	return conf
}

func parseFlagsServer(c *ConfigServerFlag) {
	flag.StringVar(&c.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		c.FlagRunAddr = envRunAddr
	}
}
