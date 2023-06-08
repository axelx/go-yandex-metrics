package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"internal/handlers"
	"internal/storage"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdMem(t *testing.T) {
	fmt.Println("Hello test")
	storage.StorageTest()

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
		// TODO: Add test cases.
		{
			name:    "first",
			args:    storage.MemStorage{Gauge: map[string]float64{}, Counter: map[string]int64{}},
			want:    want{statusCode: 200, data: "Метрика тип counter название PollCount обновлена 5\r\n"},
			request: "/update/counter/PollCount/5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.StorageTest()

			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			// создаём новый Recorder
			w := httptest.NewRecorder()
			h := http.HandlerFunc(handlers.UpdatedMem(&tt.args))
			h(w, request)

			res := w.Result()

			userResult, err := io.ReadAll(res.Body)
			fmt.Println(string(userResult), err)

			// проверяем код ответа
			assert.Equal(t, tt.want.statusCode, res.StatusCode)

			// проверяем данные ответа
			assert.Equal(t, tt.want.data, string(userResult))
		})
	}
}
