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

func (m *MemStorage) GetMetric(typeMetric, nameMetric string) (string, error) {
	err := errors.New("не найдена метрика")
	switch typeMetric {
	case "gauge":
		v, t := m.Gauge[nameMetric]
		if !t {
			return "", err
		}
		return fmt.Sprintf("%.3f", v), nil
	case "counter":
		v, t := m.Counter[nameMetric]
		if !t {
			return "", err
		}
		return strconv.FormatInt(v, 10), nil
	default:
		return "метрика не найдена", err
	}
}

func StorageTest() {
	fmt.Println("Storage test")
}
