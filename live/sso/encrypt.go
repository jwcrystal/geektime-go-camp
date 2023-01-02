package sso

import (
	"github.com/google/uuid"
	"strings"
)

// Encrypt 加密生成簽名
func Encrypt(clientId string) string {
	// 模擬簽名
	random := uuid.New().String()
	return clientId + ":" + random
}

// Decrypt 簽名解密驗證
func Decrypt(signature []byte) (clientId string, err error) {
	// 模擬解密
	decrypts := strings.Split(string(signature), ":")
	return decrypts[0], nil
}
