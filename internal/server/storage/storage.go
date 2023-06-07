package storage

import (
	"errors"
	"fmt"
	"strconv"
)

type MemStorage struct {
	Gauge   map[string]float64 //новое значение должно замещать предыдущее.
	Counter map[string]int64   //новое значение должно добавляться к предыдущему, если какое-то значение уже было известно серверу.
}

func (m *MemStorage) SetGauge(nameMetric, data string) error {
	if f, err := strconv.ParseFloat(data, 64); err == nil {
		m.Gauge[nameMetric] = f
		return nil
	}
	return errors.New("ошибка обработки параметра gauge")
}

func (m *MemStorage) SetCounter(nameMetric, data string) error {
	if i, err := strconv.ParseInt(data, 10, 64); err == nil {
		m.Counter[nameMetric] += i
		return nil
	}
	return errors.New("ошибка обработки параметра counter " + nameMetric)
}

func StorageTest() {
	fmt.Println("Storage test")
}
