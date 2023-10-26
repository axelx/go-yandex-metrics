package mos

import (
	"bufio"
	"encoding/json"
	"log"
	"os"

	"github.com/axelx/go-yandex-metrics/internal/logger"
	"github.com/axelx/go-yandex-metrics/internal/models"
)

type DataEncode struct {
	file    *os.File
	encoder *json.Encoder
	decoder *json.Decoder
}

func NewProducer(filename string) (*DataEncode, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &DataEncode{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *DataEncode) WriteMetric(metric models.Metrics) error {
	return p.encoder.Encode(&metric)
}

func (p *DataEncode) Close() error {
	return p.file.Close()
}

func NewConsumer(filename string) (*DataEncode, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return &DataEncode{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (p *DataEncode) ReadMetric() (*models.Metrics, error) {
	metric := &models.Metrics{}
	if err := p.decoder.Decode(&metric); err != nil {
		return nil, err
	}
	return metric, nil
}

func SaveMetricsToFile(fileName string, metrics []models.Metrics) error {
	if fileName == "" {
		return nil
	}
	os.Remove(fileName)
	for _, m := range metrics {
		Producer, err := NewProducer(fileName)
		if err != nil {
			log.Fatal(err)
		}
		defer Producer.Close()
		if err := Producer.WriteMetric(m); err != nil {
			log.Fatal(err)
		}
	}
	return nil
}
func ReadAllFile(fileName string) []models.Metrics {
	if fileName == "" {
		return nil
	}
	file, err := os.Open(fileName)
	if err != nil {
		logger.Error("Error mos.ReadAllFile", "about ERR"+err.Error())
		return nil
	}
	defer file.Close()

	sm := []models.Metrics{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ev := models.Metrics{}
		json.Unmarshal([]byte(scanner.Text()), &ev)
		sm = append(sm, ev)
	}

	if err := scanner.Err(); err != nil {
		logger.Error("Error mos.ReadAllFile  ", "about ERR: scanner.Err; "+err.Error())
	}
	return sm
}
