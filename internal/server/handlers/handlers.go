package handlers

import (
	"fmt"
	"net/http"
	"strings"
)

type Keeper interface {
	SetGauge(string, string) error
	SetCounter(string, string) error
}

func UpdatedMem(m Keeper) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		s := strings.Split(req.URL.String(), "/")

		// s[2] <ТИП_МЕТРИКИ>  s[3] <ИМЯ_МЕТРИКИ> s[4] <ЗНАЧЕНИЕ_МЕТРИКИ>
		if cap(s) != 5 {
			http.Error(res, "StatusNotFound", http.StatusNotFound)
			return
		}

		if s[2] == "gauge" {
			err := m.SetGauge(s[3], s[4])
			if err != nil {
				http.Error(res, fmt.Sprint(err), http.StatusBadRequest)
				return
			}
		} else if s[2] == "counter" {
			err := m.SetCounter(s[3], s[4])
			if err != nil {
				http.Error(res, fmt.Sprint(err), http.StatusBadRequest)
				return
			}
		} else {
			http.Error(res, "StatusBadRequest", http.StatusBadRequest)
			return
		}
		fmt.Println(m)

		body := fmt.Sprintf("Метрика тип %s название %s обновлена %v\r\n", s[2], s[3], s[4])
		res.Write([]byte(body))

		res.WriteHeader(http.StatusOK)
	}
}
