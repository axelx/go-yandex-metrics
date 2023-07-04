package m_os

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/axelx/go-yandex-metrics/internal/models"
	"log"
	"os"
)

type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

func NewProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *Producer) WriteMetric(metric models.Metrics) error {

	err := p.encoder.Encode(metric)
	return err

	return p.encoder.Encode(&metric)
}

func (p *Producer) Close() error {
	return p.file.Close()
}

type Consumer struct {
	file    *os.File
	decoder *json.Decoder
}

func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return &Consumer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *Consumer) ReadMetric() (*models.Metrics, error) {
	metric := &models.Metrics{}
	if err := c.decoder.Decode(&metric); err != nil {
		return nil, err
	}
	return metric, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
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

	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("-Open file err", err)
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
		log.Fatal(err)
	}
	return sm
}
