package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"internal/handlers"
	"internal/storage"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
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
	m := storage.NewStorage()
	m.SetGauge("HeapAlloc", 5.5)
	m.SetCounter("PollCount", 5)

	ts := httptest.NewServer(handlers.Router(m))
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

	m := storage.NewStorage()
	m.SetGauge("HeapAlloc", 5.5)
	m.SetCounter("PollCount", 5)

	ts := httptest.NewServer(handlers.Router(m))
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
