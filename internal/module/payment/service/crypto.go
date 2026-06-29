package service

import (
	"encoding/base64"
	"math/rand"
	"time"
)

func generateNonceStr() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 32)
	for i := range result {
		result[i] = chars[r.Intn(len(chars))]
	}
	return string(result)
}

func base64EncodeStd(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func base64DecodeStd(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
