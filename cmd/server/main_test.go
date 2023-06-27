package main

import (
	"bytes"
	"encoding/json"
	"github.com/axelx/go-yandex-metrics/internal/handlers"
	"github.com/axelx/go-yandex-metrics/internal/models"
	"github.com/axelx/go-yandex-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, bytes.NewReader([]byte(body)))
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestGetMetric(t *testing.T) {
	m := storage.New()
	g := 5.5
	c := int64(5)
	m.SetGauge("HeapAlloc", &g)
	m.SetCounter("PollCount", &c)

	hd := handlers.New(&m)
	ts := httptest.NewServer(hd.Router("info"))
	defer ts.Close()

	var testTable = []struct {
		url    string
		want   string
		status int
		body   string
	}{
		{
			url:    "/value/",
			want:   `{"id":"HeapAlloc","type":"gauge","value":5.5}`,
			status: http.StatusOK,
			body:   `{"id":"HeapAlloc",	"type":"gauge"}`,
		},
		{"/value/", `{"id":"PollCount","type":"counter","delta":5}`, http.StatusOK,
			`{"id":"PollCount",	"type":"counter"}`},
	}
	for _, v := range testTable {
		resp, data := testRequest(t, ts, "POST", v.url, v.body)

		var result models.Metrics
		if err := json.Unmarshal([]byte(data), &result); err != nil {
			panic(err)
		}

		defer resp.Body.Close()
		assert.Equal(t, v.status, resp.StatusCode)
		assert.Equal(t, v.want, data)
	}
}

func TestUpdatedMetric(t *testing.T) {
	m := storage.New()
	g := 5.5
	c := int64(5)
	m.SetGauge("HeapAlloc", &g)
	m.SetCounter("PollCount", &c)

	hd := handlers.New(&m)
	ts := httptest.NewServer(hd.Router("info"))
	defer ts.Close()

	type want struct {
		statusCode int
		data       string
	}
	tests := []struct {
		name string
		args storage.MemStorage
		want want
		url  string
		body string
	}{
		{
			name: "first",
			args: m,
			want: want{statusCode: 200, data: `{"id":"PollCount","type":"counter","delta":1}`},
			url:  "/update/",
			body: `{"id":"PollCount","type":"counter","delta":1}`,
		},
	}

	for _, v := range tests {
		resp, post := testRequest(t, ts, "POST", v.url, v.body)
		defer resp.Body.Close()
		assert.Equal(t, v.want.statusCode, resp.StatusCode)
		assert.Equal(t, v.want.data, post)
	}
}
