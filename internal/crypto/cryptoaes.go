package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"

	"github.com/axelx/go-yandex-metrics/internal/logger"
)

// EncodeRSAAES кодирование слайса байт данных с помощью публичного rsa pem ключа
func EncodeRSAAES(text []byte, fileLocationPublic string) ([]byte, error) {
	// -- общие данные
	// Генерация случайного AES-ключа
	aesKey := generateAESKey()

	// Генерация случайного nonce для режима шифрования GCM
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		logger.Log.Error("Crypto EncodeRSAAES", "Ошибка генерации nonce:"+err.Error())
		return nil, err
	}

	// Загрузка открытого ключа RSA из файла
	publicKeyData, err := os.ReadFile(fileLocationPublic)
	if err != nil {
		logger.Log.Error("Crypto EncodeRSAAES", "Ошибка чтения открытого ключа:"+err.Error())
		return nil, err
	}
	publicKey, err := parseRSAPublicKey(publicKeyData)
	if err != nil {
		logger.Log.Error("Crypto EncodeRSAAES", "Ошибка парсинга открытого ключа:"+err.Error())
		return nil, err
	}

	// Шифрование AES-ключа с использованием RSA
	encryptedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, aesKey, nil)
	if err != nil {
		logger.Log.Error("Crypto EncodeRSAAES", "Ошибка шифрования AES-ключа:"+err.Error())
		return nil, err
	}

	// Шифрование строки с использованием AES
	encryptedData, err := encryptAES([]byte(text), aesKey, nonce)
	if err != nil {
		logger.Log.Error("Crypto EncodeRSAAES", "Ошибка шифрования данных:"+err.Error())
		return nil, err
	}

	// Конвертируем [][]byte в []byte
	slices := [][]byte{encryptedData, encryptedKey, nonce}
	concatenated := bytes.Join(slices, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0})
	return concatenated, nil
}

// DecodeRSAAES раскодирование слайса байт с помощью приватного rsa ключа
func DecodeRSAAES(d []byte, fileLocationPrivate string) ([]byte, error) {
	data := bytes.Split(d, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0})
	// Загрузка закрытого ключа RSA из файла
	privateKeyData, err := os.ReadFile(fileLocationPrivate)
	if err != nil {
		logger.Log.Error("Crypto DecodeRSAAES", "Ошибка чтения закрытого ключа:"+err.Error())
		return nil, err
	}
	privateKey, err := parseRSAPrivateKey(privateKeyData)
	if err != nil {
		logger.Log.Error("Crypto DecodeRSAAES", "Ошибка парсинга закрытого ключа:"+err.Error())
		return nil, err
	}

	// Расшифровка AES-ключа с использованием RSA. data[1] = encrypted Key
	decryptedKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, data[1], nil)
	if err != nil {
		logger.Log.Error("Crypto DecodeRSAAES", "Ошибка расшифровки AES-ключа:"+err.Error())
		return nil, err
	}

	// Расшифровка строки с использованием AES
	decryptedData, err := decryptAES(data[0], decryptedKey, data[2])
	if err != nil {
		logger.Log.Error("Crypto DecodeRSAAES", "Ошибка расшифровки данных:"+err.Error())
		return nil, err
	}
	return decryptedData, nil
}

// Генерация случайного AES-ключа
func generateAESKey() []byte {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		logger.Log.Error("Crypto DecodeRSAAES", "Ошибка генерации AES-ключа:"+err.Error())
		return nil
	}
	return key
}

// Шифрование данных с использованием AES
func encryptAES(data []byte, key []byte, nonce []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания AES-шифратора: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		logger.Log.Error("Crypto DecodeRSAAES", "ошибка создания GCM режима для AES: "+err.Error())
		return nil, err
	}

	// Шифрование данных
	ciphertext := aesgcm.Seal(nil, nonce, data, nil)

	// Добавление nonce к зашифрованным данным
	ciphertext = append(nonce, ciphertext...)

	return ciphertext, nil
}

// Расшифровка данных с использованием AES
func decryptAES(ciphertext []byte, key []byte, nonce []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		logger.Log.Error("Crypto DecodeRSAAES", "ошибка создания AES-шифратора: "+err.Error())
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		logger.Log.Error("Crypto DecodeRSAAES", "ошибка создания GCM режима для AES: "+err.Error())
		return nil, err
	}

	// Извлечение nonce из зашифрованных данных
	nonceSize := aesgcm.NonceSize()
	realNonce := ciphertext[:nonceSize]

	// Расшифровка данных
	plaintext, err := aesgcm.Open(nil, realNonce, ciphertext[nonceSize:], nil)
	if err != nil {
		logger.Log.Error("Crypto DecodeRSAAES", "ошибка расшифровки данных:"+err.Error())
		return nil, err
	}

	return plaintext, nil
}

// Парсинг открытого ключа RSA
func parseRSAPublicKey(keyData []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "PUBLIC KEY" {
		logger.Log.Error("Crypto parseRSAPublicKey", "открытый ключ невалиден")
		return nil, fmt.Errorf("открытый ключ невалиден")
	}
	pkey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		logger.Log.Error("Crypto parseRSAPublicKey", "ParsePKIXPublicKey: "+err.Error())
		return nil, err
	}

	rsaKey, ok := pkey.(*rsa.PublicKey)
	if !ok {
		logger.Log.Error("Crypto", "got unexpected key type: %T")
		return nil, err
	}
	return rsaKey, nil
}

// Парсинг закрытого ключа RSA
func parseRSAPrivateKey(keyData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		logger.Log.Error("Crypto ", "закрытый ключ невалиден")
		return nil, fmt.Errorf("закрытый ключ невалиден")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
