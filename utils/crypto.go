package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateRandomString 生成指定字节长度的随机 hex 字符串
func GenerateRandomString(byteLen int) (string, error) {
	bytes := make([]byte, byteLen)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
