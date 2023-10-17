package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"

	"github.com/axelx/go-yandex-metrics/internal/logger"
)

// Encode кодирование слайса байт с помощью публичного rsa pem ключа
func Encode(text []byte, fileLocation string) []byte {

	// Читаем публичный ключ из файла
	pubKeyLocation := fileLocation
	pubKeyBytes, err := ioutil.ReadFile(pubKeyLocation)
	if err != nil {
		logger.Log.Error("Crypto Encode", "Ошибка чтения файла с публичным ключом: "+err.Error())
	}

	// Разбираем публичный ключ
	block, _ := pem.Decode(pubKeyBytes)
	if block == nil {
		logger.Log.Error("Crypto Encode", "pem.Decode(pubKeyBytes): "+err.Error())
	}

	pubKey, err := x509.ParsePKCS1PublicKey(block.Bytes)

	if err != nil {
		logger.Log.Error("Crypto Encode", "Ошибка разбора публичного ключа: "+err.Error())
	}

	// Шифруем сообщение публичным ключом
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, []byte(text))
	if err != nil {
		logger.Log.Error("Crypto Encode", "Ошибка шифрования сообщения: "+err.Error())
	}
	return ciphertext
}

// Decode раскодирование слайса байт с помощью приватного rsa ключа
func Decode(encodeText []byte, privateKeyFile string) []byte {
	//privateKeyFile := "id_rsa_ya.pem"
	privateKeyData, err := ioutil.ReadFile(privateKeyFile)
	if err != nil {
		logger.Log.Error("Crypto Decode", "Ошибка чтения файла приватного ключа: "+err.Error())
	}

	// Декодирование PEM блока приватного ключа
	block, _ := pem.Decode(privateKeyData)
	if block == nil {
		logger.Log.Error("Crypto Decode", "Ошибка декодирования PEM блока приватного ключа: "+err.Error())
	}

	// Парсинг приватного ключа RSA из DER формата
	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		logger.Log.Error("Crypto Decode", "Ошибка парсинга приватного ключа: "+err.Error())
	}

	// Расшифровка зашифрованных данных с использованием приватного ключа RSA
	plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, privKey, encodeText)
	if err != nil {
		logger.Log.Error("Crypto Decode", "Ошибка расшифровки:: "+err.Error())
	}

	return plaintext
}
