package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

func GetHashSHA256Base64(data []byte, key string) string {
	// создаём новый hash.Hash, вычисляющий контрольную сумму SHA-256
	h := hmac.New(sha256.New, []byte(key))
	// передаём байты для хеширования
	h.Write(data)
	// вычисляем хеш
	sEnc := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return sEnc
}
