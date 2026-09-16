package common

import (
	"net/http"

	"go-admin/internal/logger"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func SuccessWithPage(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data: PageData{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
}

func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

func ErrorWithHttpStatus(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    401,
		Message: message,
		Data:    nil,
	})
}

func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Code:    403,
		Message: message,
		Data:    nil,
	})
}

// FailWith 输出 handler 中的错误，并按错误语义选择正确的业务码：
//
//   - 业务错误（BizError）：按其 Code 返回 400 / 404，原文案透出，用户看得懂也改得了
//   - 系统错误（DB、IO、签名失败等）：返回 500 + 通用文案，真实错误只记日志
//
// 之所以对系统错误隐藏细节：原始 error 常带 SQL、表名字段、内部路径等信息，
// 直接回给调用方等于泄漏实现细节；而排查所需的信息日志里已经有了。
func FailWith(c *gin.Context, err error) {
	if err == nil {
		return
	}

	if be, ok := AsBizError(err); ok {
		Error(c, be.Code, be.Msg)
		return
	}

	// logger.Log 初值为 no-op，永远不为 nil，无需空值保护
	logger.Log.Errorf("[internal] %s %s -> %v",
		c.Request.Method, c.Request.URL.Path, err)
	Error(c, CodeInternalError, "服务器内部错误")
}
