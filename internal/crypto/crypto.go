package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"os"

	"github.com/axelx/go-yandex-metrics/internal/logger"
)

// Encode кодирование слайса байт с помощью публичного rsa pem ключа
func Encode(text []byte, fileLocationPublic string) ([]byte, error) {
	file, err := os.Open(fileLocationPublic)
	if err != nil {
		logger.Error("Crypto Encode", "Ошибка чтения файла приватного ключа: "+err.Error())
	}

	pubKeyBytes, err := io.ReadAll(file)
	if err != nil {
		logger.Error("Crypto Encode", "io.ReadAll(file): "+err.Error())
	}

	// Разбираем публичный ключ
	block, _ := pem.Decode(pubKeyBytes)
	if block == nil {
		logger.Error("Crypto Encode", "pem.Decode(pubKeyBytes): "+err.Error())
		return nil, err
	}

	pubKey, err := x509.ParsePKCS1PublicKey(block.Bytes)

	if err != nil {
		logger.Error("Crypto Encode", "Ошибка разбора публичного ключа: "+err.Error())
	}

	// Шифруем сообщение публичным ключом
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, []byte(text))
	if err != nil {
		logger.Error("Crypto Encode", "Ошибка шифрования сообщения: "+err.Error())
	}
	return ciphertext, nil
}

// Decode раскодирование слайса байт с помощью приватного rsa ключа
func Decode(encodeText []byte, fileLocationPrivate string) ([]byte, error) {
	file, err := os.Open(fileLocationPrivate)
	if err != nil {
		logger.Error("Crypto Decode", "Ошибка чтения файла приватного ключа: "+err.Error())
	}
	privateKeyData, err := io.ReadAll(file)
	if err != nil {
		logger.Error("Crypto Decode", "io.ReadAll(file) "+err.Error())
	}

	// Декодирование PEM блока приватного ключа
	block, _ := pem.Decode(privateKeyData)
	if block == nil {
		logger.Error("Crypto Decode", "Ошибка декодирования PEM блока приватного ключа: "+err.Error())
		return nil, err
	}

	// Парсинг приватного ключа RSA из DER формата
	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		logger.Error("Crypto Decode", "Ошибка парсинга приватного ключа: "+err.Error())
	}

	// Расшифровка зашифрованных данных с использованием приватного ключа RSA
	plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, privKey, encodeText)
	if err != nil {
		logger.Error("Crypto Decode", "Ошибка расшифровки:: "+err.Error())
	}

	return plaintext, nil
}
