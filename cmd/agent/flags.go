package main

import (
	"flag"
	"os"
	"strconv"
)

var flagServerAddr string
var flagReportFrequency int
var flagPollFrequency int

func parseFlags() {
	flag.StringVar(&flagServerAddr, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&flagReportFrequency, "r", 10, "report frequency to run server")
	flag.IntVar(&flagPollFrequency, "p", 2, "poll frequency")
	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		flagServerAddr = envRunAddr

	}
	if envReportFrequency := os.Getenv("REPORT_INTERVAL"); envReportFrequency != "" {
		if v, err := strconv.Atoi(envReportFrequency); err == nil {
			flagReportFrequency = v
		}
	}
	if envPollFrequency := os.Getenv("POLL_INTERVAL"); envPollFrequency != "" {
		if v, err := strconv.Atoi(envPollFrequency); err == nil {
			flagPollFrequency = v
		}
	}
}

//ADDRESS=localhost:8082 REPORT_INTERVAL=4  POLL_INTERVAL=3 go run . -a localhost:8081 -p=1 -r=2
