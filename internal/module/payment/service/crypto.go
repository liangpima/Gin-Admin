package service

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
)

const nonceChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// generateNonceStr 生成 32 位随机字符串，用作微信支付签名中的 nonce_str。
//
// 必须使用 crypto/rand：nonce 参与签名，而 math/rand 的伪随机序列
// （尤其是以时间戳播种时）是可预测的，不适合用于支付场景。
func generateNonceStr() (string, error) {
	buf := make([]byte, 32)
	poolSize := big.NewInt(int64(len(nonceChars)))

	for i := range buf {
		idx, err := rand.Int(rand.Reader, poolSize)
		if err != nil {
			return "", fmt.Errorf("生成 nonce 失败: %w", err)
		}
		buf[i] = nonceChars[idx.Int64()]
	}

	return string(buf), nil
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
