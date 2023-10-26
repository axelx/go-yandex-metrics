package storage

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/axelx/go-yandex-metrics/internal/logger"
	"github.com/axelx/go-yandex-metrics/internal/models"
	"github.com/axelx/go-yandex-metrics/internal/mos"
)

var (
	ErrNotFoundMetric = errors.New("не найдена метрика")
	ErrMetricIsNil    = errors.New("нулевая метрика или пустое название метрики")
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

func (m *MemStorage) GetMetric(typeMetric models.MetricType, nameMetric string) (string, error) {
	switch typeMetric {
	case models.MetricGauge:
		v, t := m.gauge[nameMetric]
		if !t {
			return "", ErrNotFoundMetric
		}
		return fmt.Sprint(v), nil
	case models.MetricCounter:
		v, t := m.counter[nameMetric]
		if !t {
			return "", ErrNotFoundMetric
		}
		return strconv.FormatInt(v, 10), nil
	default:
		return "метрика не найдена", ErrNotFoundMetric
	}
}

func (m *MemStorage) SetJSONGauge(nameMetric string, data *float64) error {
	if nameMetric == "" || data == nil {
		return ErrMetricIsNil
	}
	m.gauge[nameMetric] = *data
	if m.UpdateInterval == 0 {
		m.FileUpdate(models.Metrics{ID: nameMetric, MType: "gauge", Value: data})
	}
	return nil
}

func (m *MemStorage) SetJSONCounter(nameMetric string, data *int64) error {
	if nameMetric == "" || data == nil {
		return ErrMetricIsNil
	}
	m.counter[nameMetric] += *data
	t := m.counter[nameMetric]
	if m.UpdateInterval == 0 {
		m.FileUpdate(models.Metrics{ID: nameMetric, MType: "counter", Delta: &t})
	}
	return nil
}

func (m *MemStorage) GetJSONMetric(typeMetric models.MetricType, nameMetric string) (models.Metrics, error) {
	mt := models.Metrics{}
	switch typeMetric {
	case "gauge":
		v, t := m.gauge[nameMetric]
		if !t {
			return mt, ErrNotFoundMetric
		}
		mt = models.Metrics{MType: typeMetric, ID: nameMetric, Value: &v}
		return mt, nil
	case "counter":
		v, t := m.counter[nameMetric]
		if !t {
			return mt, ErrNotFoundMetric
		}
		mt = models.Metrics{MType: typeMetric, ID: nameMetric, Delta: &v}
		return mt, nil
	default:
		return mt, ErrNotFoundMetric
	}
}

func (m *MemStorage) GetTypeMetric(t models.MetricType) interface{} {
	switch t {
	case models.MetricGauge:
		return reflect.ValueOf(m).Interface().(*MemStorage).gauge
	case models.MetricCounter:
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
	mos.SaveMetricsToFile(m.fileName, sm)
	return nil
}

func (m *MemStorage) toModelMetric() []models.Metrics {
	metrics := []models.Metrics{}
	for n, v := range m.gauge {
		vt := v
		metrics = append(metrics, models.Metrics{ID: n, MType: "gauge", Value: &vt})
	}
	for n, v := range m.counter {
		vt := v
		metrics = append(metrics, models.Metrics{ID: n, MType: "counter", Delta: &vt})
	}
	return metrics
}

// RestoreFromFile восстановление метрик из файла
func (m *MemStorage) RestoreFromFile() {
	if !m.restore {
		return
	}
	sv := m.ReadFile()
	for _, metric := range sv {
		if metric.MType == models.MetricGauge {
			m.gauge[metric.ID] = *metric.Value
		} else if metric.MType == models.MetricCounter {
			m.counter[metric.ID] = *metric.Delta
		}

		s := fmt.Sprintf("load metrics: %s, %s, %f, %d", metric.MType, metric.ID, *metric.Value, *metric.Delta)
		logger.Info("load metrics ", "info"+s)
	}
}

// UpdateFile обновляем метрики в файле
func (m *MemStorage) UpdateFile(ctx context.Context) {
	if m.UpdateInterval == 0 {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		default:
			m.SaveMetricToFile()
			logger.Info("updateMemstorage from file ", "")
			time.Sleep(time.Duration(m.UpdateInterval) * time.Second)
		}
	}
}
