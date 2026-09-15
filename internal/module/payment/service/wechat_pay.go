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
	neturl "net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type WechatPayConfig struct {
	AppID string
	MchID string
	// Key 商户 API 私钥（PEM 格式），用于请求签名
	Key string
	// APIv3Key 微信支付 APIv3 密钥（32 位字符串），用于回调报文解密。
	// 它与上面的商户私钥是两个完全不同的凭据，不可互相替代。
	APIv3Key  string
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
	nonceStr, err := generateNonceStr()
	if err != nil {
		return nil, err
	}
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

// 平台证书缓存。
//
// 回调验签每次都要用平台公钥，若每次都现拉 /v3/certificates，
// 等于每笔回调多一次带签名的 HTTPS 往返 —— 既拖慢回调，
// 也会把「获取证书失败」放大成「全部回调验签失败」。
var (
	certCacheMu   sync.RWMutex
	certCache     = make(map[string]*rsa.PublicKey)
	certCacheTime time.Time
)

// certCacheTTL 平台证书缓存时长。微信平台证书轮换周期远长于此，10 分钟是安全与性能的折中
const certCacheTTL = 10 * time.Minute

func (g *WechatPayGateway) getPlatformPublicKey(serial string) (*rsa.PublicKey, error) {
	if pub := cachedPlatformPublicKey(serial); pub != nil {
		return pub, nil
	}

	keys, err := g.fetchPlatformPublicKeys()
	if err != nil {
		// 拉取失败时退回旧缓存：宁可用可能过期的证书，也不要让回调整体中断
		if pub := cachedPlatformPublicKey(serial, true); pub != nil {
			return pub, nil
		}
		return nil, err
	}

	certCacheMu.Lock()
	certCache = keys
	certCacheTime = time.Now()
	certCacheMu.Unlock()

	pub, ok := keys[serial]
	if !ok {
		return nil, fmt.Errorf("certificate with serial %s not found", serial)
	}
	return pub, nil
}

// cachedPlatformPublicKey 读取缓存的平台公钥。
// allowStale 为 true 时忽略过期判断（用于拉取失败时降级）
func cachedPlatformPublicKey(serial string, allowStale ...bool) *rsa.PublicKey {
	certCacheMu.RLock()
	defer certCacheMu.RUnlock()

	stale := len(allowStale) > 0 && allowStale[0]
	if !stale && (certCacheTime.IsZero() || time.Since(certCacheTime) > certCacheTTL) {
		return nil
	}
	return certCache[serial]
}

func (g *WechatPayGateway) fetchPlatformPublicKeys() (map[string]*rsa.PublicKey, error) {
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

	keys := make(map[string]*rsa.PublicKey, len(certsResp.Data))
	for _, cert := range certsResp.Data {
		// Decrypt certificate using APIv3 key
		plaintext, err := g.decryptResource(
			cert.EncryptCertificate.Ciphertext,
			cert.EncryptCertificate.Nonce,
			cert.EncryptCertificate.AssociatedData,
		)
		if err != nil {
			// 单张证书解密失败（如新增了未知算法）不应中断整体，跳过即可
			continue
		}

		block, _ := pem.Decode(plaintext)
		if block == nil {
			continue
		}

		certObj, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}

		pubKey, ok := certObj.PublicKey.(*rsa.PublicKey)
		if !ok {
			continue
		}
		keys[cert.SerialNo] = pubKey
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("no usable platform certificate")
	}
	return keys, nil
}

func (g *WechatPayGateway) decryptResource(ciphertext, nonce, associatedData string) ([]byte, error) {
	// 回调解密用的是「APIv3 密钥」——商户在微信支付平台单独设置的 32 位字符串，
	// 而不是请求签名所用的商户私钥（config.Key）。二者混用会导致解密必然失败。
	keyBytes := []byte(g.config.APIv3Key)
	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("APIv3 密钥长度必须为 32 字节，当前 %d 字节", len(keyBytes))
	}

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

	// APIv3 报文的 nonce 为 12 字节，与 GCM 标准 nonce 长度一致
	if len(nonce) != aesGCM.NonceSize() {
		return nil, fmt.Errorf("nonce 长度必须为 %d 字节，当前 %d 字节", aesGCM.NonceSize(), len(nonce))
	}

	plaintext, err := aesGCM.Open(nil, []byte(nonce), ciphertextBytes, []byte(associatedData))
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

	authorization, err := g.generateAuthorization(method, url, string(body))
	if err != nil {
		return nil, fmt.Errorf("生成支付请求签名失败: %w", err)
	}
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

func (g *WechatPayGateway) generateAuthorization(method, url, body string) (string, error) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonceStr, err := generateNonceStr()
	if err != nil {
		return "", err
	}
	message := buildSignatureMessage(method, url, timestamp, nonceStr, body)

	pk, err := parsePrivateKey(g.config.Key)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256([]byte(message))
	sign, err := rsa.SignPKCS1v15(rand.Reader, pk, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",signature="%s",timestamp="%s",serial_no="%s"`,
		g.config.MchID, nonceStr, base64.StdEncoding.EncodeToString(sign), timestamp, g.config.SerialNo), nil
}

// buildSignatureMessage 构造微信支付 APIv3 的待签名串。
//
// 格式（每行末尾都要有 \n，含最后一行）：
//
//	HTTP方法\nURL路径(含query)\n时间戳\n随机串\n报文主体\n
//
// 两个易错点：URL 必须是去掉域名的路径；报文主体必须参与签名 ——
// 漏掉 body 会让所有带请求体的调用（下单/退款）签名校验失败。
func buildSignatureMessage(method, url, timestamp, nonce, body string) string {
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n", method, canonicalURL(url), timestamp, nonce, body)
}

// canonicalURL 取参与签名的 URL：即绝对地址去掉协议与域名后的「路径 + query」。
//
// 微信 APIv3 要求签名串里的 URL 不含域名（/v3/pay/transactions/native 而非完整 https:// 地址），
// 传入完整地址会导致签名比对失败，表现为所有主动请求（下单/退款/查询）返回 401。
// 解析失败时退回原值，不至于因为格式问题让请求彻底发不出去。
func canonicalURL(rawURL string) string {
	u, err := neturl.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	if u.RequestURI() == "" {
		return rawURL
	}
	return u.RequestURI()
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
