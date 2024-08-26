package crypto_test

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/sagoo-cloud/nexframe/utils/crypto"
	"os"
	"path/filepath"
	"testing"
)

const (
	testKey       = "0123456789abcdef0123456789abcdef" // 32 bytes for AES-256
	testPlaintext = "Hello, World! This is a test."
	customTestKey = "fedcba9876543210fedcba9876543210" // 32 bytes for AES-256
)

func TestMain(m *testing.M) {
	// 创建临时目录结构
	tempDir, err := os.MkdirTemp("", "aes-test")
	if err != nil {
		panic("Failed to create temp directory: " + err.Error())
	}
	defer os.RemoveAll(tempDir)

	// 创建 etc/encryption 子目录
	keyDir := filepath.Join(tempDir, "etc", "encryption")
	err = os.MkdirAll(keyDir, 0755)
	if err != nil {
		panic("Failed to create key directory: " + err.Error())
	}

	// 创建临时密钥文件
	keyFilePath := filepath.Join(keyDir, "aes")
	err = os.WriteFile(keyFilePath, []byte(testKey), 0600)
	if err != nil {
		panic("Failed to write test key file: " + err.Error())
	}

	// 设置环境变量，使测试使用我们创建的临时密钥文件
	os.Setenv("AES_KEY_FILE", keyFilePath)

	// 运行测试
	code := m.Run()

	// 退出
	os.Exit(code)
}

func TestEncryptDecrypt(t *testing.T) {
	// 测试使用默认密钥
	encrypted, err := crypto.EncryptString(testPlaintext)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	decrypted, err := crypto.DecryptString(encrypted)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if decrypted != testPlaintext {
		t.Errorf("Decrypted text doesn't match original. Got %s, want %s", decrypted, testPlaintext)
	}

	// 测试使用自定义密钥
	encryptedCustom, err := crypto.EncryptString(testPlaintext, customTestKey)
	if err != nil {
		t.Fatalf("Encryption with custom key failed: %v", err)
	}

	decryptedCustom, err := crypto.DecryptString(encryptedCustom, customTestKey)
	if err != nil {
		t.Fatalf("Decryption with custom key failed: %v", err)
	}

	if decryptedCustom != testPlaintext {
		t.Errorf("Decrypted text with custom key doesn't match original. Got %s, want %s", decryptedCustom, testPlaintext)
	}
}

func TestEncryptDecryptLargeData(t *testing.T) {
	largeData := make([]byte, 1000000) // 1 MB of data
	_, err := rand.Read(largeData)
	if err != nil {
		t.Fatalf("Failed to generate large random data: %v", err)
	}

	largeString := hex.EncodeToString(largeData)

	encrypted, err := crypto.EncryptString(largeString)
	if err != nil {
		t.Fatalf("Encryption of large data failed: %v", err)
	}

	decrypted, err := crypto.DecryptString(encrypted)
	if err != nil {
		t.Fatalf("Decryption of large data failed: %v", err)
	}

	if decrypted != largeString {
		t.Error("Decrypted large data doesn't match original")
	}
}

func TestInvalidKey(t *testing.T) {
	_, err := crypto.EncryptString(testPlaintext, "invalid-key")
	if err == nil {
		t.Error("Expected error with invalid key, got nil")
	}
}

func TestInvalidCiphertext(t *testing.T) {
	_, err := crypto.DecryptString("invalid-ciphertext")
	if err == nil {
		t.Error("Expected error with invalid ciphertext, got nil")
	}
}

func TestNewAes(t *testing.T) {
	// Test with default key
	aes, err := crypto.NewAes()
	if err != nil {
		t.Fatalf("Failed to create AES with default key: %v", err)
	}

	encrypted, err := aes.Encrypt(testPlaintext)
	if err != nil {
		t.Fatalf("Encryption with default key failed: %v", err)
	}

	decrypted, err := aes.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decryption with default key failed: %v", err)
	}

	if decrypted != testPlaintext {
		t.Errorf("Decrypted text doesn't match original. Got %s, want %s", decrypted, testPlaintext)
	}

	// Test with custom key
	customAes, err := crypto.NewAes(customTestKey)
	if err != nil {
		t.Fatalf("Failed to create AES with custom key: %v", err)
	}

	encryptedCustom, err := customAes.Encrypt(testPlaintext)
	if err != nil {
		t.Fatalf("Encryption with custom key failed: %v", err)
	}

	decryptedCustom, err := customAes.Decrypt(encryptedCustom)
	if err != nil {
		t.Fatalf("Decryption with custom key failed: %v", err)
	}

	if decryptedCustom != testPlaintext {
		t.Errorf("Decrypted text with custom key doesn't match original. Got %s, want %s", decryptedCustom, testPlaintext)
	}
}
