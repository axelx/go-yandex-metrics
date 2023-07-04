package storage

import (
	"errors"
	"fmt"
	"github.com/axelx/go-yandex-metrics/internal/models"
	"github.com/axelx/go-yandex-metrics/internal/mos"
	"reflect"
	"strconv"
)

type MemStorage struct {
	gauge          map[string]float64 //новое значение должно замещать предыдущее.
	counter        map[string]int64   //новое значение должно добавляться к предыдущему, если какое-то значение уже было известно серверу.
	fileName       string
	UpdateInterval int
	restore        bool
}

func New(filename string, updateInterval int, restoreFromFile bool) MemStorage {
	return MemStorage{
		gauge:          map[string]float64{},
		counter:        map[string]int64{},
		fileName:       filename,
		UpdateInterval: updateInterval,
		restore:        restoreFromFile,
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

func (m *MemStorage) SetJSONGauge(nameMetric string, data *float64) error {
	m.gauge[nameMetric] = *data
	if m.UpdateInterval == 0 {
		m.FileUpdate(models.Metrics{ID: nameMetric, MType: "gauge", Value: data})
	}
	return nil
}

func (m *MemStorage) SetJSONCounter(nameMetric string, data *int64) error {
	m.counter[nameMetric] += *data
	t := m.counter[nameMetric]
	if m.UpdateInterval == 0 {
		m.FileUpdate(models.Metrics{ID: nameMetric, MType: "counter", Delta: &t})
	}
	return nil
}

func (m *MemStorage) GetJSONMetric(typeMetric, nameMetric string) (models.Metrics, error) {
	err := errors.New("не найдена метрика")
	mt := models.Metrics{}
	switch typeMetric {
	case "gauge":
		v, t := m.gauge[nameMetric]
		if !t {
			return mt, err
		}
		mt = models.Metrics{MType: typeMetric, ID: nameMetric, Value: &v}
		return mt, nil
	case "counter":
		v, t := m.counter[nameMetric]
		if !t {
			return mt, err
		}
		mt = models.Metrics{MType: typeMetric, ID: nameMetric, Delta: &v}
		return mt, nil
	default:
		return mt, err
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

func (m *MemStorage) FileUpdate(metric models.Metrics) {
	sm := m.ReadFile()
	sm = dataUpdateOrAdd(sm, metric)
	mos.SaveMetricsToFile(m.fileName, sm)
}
func dataUpdateOrAdd(sm []models.Metrics, metric models.Metrics) []models.Metrics {
	addF := true
	for _, m := range sm {
		if m.MType == "gauge" {
			if m.ID == metric.ID {
				*m.Value = *metric.Value
				addF = false
			}
		} else if m.MType == "counter" {
			if m.ID == metric.ID {
				*m.Delta = *metric.Delta
				addF = false
			}
		}
	}
	if addF {
		sm = append(sm, metric)
	}

	return sm
}

func (m *MemStorage) ReadFile() []models.Metrics {
	return mos.ReadAllFile(m.fileName)
}
func (m *MemStorage) SaveMetricToFile() error {
	if m.fileName == "" {
		return nil
	}
	sm := m.toModelMetric()
	mos.SaveMetricsToFile("/tmp/metrics-db.json", sm)
	return nil
}

func (m *MemStorage) toModelMetric() []models.Metrics {
	metrics := []models.Metrics{}
	for n, v := range m.gauge {
		metrics = append(metrics, models.Metrics{ID: n, MType: "gauge", Value: &v})
	}
	for n, v := range m.counter {
		metrics = append(metrics, models.Metrics{ID: n, MType: "counter", Delta: &v})
	}
	return metrics
}

func (m *MemStorage) RestoreFromFile() {
	if !m.restore {
		return
	}
	sv := m.ReadFile()
	fmt.Println("sv", m.restore)
	for _, metric := range sv {
		fmt.Println(metric)
		if metric.MType == "gauge" {
			m.gauge[metric.ID] = *metric.Value
		} else if metric.MType == "counter" {
			m.counter[metric.ID] = *metric.Delta
		}
	}
}
