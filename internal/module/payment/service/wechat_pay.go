package service

import (
	"bytes"
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type WechatPayConfig struct {
	AppID     string
	MchID     string
	Key       string
	SerialNo  string
	NotifyURL string
}

type WechatPayGateway struct {
	config WechatPayConfig
	client *http.Client
}

func NewWechatPayGateway(cfg WechatPayConfig) *WechatPayGateway {
	return &WechatPayGateway{
		config: cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (g *WechatPayGateway) Prepay(ctx context.Context, orderNo, subject, body string, amount int64, openID string) (map[string]interface{}, error) {
	order := map[string]interface{}{
		"appid":        g.config.AppID,
		"mchid":        g.config.MchID,
		"description":  subject,
		"out_trade_no": orderNo,
		"notify_url":   g.config.NotifyURL,
		"amount": map[string]interface{}{
			"total":    amount,
			"currency": "CNY",
		},
	}

	apiURL := "https://api.mch.weixin.qq.com/v3/pay/transactions/native"
	if openID != "" {
		order["payer"] = map[string]interface{}{"openid": openID}
		apiURL = "https://api.mch.weixin.qq.com/v3/pay/transactions/jsapi"
	}

	bodyBytes, _ := json.Marshal(order)
	resp, err := g.doRequest("POST", apiURL, bodyBytes)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if openID == "" {
		return map[string]interface{}{"code_url": result["code_url"]}, nil
	}

	return g.generateJSAPIPayInfo(result["prepay_id"].(string))
}

func (g *WechatPayGateway) generateJSAPIPayInfo(prepayID string) (map[string]interface{}, error) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonceStr := generateNonceStr()
	packageStr := "prepay_id=" + prepayID

	message := fmt.Sprintf("%s\n%s\n%s\n%s\n", g.config.AppID, timestamp, nonceStr, packageStr)

	pk, err := parsePrivateKey(g.config.Key)
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256([]byte(message))
	sign, err := rsa.SignPKCS1v15(rand.Reader, pk, crypto.SHA256, hash[:])
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"appId":     g.config.AppID,
		"timeStamp": timestamp,
		"nonceStr":  nonceStr,
		"package":   packageStr,
		"signType":  "RSA",
		"paySign":   base64.StdEncoding.EncodeToString(sign),
	}, nil
}

func (g *WechatPayGateway) ParseNotify(body []byte, headers http.Header) (*PayNotifyResult, error) {
	result := &PayNotifyResult{
		Status:  "fail",
		RawData: string(body),
	}

	// 1. Verify signature
	timestamp := headers.Get("Wechatpay-Timestamp")
	nonce := headers.Get("Wechatpay-Nonce")
	signature := headers.Get("Wechatpay-Signature")
	serial := headers.Get("Wechatpay-Serial")

	if err := g.verifySignature(timestamp, nonce, string(body), signature, serial); err != nil {
		return nil, fmt.Errorf("wechatpay signature verify failed: %w", err)
	}

	// 2. Parse notification body
	var notifyBody struct {
		ID           string `json:"id"`
		CreateTime   string `json:"create_time"`
		ResourceType string `json:"resource_type"`
		EventType    string `json:"event_type"`
		Summary      string `json:"summary"`
		Resource     struct {
			Algorithm      string `json:"algorithm"`
			Ciphertext     string `json:"ciphertext"`
			AssociatedData string `json:"associated_data"`
			Nonce          string `json:"nonce"`
			OriginalType   string `json:"original_type"`
		} `json:"resource"`
	}
	if err := json.Unmarshal(body, &notifyBody); err != nil {
		return nil, fmt.Errorf("parse notify body failed: %w", err)
	}

	if notifyBody.Resource.Ciphertext == "" {
		return nil, fmt.Errorf("ciphertext is empty")
	}

	// 3. Decrypt resource
	plaintext, err := g.decryptResource(
		notifyBody.Resource.Ciphertext,
		notifyBody.Resource.Nonce,
		notifyBody.Resource.AssociatedData,
	)
	if err != nil {
		return nil, fmt.Errorf("decrypt resource failed: %w", err)
	}

	// 4. Parse decrypted data
	var decrypted struct {
		OutTradeNo    string `json:"out_trade_no"`
		TransactionID string `json:"transaction_id"`
		TradeState    string `json:"trade_state"`
		SuccessTime   string `json:"success_time"`
		Payer         struct {
			OpenID string `json:"openid"`
		} `json:"payer"`
		Amount struct {
			Total    int64  `json:"total"`
			PayerTotal int64 `json:"payer_total"`
			Currency string `json:"currency"`
		} `json:"amount"`
	}
	if err := json.Unmarshal(plaintext, &decrypted); err != nil {
		return nil, fmt.Errorf("parse decrypted data failed: %w", err)
	}

	result.OrderNo = decrypted.OutTradeNo
	result.TradeNo = decrypted.TransactionID
	result.Amount = decrypted.Amount.Total

	if decrypted.TradeState == "SUCCESS" {
		result.Status = "success"
		if t, err := time.Parse("2006-01-02T15:04:05-07:00", decrypted.SuccessTime); err == nil {
			result.PaidAt = &t
		}
	}

	return result, nil
}

func (g *WechatPayGateway) verifySignature(timestamp, nonce, body, signature, serial string) error {
	// Build message for verification
	message := fmt.Sprintf("%s\n%s\n%s\n", timestamp, nonce, body)

	sigBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("base64 decode signature failed: %w", err)
	}

	// Get platform public key (from cache or fetch)
	pubKey, err := g.getPlatformPublicKey(serial)
	if err != nil {
		return fmt.Errorf("get platform public key failed: %w", err)
	}

	hash := sha256.Sum256([]byte(message))
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash[:], sigBytes)
}

func (g *WechatPayGateway) getPlatformPublicKey(serial string) (*rsa.PublicKey, error) {
	// Fetch platform certificates from WeChat Pay API
	certsURL := "https://api.mch.weixin.qq.com/v3/certificates"
	resp, err := g.doRequest("GET", certsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch platform certificates failed: %w", err)
	}

	var certsResp struct {
		Data []struct {
			SerialNo string `json:"serial_no"`
			EncryptCertificate struct {
				Algorithm  string `json:"algorithm"`
				Ciphertext string `json:"ciphertext"`
				Nonce      string `json:"nonce"`
				AssociatedData string `json:"associated_data"`
			} `json:"encrypt_certificate"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp, &certsResp); err != nil {
		return nil, fmt.Errorf("parse certificates response failed: %w", err)
	}

	for _, cert := range certsResp.Data {
		if cert.SerialNo == serial {
			// Decrypt certificate using APIv3 key
			plaintext, err := g.decryptResource(
				cert.EncryptCertificate.Ciphertext,
				cert.EncryptCertificate.Nonce,
				cert.EncryptCertificate.AssociatedData,
			)
			if err != nil {
				return nil, fmt.Errorf("decrypt certificate failed: %w", err)
			}

			block, _ := pem.Decode(plaintext)
			if block == nil {
				return nil, fmt.Errorf("failed to decode PEM certificate")
			}

			certObj, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("parse certificate failed: %w", err)
			}

			pubKey, ok := certObj.PublicKey.(*rsa.PublicKey)
			if !ok {
				return nil, fmt.Errorf("not an RSA public key")
			}
			return pubKey, nil
		}
	}

	return nil, fmt.Errorf("certificate with serial %s not found", serial)
}

func (g *WechatPayGateway) decryptResource(ciphertext, nonce, associatedData string) ([]byte, error) {
	// APIv3 key is the WechatPay key (not the merchant private key)
	// In production, this should be configured separately
	// The key is 32 bytes base64-encoded APIv3 key from merchant platform
	apiKey := g.config.Key
	if len(apiKey) < 32 {
		return nil, fmt.Errorf("APIv3 key too short, need at least 32 bytes")
	}

	keyBytes := []byte(apiKey[:32])

	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("base64 decode ciphertext failed: %w", err)
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceBytes := []byte(nonce)
	aadBytes := []byte(associatedData)

	plaintext, err := aesGCM.Open(nil, nonceBytes, ciphertextBytes, aadBytes)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func (g *WechatPayGateway) doRequest(method, url string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(context.Background(), method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	authorization := g.generateAuthorization(method, url, string(body))
	req.Header.Set("Authorization", authorization)

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("wechatpay request failed(%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (g *WechatPayGateway) generateAuthorization(method, url, body string) string {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonceStr := generateNonceStr()
	message := fmt.Sprintf("%s\n%s\n%s\n%s\n", method, url, timestamp, nonceStr)

	pk, err := parsePrivateKey(g.config.Key)
	if err != nil {
		return ""
	}

	hash := sha256.Sum256([]byte(message))
	sign, _ := rsa.SignPKCS1v15(rand.Reader, pk, crypto.SHA256, hash[:])

	return fmt.Sprintf(`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",signature="%s",timestamp="%s",serial_no="%s"`,
		g.config.MchID, nonceStr, base64.StdEncoding.EncodeToString(sign), timestamp, g.config.SerialNo)
}

func parsePrivateKey(key string) (*rsa.PrivateKey, error) {
	key = strings.ReplaceAll(key, "-----BEGIN PRIVATE KEY-----", "")
	key = strings.ReplaceAll(key, "-----END PRIVATE KEY-----", "")
	key = strings.ReplaceAll(key, "\n", "")
	key = strings.TrimSpace(key)

	return jwt.ParseRSAPrivateKeyFromPEM([]byte("-----BEGIN PRIVATE KEY-----\n" + key + "\n-----END PRIVATE KEY-----"))
}

// WechatRefund applies refund via WeChat Pay V3 API
func (g *WechatPayGateway) Refund(ctx context.Context, orderNo, refundNo string, amount, refundAmount int64) error {
	body := map[string]interface{}{
		"out_trade_no":   orderNo,
		"out_request_no": refundNo,
		"notify_url":     g.config.NotifyURL,
		"amount": map[string]interface{}{
			"refund":   refundAmount,
			"total":    amount,
			"currency": "CNY",
		},
	}

	bodyBytes, _ := json.Marshal(body)
	_, err := g.doRequest("POST", "https://api.mch.weixin.qq.com/v3/refund/domestic/refunds", bodyBytes)
	return err
}

// WechatQueryRefund queries refund status
func (g *WechatPayGateway) WechatQueryRefund(ctx context.Context, refundNo string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.mch.weixin.qq.com/v3/refund/domestic/refunds/%s", refundNo)
	resp, err := g.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// WechatQueryOrder queries order status from WeChat
func (g *WechatPayGateway) WechatQueryOrder(ctx context.Context, orderNo string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.mch.weixin.qq.com/v3/pay/transactions/out-trade-no/%s?mchid=%s", orderNo, g.config.MchID)
	resp, err := g.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}
