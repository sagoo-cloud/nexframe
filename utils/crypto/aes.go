package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const defaultAesKeyFileName = "etc/encryption/aes"

var (
	defaultKey string
	once       sync.Once
)

func getDefaultKeyFilePath() string {
	execPath, err := os.Executable()
	if err != nil {
		panic("Failed to get executable path: " + err.Error())
	}
	return filepath.Join(filepath.Dir(execPath), defaultAesKeyFileName)
}

func getDefaultKey() string {
	once.Do(func() {
		keyFile := os.Getenv("AES_KEY_FILE")
		if keyFile == "" {
			keyFile = getDefaultKeyFilePath()
		}
		keyByte, err := os.ReadFile(keyFile)
		if err != nil {
			panic("Failed to read AES key from file: " + err.Error())
		}
		defaultKey = strings.TrimSpace(string(keyByte))
	})
	return defaultKey
}

// NewAes creates a new AES client with the given key or uses the default key if not provided
func NewAes(key ...string) (*Aes, error) {
	var keyToUse string
	if len(key) > 0 && key[0] != "" {
		keyToUse = key[0]
	} else {
		keyToUse = getDefaultKey()
	}

	block, err := aes.NewCipher([]byte(keyToUse))
	if err != nil {
		return nil, err
	}
	return &Aes{block: block}, nil
}

// Aes represents an AES encryption/decryption client
type Aes struct {
	block cipher.Block
}

// Encrypt encrypts the plaintext
func (a *Aes) Encrypt(plaintext string) (string, error) {
	cipherData := make([]byte, aes.BlockSize+len(plaintext))
	iv := cipherData[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(a.block, iv)
	stream.XORKeyStream(cipherData[aes.BlockSize:], []byte(plaintext))

	return hex.EncodeToString(cipherData), nil
}

// Decrypt decrypts the ciphertext
func (a *Aes) Decrypt(d string) (string, error) {
	cipherData, err := hex.DecodeString(d)
	if err != nil {
		return "", err
	}
	if len(cipherData) < aes.BlockSize {
		return "", errors.New("cipherData too short")
	}

	iv := cipherData[:aes.BlockSize]
	cipherData = cipherData[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(a.block, iv)
	stream.XORKeyStream(cipherData, cipherData)

	return string(cipherData), nil
}

// EncryptString encrypts a string using either the provided key or the default key
func EncryptString(plaintext string, key ...string) (string, error) {
	newAes, err := NewAes(key...)
	if err != nil {
		return "", err
	}
	return newAes.Encrypt(plaintext)
}

// DecryptString decrypts a string using either the provided key or the default key
func DecryptString(ciphertext string, key ...string) (string, error) {
	newAes, err := NewAes(key...)
	if err != nil {
		return "", err
	}
	return newAes.Decrypt(ciphertext)
}
