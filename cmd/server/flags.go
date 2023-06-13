package main

import (
	"flag"
	"internal/config"
	"os"
)

func parseFlags(c *config.ConfigServerFlag) {
	flag.StringVar(&c.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		c.FlagRunAddr = envRunAddr
	}
}

//ADDRESS=localhost:8082 go run . -a localhost:8081
