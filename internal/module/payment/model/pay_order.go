package model

import (
	"time"

	"go-admin/internal/common"

	"gorm.io/gorm"
)

type PayOrder struct {
	common.BaseModel
	TenantID    uint           `gorm:"index;comment:租户ID" json:"tenantId"`
	OrderNo     string         `gorm:"type:varchar(64);uniqueIndex;comment:订单号" json:"orderNo"`
	TradeNo     string         `gorm:"type:varchar(128);index;comment:第三方交易号" json:"tradeNo"`
	Subject     string         `gorm:"type:varchar(200);comment:订单标题" json:"subject"`
	Body        string         `gorm:"type:varchar(500);comment:订单描述" json:"body"`
	Amount      int64          `gorm:"comment:金额(分)" json:"amount"`
	Currency    string         `gorm:"type:varchar(10);default:CNY;comment:币种" json:"currency"`
	Channel     string         `gorm:"type:varchar(20);comment:支付渠道 wechat/alipay" json:"channel"`
	Status      int8           `gorm:"type:tinyint;default:0;comment:0待支付 1已支付 2已关闭 3已退款" json:"status"`
	PaidAt      *time.Time     `gorm:"comment:支付时间" json:"paidAt"`
	RefundAt    *time.Time     `gorm:"comment:退款时间" json:"refundAt"`
	RefundAmt   int64          `gorm:"comment:退款金额(分)" json:"refundAmt"`
	OpenID      string         `gorm:"type:varchar(128);comment:用户OpenID" json:"openId"`
	NotifyURL   string         `gorm:"type:varchar(500);comment:回调地址" json:"notifyUrl"`
	Extra       string         `gorm:"type:text;comment:扩展信息" json:"extra"`
	PayInfo     string         `gorm:"type:text;comment:支付参数" json:"-"`
	RawNotify   string         `gorm:"type:text;comment:原始回调数据" json:"-"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PayOrder) TableName() string {
	return "pay_order"
}
