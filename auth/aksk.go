package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
)

// VerifySignature 验证签名
func VerifySignature(ak, sk, timeStr, sign string) bool {
	timestamp, err := strconv.ParseInt(timeStr, 10, 64) // 时间戳
	if err != nil {
		// 处理解析错误
		return false
	}
	// 计算签名
	message := "ak=" + ak + "&time=" + strconv.FormatInt(timestamp, 10) // 消息
	newSign := GenerateSignature(message, sk)
	// 使用 constant time comparison 避免潜在的时间攻击
	return hmac.Equal([]byte(sign), []byte(newSign))
}

// GenerateSignature 生成签名
func GenerateSignature(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}
