package service

import "testing"

// TestCanonicalURL 校验参与签名的 URL 只保留「路径 + query」。
//
// 微信 APIv3 要求签名串中的 URL 不含协议与域名。若误传完整地址，
// 下单/退款/查询会一律返回 401，且失败会被当作 PayInfo["payError"] 静默吞掉。
func TestCanonicalURL(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"完整地址取路径", "https://api.mch.weixin.qq.com/v3/pay/transactions/native", "/v3/pay/transactions/native"},
		{"保留查询串", "https://api.mch.weixin.qq.com/v3/pay/transactions/out-trade-no/PAY123?mchid=1900000001",
			"/v3/pay/transactions/out-trade-no/PAY123?mchid=1900000001"},
		{"本身就是路径则原样返回", "/v3/refund/domestic/refunds", "/v3/refund/domestic/refunds"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := canonicalURL(c.in); got != c.want {
				t.Errorf("canonicalURL(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

// TestBuildSignatureMessage 校验待签名串的构造规则。
//
// 这是本次修复的核心：早前签名串里根本没有报文主体，导致所有带请求体的
// 调用（下单/退款）签名必然不匹配。
func TestBuildSignatureMessage(t *testing.T) {
	const (
		url  = "https://api.mch.weixin.qq.com/v3/pay/transactions/native"
		body = `{"amount":{"total":1}}`
	)

	got := buildSignatureMessage("POST", url, "1700000000", "nonce", body)
	want := "POST\n/v3/pay/transactions/native\n1700000000\nnonce\n" + body + "\n"
	if got != want {
		t.Errorf("签名串不符合 APIv3 规范:\n got: %q\nwant: %q", got, want)
	}

	// GET 无请求体时，主体部分为空串，但结尾的 \n 仍要保留
	got = buildSignatureMessage("GET", "https://api.mch.weixin.qq.com/v3/certificates", "1700000000", "nonce", "")
	want = "GET\n/v3/certificates\n1700000000\nnonce\n\n"
	if got != want {
		t.Errorf("GET 签名串不正确:\n got: %q\nwant: %q", got, want)
	}

	// 主体不同则签名串必须不同（防止 body 又被漏掉）
	a := buildSignatureMessage("POST", url, "1", "n", `{"a":1}`)
	b := buildSignatureMessage("POST", url, "1", "n", `{"a":2}`)
	if a == b {
		t.Error("报文主体未参与签名：不同 body 得到了相同的待签名串")
	}
}
