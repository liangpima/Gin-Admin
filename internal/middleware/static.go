package middleware

import (
	"github.com/gin-gonic/gin"
)

// UploadSecurity 为静态上传目录补充安全响应头。
//
// 上传目录对匿名访问开放（前端需要直链展示图片），但白名单里包含 .svg，
// 而 SVG 可以内嵌 <script>，在同源自下访问等同于存储型 XSS。
//
// 这里不改动文件的 Content-Type（否则图片会变成下载），而是用 CSP 把
// 「把上传文件当作文档打开」的行为限制住：
//   - sandbox：以文档形式打开时进入沙箱，脚本、表单、同源访问全部禁用
//   - default-src 'none'：禁止加载任何外部子资源
//   - nosniff：禁止浏览器嗅探 MIME，避免把 text/plain 当 HTML 执行
//
// 图片以 <img src> 形式引用不受 CSP sandbox 影响，正常展示。
func UploadSecurity() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Content-Security-Policy",
			"default-src 'none'; img-src 'self' data:; sandbox")
		c.Next()
	}
}
