package service

import (
	"bytes"
	"errors"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/axelx/go-yandex-metrics/internal/logger"
)

func PrepareFloat64Data(data string) (float64, error) {
	f, err := strconv.ParseFloat(data, 64)
	if err != nil {
		return 0, errors.New("ошибка обработки параметра float64 data ")
	}
	return f, nil
}

func PrepareInt64Data(data string) (int64, error) {
	i, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		return 0, errors.New("ошибка обработки параметра int64 data ")
	}
	return i, nil
}

func Int64ToPointerInt64(i int64) *int64 {
	return &i
}

func Float64ToPointerFloat64(f float64) *float64 {
	return &f
}

func UnPointer[K int64 | float64](val *K) K {
	if val == nil {
		return 0
	}
	return *val
}

func ToPointer[K int64 | float64](val K) *K {
	return &val
}

// StringToInt
func StringToInt(strVar string) int {
	intVar, err := strconv.Atoi(strVar)
	if err != nil {
		logger.Error("Error StringToInt", "convert; about err: "+err.Error()+"string: "+string(strVar))
		return 0
	}
	return intVar
}

// StreamToByte
func StreamToByte(stream io.Reader) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.Bytes()
}

// GetIP - получаем IP адрес запуска метода
func GetIP() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		logger.Error("Error service.GetIP", "Произошла ошибка при получении интерфейсов: "+err.Error())
		return ""
	}

	resIP := ""

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			logger.Error("Error service.GetIP", "Произошла ошибка при получении адресов интерфейса:"+err.Error())
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					resIP = ipNet.IP.String()
				}
			}
		}
	}
	return resIP
}

// ISinTrustedNetwork - проверяем находится ли IP адрес в диапазоне доверенной сети
func ISinTrustedNetwork(checkIP, cidr string) bool {
	mask := strings.Split(cidr, "/")
	subnetBIT := StringToInt(mask[1])

	if mask[0] != "" && subnetBIT == 0 {
		return false
	}

	ip := net.ParseIP(checkIP)
	ipNet := net.IPNet{
		IP:   net.ParseIP(mask[0]),
		Mask: net.CIDRMask(subnetBIT, 32),
	}

	if ipNet.Contains(ip) {
		logger.Info("service.ISinTrustedNetwork", "IP-адрес находится в подсети CIDR")
	} else {
		logger.Info("service.ISinTrustedNetwork", "IP-адрес НЕ находится в подсети CIDR")
	}

	return ipNet.Contains(ip)
}
