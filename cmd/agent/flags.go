package main

import (
	"flag"
)

// неэкспортированная переменная flagRunAddr содержит адрес и порт для запуска сервера
var flagRunAddr string
var flagReportFrequency int
var flagPollFrequency int

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func parseFlags() {
	// регистрируем переменную flagRunAddr
	// как аргумент -a со значением :8080 по умолчанию
	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.IntVar(&flagReportFrequency, "r", 10, "report frequency to run server")
	flag.IntVar(&flagPollFrequency, "p", 2, "poll frequency")
	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()
}
