package storage

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

type MemStorage struct {
	gauge   map[string]float64 //новое значение должно замещать предыдущее.
	counter map[string]int64   //новое значение должно добавляться к предыдущему, если какое-то значение уже было известно серверу.
}

func New() MemStorage {
	return MemStorage{
		gauge:   map[string]float64{},
		counter: map[string]int64{},
	}
}

func (m *MemStorage) SetGauge(nameMetric string, data float64) error {
	m.gauge[nameMetric] = data
	return nil
}

func (m *MemStorage) SetCounter(nameMetric string, data int64) error {
	m.counter[nameMetric] += data
	return nil
}

func (m *MemStorage) GetMetric(typeMetric, nameMetric string) (string, error) {
	err := errors.New("не найдена метрика")
	switch typeMetric {
	case "gauge":
		v, t := m.gauge[nameMetric]
		if !t {
			return "", err
		}
		return fmt.Sprint(v), nil
	case "counter":
		v, t := m.counter[nameMetric]
		if !t {
			return "", err
		}
		return strconv.FormatInt(v, 10), nil
	default:
		return "метрика не найдена", err
	}
}

func (m *MemStorage) GetTypeMetric(t string) interface{} {
	switch t {
	case "gauge":
		return reflect.ValueOf(m).Interface().(*MemStorage).gauge
	case "counter":
		return reflect.ValueOf(m).Interface().(*MemStorage).counter
	}

	return nil
}
