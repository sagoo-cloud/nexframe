package crypto

import (
	"crypto/sha256"
	"encoding/hex"
)

// Sha256 计算给定字节切片的 SHA-256 哈希值，并返回其十六进制表示
func Sha256(b []byte) string {
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:])
}

// Sha256String 计算给定字符串的 SHA-256 哈希值，并返回其十六进制表示
func Sha256String(s string) string {
	return Sha256([]byte(s))
}

// Sha256WithError 计算给定字节切片的 SHA-256 哈希值，返回其十六进制表示和可能的错误
func Sha256WithError(b []byte) (string, error) {
	return Sha256(b), nil
}
