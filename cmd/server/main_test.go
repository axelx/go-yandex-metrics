package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axelx/go-yandex-metrics/internal/handlers"
	"github.com/axelx/go-yandex-metrics/internal/models"
	"github.com/axelx/go-yandex-metrics/internal/pg"
	"github.com/axelx/go-yandex-metrics/internal/storage"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestGetMetric(t *testing.T) {
	m := storage.New("", 300, false)
	m.SetGauge("HeapAlloc", 5.5)
	m.SetCounter("PollCount", 5)

	NewDBStorage := pg.NewDBStorage()
	hd := handlers.New(&m, NewDBStorage.DB, NewDBStorage, "")
	ts := httptest.NewServer(hd.Router())
	defer ts.Close()

	var testTable = []struct {
		url    string
		want   string
		status int
	}{
		{"/value/counter/PollCount", "5", http.StatusOK},
		{"/value/gauge/HeapAlloc", "5.5", http.StatusOK},
	}
	for _, v := range testTable {
		resp, get := testRequest(t, ts, "GET", v.url)
		defer resp.Body.Close()
		assert.Equal(t, v.status, resp.StatusCode)
		assert.Equal(t, v.want, get)
	}
}

func TestUpdatedMetric(t *testing.T) {
	m := storage.New("", 300, false)
	m.SetGauge("HeapAlloc", 5.5)
	m.SetCounter("PollCount", 5)

	NewDBStorage := pg.NewDBStorage()
	hd := handlers.New(&m, NewDBStorage.DB, NewDBStorage, "")
	ts := httptest.NewServer(hd.Router())
	defer ts.Close()

	type want struct {
		statusCode int
		data       string
	}
	tests := []struct {
		name    string
		args    storage.MemStorage
		want    want
		request string
	}{
		{
			name:    "first",
			args:    m,
			want:    want{statusCode: 200, data: "5"},
			request: "/update/counter/PollCount/5",
		},
	}

	for _, v := range tests {
		resp, post := testRequest(t, ts, "POST", v.request)
		defer resp.Body.Close()
		assert.Equal(t, v.want.statusCode, resp.StatusCode)
		assert.Equal(t, v.want.data, post)
	}
}

func testJSONRequest(t *testing.T, ts *httptest.Server, method, path string, body string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, bytes.NewReader([]byte(body)))
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestGetJsonMetric(t *testing.T) {
	m := storage.New("", 300, false)
	g := 5.5
	c := int64(5)
	m.SetJSONGauge("HeapAlloc", &g)
	m.SetJSONCounter("PollCount", &c)

	NewDBStorage := pg.NewDBStorage()
	hd := handlers.New(&m, NewDBStorage.DB, NewDBStorage, "")
	ts := httptest.NewServer(hd.Router())
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
		resp, data := testJSONRequest(t, ts, "POST", v.url, v.body)

		var result models.Metrics
		if err := json.Unmarshal([]byte(data), &result); err != nil {
			panic(err)
		}

		defer resp.Body.Close()
		assert.Equal(t, v.status, resp.StatusCode)
		assert.Equal(t, v.want, data)
	}
}

func TestJsonUpdatedMetric(t *testing.T) {
	m := storage.New("", 300, false)
	g := 5.5
	c := int64(5)
	m.SetJSONGauge("HeapAlloc", &g)
	m.SetJSONCounter("PollCount", &c)

	NewDBStorage := pg.NewDBStorage()
	hd := handlers.New(&m, NewDBStorage.DB, NewDBStorage, "")
	ts := httptest.NewServer(hd.Router())
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
			want: want{statusCode: 200, data: `{"id":"PollCount","type":"counter","delta":6}`},
			url:  "/update/",
			body: `{"id":"PollCount","type":"counter","delta":1}`,
		},
	}

	for _, v := range tests {
		resp, post := testJSONRequest(t, ts, "POST", v.url, v.body)
		defer resp.Body.Close()
		assert.Equal(t, v.want.statusCode, resp.StatusCode)
		assert.Equal(t, v.want.data, post)
	}
}
