// Модуль metrics собирает метрики системы в рантайме и отправляет их по установленному урлу
package metrics

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"

	"github.com/axelx/go-yandex-metrics/internal/config"
	"github.com/axelx/go-yandex-metrics/internal/crypto"
	"github.com/axelx/go-yandex-metrics/internal/hash"
	"github.com/axelx/go-yandex-metrics/internal/logger"
	"github.com/axelx/go-yandex-metrics/internal/models"
	"github.com/axelx/go-yandex-metrics/internal/service"
)

func sendRequestSliceMetrics(c config.ConfigAgent, metrics []models.Metrics) error {
	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	if c.CryptoKey != "" {
		metricsJSON, err = crypto.EncodeRSAAES(metricsJSON, c.CryptoKey)
		if err != nil {
			return err
		}
	}
	err = sendRequest("updates/", c, metricsJSON)
	if err != nil {
		logger.Error("Error sendRequest", "about err: "+err.Error())
		return err
	}
	return nil
}

// SendRequestMetric отправка метрик по указанному в кофиге урлу
func SendRequestMetric(c config.ConfigAgent, metric models.Metrics) error {
	metricJSON, err := json.Marshal(metric)
	if err != nil {
		logger.Error("Error SendRequestMetric", "about err: "+err.Error())
		return err
	}
	if c.CryptoKey != "" {
		metricJSON, err = crypto.EncodeRSAAES(metricJSON, c.CryptoKey)
		if err != nil {
			return err
		}
	}

	err = sendRequest("update/", c, metricJSON)
	if err != nil {
		return err
	}
	return nil
}

func sendRequest(url string, c config.ConfigAgent, metricsJSON []byte) error {
	buf := bytes.NewBuffer(nil)
	zb := gzip.NewWriter(buf)
	_, err := zb.Write([]byte(metricsJSON))
	if err != nil {
		logger.Error("Error zb.Write([]byte(metricsJSON)", "sendRequest; about err: "+err.Error())
		return err
	}
	zb.Close()

	req, err := http.NewRequest("POST", c.BaseURL+url, buf)
	if err != nil {
		logger.Error("Error create request", "sendRequest; about err: "+err.Error())
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("X-Real-IP", service.GetIP())
	if c.HashKey != "" {
		req.Header.Set("HashSHA256", hash.GetHashSHA256Base64(metricsJSON, c.HashKey))
	}
	resp, err := c.Client.Do(req)
	resp.Body.Close()

	if err != nil {
		logger.Error("Error reporting metrics:", "metricJSON; about err: "+err.Error()+
			"about metricJSON: "+string(metricsJSON))
		return ErrDialUp
	}
	return nil
}
