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
	"log"
	"os"
)

// EncodeRSAAES кодирование слайса байт данных с помощью публичного rsa pem ключа
func EncodeRSAAES(text []byte, fileLocationPublic string) ([]byte, error) {
	// -- общие данные
	// Генерация случайного AES-ключа
	aesKey := generateAESKey()

	// Генерация случайного nonce для режима шифрования GCM
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		fmt.Println("Ошибка генерации nonce:", err)
		os.Exit(1)
	}

	// Загрузка открытого ключа RSA из файла
	publicKeyData, err := os.ReadFile(fileLocationPublic)
	if err != nil {
		fmt.Println("Ошибка чтения открытого ключа:", err)
		os.Exit(1)
	}
	publicKey, err := parseRSAPublicKey(publicKeyData)
	if err != nil {
		fmt.Println("Ошибка парсинга открытого ключа:", err)
		os.Exit(1)
	}

	// Шифрование AES-ключа с использованием RSA
	encryptedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, aesKey, nil)
	if err != nil {
		fmt.Println("Ошибка шифрования AES-ключа:", err)
		os.Exit(1)
	}

	// Шифрование строки с использованием AES
	encryptedData, err := encryptAES([]byte(text), aesKey, nonce)
	if err != nil {
		fmt.Println("Ошибка шифрования данных:", err)
		os.Exit(1)
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
		fmt.Println("Ошибка чтения закрытого ключа:", err)
		os.Exit(1)
	}
	privateKey, err := parseRSAPrivateKey(privateKeyData)
	if err != nil {
		fmt.Println("Ошибка парсинга закрытого ключа:", err)
		os.Exit(1)
	}

	// Расшифровка AES-ключа с использованием RSA. data[1] = encrypted Key
	decryptedKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, data[1], nil)
	if err != nil {
		fmt.Println("Ошибка расшифровки AES-ключа:", err)
		os.Exit(1)
	}

	// Расшифровка строки с использованием AES
	decryptedData, err := decryptAES(data[0], decryptedKey, data[2])
	if err != nil {
		fmt.Println("Ошибка расшифровки данных:", err)
		os.Exit(1)
	}
	return decryptedData, nil
}

// Генерация случайного AES-ключа
func generateAESKey() []byte {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		fmt.Println("Ошибка генерации AES-ключа:", err)
		os.Exit(1)
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
		return nil, fmt.Errorf("ошибка создания GCM режима для AES: %w", err)
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
		return nil, fmt.Errorf("ошибка создания AES-шифратора: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания GCM режима для AES: %w", err)
	}

	// Извлечение nonce из зашифрованных данных
	nonceSize := aesgcm.NonceSize()
	realNonce := ciphertext[:nonceSize]

	// Расшифровка данных
	plaintext, err := aesgcm.Open(nil, realNonce, ciphertext[nonceSize:], nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка расшифровки данных: %w", err)
	}

	return plaintext, nil
}

// Парсинг открытого ключа RSA
func parseRSAPublicKey(keyData []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("открытый ключ невалиден")
	}
	pkey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	rsaKey, ok := pkey.(*rsa.PublicKey)
	if !ok {
		log.Fatalf("got unexpected key type: %T", pkey)
	}
	return rsaKey, nil
}

// Парсинг закрытого ключа RSA
func parseRSAPrivateKey(keyData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("закрытый ключ невалиден")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
