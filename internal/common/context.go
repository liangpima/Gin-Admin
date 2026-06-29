package common

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// NormalizeIP 将 IPv6 回环地址转为 IPv4 格式，提升可读性
func NormalizeIP(ip string) string {
	if ip == "::1" || ip == "::ffff:127.0.0.1" {
		return "127.0.0.1"
	}
	// 处理 ::ffff:x.x.x.x 格式
	if strings.HasPrefix(ip, "::ffff:") {
		return ip[7:]
	}
	return ip
}

func GetTenantID(c *gin.Context) uint {
	if id, exists := c.Get(ContextKeyTenantID); exists {
		if v, ok := id.(uint); ok {
			return v
		}
	}
	return 0
}

func GetCurrentUserID(c *gin.Context) uint {
	if id, exists := c.Get(ContextKeyUserID); exists {
		if v, ok := id.(uint); ok {
			return v
		}
	}
	return 0
}

func GetCurrentUsername(c *gin.Context) string {
	if name, exists := c.Get(ContextKeyUsername); exists {
		if v, ok := name.(string); ok {
			return v
		}
	}
	return ""
}

func GetDeptID(c *gin.Context) uint {
	if id, exists := c.Get(ContextKeyDeptID); exists {
		if v, ok := id.(uint); ok {
			return v
		}
	}
	return 0
}

func GetUintParam(c *gin.Context, key string) (uint, error) {
	val := c.Param(key)
	id, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func GetPageInfo(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return page, pageSize
}
