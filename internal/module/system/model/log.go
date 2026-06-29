package model

import (
	"time"
)

type SysOperationLog struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	TenantID       uint      `gorm:"index;comment:租户ID" json:"tenantId"`
	Title          string    `gorm:"type:varchar(64);comment:模块标题" json:"title"`
	Action         string    `gorm:"type:varchar(64);comment:操作类型" json:"action"`
	Method         string    `gorm:"type:varchar(200);comment:请求方法" json:"method"`
	RequestMethod  string    `gorm:"type:varchar(10);comment:HTTP方法" json:"requestMethod"`
	RequestURL     string    `gorm:"type:varchar(500);comment:请求URL" json:"requestUrl"`
	RequestParam   string    `gorm:"type:text;comment:请求参数" json:"requestParam"`
	ResponseResult string    `gorm:"type:text;comment:返回结果" json:"responseResult"`
	Status         int8      `gorm:"type:tinyint;default:1;comment:状态 0失败 1成功" json:"status"`
	ErrorMsg       string    `gorm:"type:text;comment:错误消息" json:"errorMsg"`
	IP             string    `gorm:"type:varchar(128);comment:操作IP" json:"ip"`
	Location       string    `gorm:"type:varchar(255);comment:操作地点" json:"location"`
	UserAgent      string    `gorm:"type:varchar(500);comment:浏览器UA" json:"userAgent"`
	OperatorID     uint      `gorm:"comment:操作人ID" json:"operatorId"`
	OperatorName   string    `gorm:"type:varchar(64);comment:操作人名称" json:"operatorName"`
	CostTime       int64     `gorm:"comment:耗时(ms)" json:"costTime"`
	CreatedAt      time.Time `gorm:"comment:创建时间" json:"createdAt"`
}

func (SysOperationLog) TableName() string {
	return "sys_operation_log"
}

type SysLoginLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	TenantID  uint      `gorm:"index;comment:租户ID" json:"tenantId"`
	Username  string    `gorm:"type:varchar(64);comment:用户名" json:"username"`
	IP        string    `gorm:"type:varchar(128);comment:登录IP" json:"ip"`
	Location  string    `gorm:"type:varchar(255);comment:登录地点" json:"location"`
	Browser   string    `gorm:"type:varchar(128);comment:浏览器" json:"browser"`
	OS        string    `gorm:"type:varchar(128);comment:操作系统" json:"os"`
	Status    int8      `gorm:"type:tinyint;default:1;comment:状态 0失败 1成功" json:"status"`
	Msg       string    `gorm:"type:varchar(255);comment:消息" json:"msg"`
	LoginTime time.Time `gorm:"comment:登录时间" json:"loginTime"`
}

func (SysLoginLog) TableName() string {
	return "sys_login_log"
}
