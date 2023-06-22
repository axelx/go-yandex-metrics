package service

import (
	"errors"
	"strconv"
)

func PrepareFloat64Data(data string) (float64, error) {
	f, err := strconv.ParseFloat(data, 64)
	if err != nil {
		return 0, errors.New("ошибка обработки параметра float64 data ")
	}
	return f, nil
}

func PrepareInt64Data(data string) (int64, error) {
	i, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		return 0, errors.New("ошибка обработки параметра int64 data ")
	}
	return i, nil
}
