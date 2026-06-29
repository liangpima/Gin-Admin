package service

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type AlipayConfig struct {
	AppID       string
	PrivateKey  string
	NotifyURL   string
	ReturnURL   string
	PublicKeyID string
}

type AlipayGateway struct {
	config AlipayConfig
	client *http.Client
}

func NewAlipayGateway(cfg AlipayConfig) *AlipayGateway {
	return &AlipayGateway{
		config: cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (g *AlipayGateway) Prepay(ctx context.Context, orderNo, subject string, amount int64, returnURL string) (map[string]interface{}, error) {
	amountStr := fmt.Sprintf("%.2f", float64(amount)/100)

	params := map[string]string{
		"app_id":     g.config.AppID,
		"method":     "alipay.trade.page.pay",
		"charset":    "utf-8",
		"sign_type":  "RSA2",
		"timestamp":  time.Now().Format("2006-01-02 15:04:05"),
		"version":    "1.0",
		"notify_url": g.config.NotifyURL,
	}

	// 验证 returnURL 防止开放重定向
	if returnURL != "" {
		if err := validateReturnURL(returnURL); err != nil {
			return nil, err
		}
		params["return_url"] = returnURL
	}

	bizContent := map[string]interface{}{
		"out_trade_no": orderNo,
		"total_amount": amountStr,
		"subject":      subject,
		"product_code": "FAST_INSTANT_TRADE_PAY",
	}
	bizBytes, _ := json.Marshal(bizContent)
	params["biz_content"] = string(bizBytes)

	sign, err := g.sign(params)
	if err != nil {
		return nil, err
	}
	params["sign"] = sign

	formData := buildFormQuery(params)
	formURL := "https://openapi.alipay.com/gateway.do?" + formData

	return map[string]interface{}{
		"form_url": formURL,
		"method":   "redirect",
	}, nil
}

func (g *AlipayGateway) PrepayApp(ctx context.Context, orderNo, subject string, amount int64) (map[string]interface{}, error) {
	amountStr := fmt.Sprintf("%.2f", float64(amount)/100)

	params := map[string]string{
		"app_id":     g.config.AppID,
		"method":     "alipay.trade.app.pay",
		"charset":    "utf-8",
		"sign_type":  "RSA2",
		"timestamp":  time.Now().Format("2006-01-02 15:04:05"),
		"version":    "1.0",
		"notify_url": g.config.NotifyURL,
	}

	bizContent := map[string]interface{}{
		"out_trade_no": orderNo,
		"total_amount": amountStr,
		"subject":      subject,
	}
	bizBytes, _ := json.Marshal(bizContent)
	params["biz_content"] = string(bizBytes)

	sign, err := g.sign(params)
	if err != nil {
		return nil, err
	}
	params["sign"] = sign

	return map[string]interface{}{
		"order_string": buildFormQuery(params),
	}, nil
}

func (g *AlipayGateway) QueryTrade(ctx context.Context, orderNo string) (map[string]interface{}, error) {
	params := map[string]string{
		"app_id":    g.config.AppID,
		"method":    "alipay.trade.query",
		"charset":   "utf-8",
		"sign_type": "RSA2",
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"version":   "1.0",
	}

	bizContent := map[string]interface{}{
		"out_trade_no": orderNo,
	}
	bizBytes, _ := json.Marshal(bizContent)
	params["biz_content"] = string(bizBytes)

	sign, err := g.sign(params)
	if err != nil {
		return nil, err
	}
	params["sign"] = sign

	respBody, err := g.doRequest("POST", "https://openapi.alipay.com/gateway.do", params)
	if err != nil {
		return nil, err
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (g *AlipayGateway) Refund(ctx context.Context, orderNo, refundNo string, amount int64) (map[string]interface{}, error) {
	amountStr := fmt.Sprintf("%.2f", float64(amount)/100)

	params := map[string]string{
		"app_id":     g.config.AppID,
		"method":     "alipay.trade.refund",
		"charset":    "utf-8",
		"sign_type":  "RSA2",
		"timestamp":  time.Now().Format("2006-01-02 15:04:05"),
		"version":    "1.0",
		"notify_url": g.config.NotifyURL,
	}

	bizContent := map[string]interface{}{
		"out_trade_no":   orderNo,
		"refund_amount":  amountStr,
		"out_request_no": refundNo,
	}
	bizBytes, _ := json.Marshal(bizContent)
	params["biz_content"] = string(bizBytes)

	sign, err := g.sign(params)
	if err != nil {
		return nil, err
	}
	params["sign"] = sign

	respBody, err := g.doRequest("POST", "https://openapi.alipay.com/gateway.do", params)
	if err != nil {
		return nil, err
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (g *AlipayGateway) ParseNotify(body []byte) (*PayNotifyResult, error) {
	result := &PayNotifyResult{
		Status:  "fail",
		RawData: string(body),
	}

	form, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, err
	}

	sign := form.Get("sign")

	params := make(map[string]string)
	for k := range form {
		if k != "sign" && k != "sign_type" {
			params[k] = form.Get(k)
		}
	}

	if err := g.verify(params, sign); err != nil {
		return nil, fmt.Errorf("alipay verify failed: %w", err)
	}

	result.TradeNo = form.Get("trade_no")
	result.OrderNo = form.Get("out_trade_no")

	// 解析实际支付金额（元转分）
	if totalAmount := form.Get("total_amount"); totalAmount != "" {
		var amt float64
		fmt.Sscanf(totalAmount, "%f", &amt)
		result.Amount = int64(amt * 100)
	}

	tradeStatus := form.Get("trade_status")
	if tradeStatus == "TRADE_SUCCESS" || tradeStatus == "TRADE_FINISHED" {
		result.Status = "success"
		gmtPayStr := form.Get("gmt_payment")
		if t, err := time.Parse("2006-01-02 15:04:05", gmtPayStr); err == nil {
			result.PaidAt = &t
		}
	}

	return result, nil
}

func (g *AlipayGateway) sign(params map[string]string) (string, error) {
	sortedKeys := make([]string, 0, len(params))
	for k := range params {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	var pairs []string
	for _, k := range sortedKeys {
		if params[k] != "" {
			pairs = append(pairs, k+"="+params[k])
		}
	}
	content := strings.Join(pairs, "&")

	pk, err := parsePrivateKey(g.config.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("parse private key failed: %w", err)
	}

	hash := sha256.Sum256([]byte(content))
	signature, err := rsa.SignPKCS1v15(rand.Reader, pk, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}

	return base64EncodeStd(signature), nil
}

func (g *AlipayGateway) verify(params map[string]string, signature string) error {
	sortedKeys := make([]string, 0, len(params))
	for k := range params {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	var pairs []string
	for _, k := range sortedKeys {
		if params[k] != "" {
			pairs = append(pairs, k+"="+params[k])
		}
	}
	content := strings.Join(pairs, "&")

	sigBytes, err := base64DecodeStd(signature)
	if err != nil {
		return err
	}

	publicKey, err := g.getPublicKey()
	if err != nil {
		return err
	}

	hash := sha256.Sum256([]byte(content))
	return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], sigBytes)
}

func (g *AlipayGateway) getPublicKey() (*rsa.PublicKey, error) {
	key := g.config.PublicKeyID
	if key == "" {
		return nil, fmt.Errorf("alipay public key not configured, set oss config pay.alipay_public_key")
	}

	key = strings.ReplaceAll(key, "-----BEGIN PUBLIC KEY-----", "")
	key = strings.ReplaceAll(key, "-----END PUBLIC KEY-----", "")
	key = strings.ReplaceAll(key, "\n", "")
	key = strings.TrimSpace(key)

	return parsePublicKey(key)
}

func (g *AlipayGateway) doRequest(method, requestURL string, params map[string]string) ([]byte, error) {
	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, requestURL, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func buildFormQuery(params map[string]string) string {
	sortedKeys := make([]string, 0, len(params))
	for k := range params {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	var pairs []string
	for _, k := range sortedKeys {
		if params[k] != "" {
			pairs = append(pairs, k+"="+url.QueryEscape(params[k]))
		}
	}
	return strings.Join(pairs, "&")
}

func parsePublicKey(key string) (*rsa.PublicKey, error) {
	pemStr := "-----BEGIN PUBLIC KEY-----\n" + key + "\n-----END PUBLIC KEY-----"

	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPub, nil
}

// validateReturnURL 校验 returnURL，防止开放重定向攻击
func validateReturnURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("无效的回调地址")
	}

	// 只允许 http/https 协议
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("回调地址只支持 http/https 协议")
	}

	// 禁止 javascript: data: 等危险协议
	scheme := strings.ToLower(u.Scheme)
	if scheme == "javascript" || scheme == "data" || scheme == "vbscript" {
		return fmt.Errorf("不支持的协议类型")
	}

	// 主机名不能为空
	if u.Host == "" {
		return fmt.Errorf("回调地址缺少主机名")
	}

	return nil
}
