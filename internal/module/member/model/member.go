package model

import (
	"time"

	"go-admin/internal/common"

	"gorm.io/gorm"
)

type Member struct {
	common.TenantBaseModel
	MemberNo      string        `gorm:"type:varchar(32);uniqueIndex;comment:会员编号" json:"memberNo"`
	Username      string        `gorm:"type:varchar(64);comment:用户名" json:"username"`
	Nickname      string        `gorm:"type:varchar(64);comment:昵称" json:"nickname"`
	Avatar        string        `gorm:"type:varchar(512);comment:头像" json:"avatar"`
	Phone         string        `gorm:"type:varchar(20);uniqueIndex;comment:手机号" json:"phone"`
	Gender        int8          `gorm:"type:tinyint;default:0;comment:性别 0未知 1男 2女" json:"gender"`
	Birthday      *time.Time    `gorm:"comment:出生日期" json:"birthday"`
	LevelID       uint          `gorm:"comment:等级ID" json:"levelId"`
	Status        int8          `gorm:"type:tinyint;default:1;comment:状态 0停用 1正常" json:"status"`
	Points        int64         `gorm:"comment:积分" json:"points"`
	WechatOpenid  string        `gorm:"type:varchar(128);index;comment:微信小程序openid" json:"wechatOpenid"`
	RegisterTime  time.Time     `gorm:"comment:注册时间" json:"registerTime"`
	LastVisitTime *time.Time    `gorm:"comment:最后访问时间" json:"lastVisitTime"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Member) TableName() string {
	return "pay_member"
}

type MemberLevel struct {
	common.TenantBaseModel
	Name      string  `gorm:"type:varchar(64);not null;comment:等级名称" json:"name"`
	MinPoints int64   `gorm:"comment:最低积分" json:"minPoints"`
	Discount  float64 `gorm:"type:decimal(3,1);default:10.0;comment:折扣 10表示不打折 8表示八折" json:"discount"`
	Icon      string  `gorm:"type:varchar(256);comment:等级图标" json:"icon"`
	Sort      int     `gorm:"comment:排序" json:"sort"`
	Status    int8    `gorm:"type:tinyint;default:1;comment:状态" json:"status"`
}

func (MemberLevel) TableName() string {
	return "pay_member_level"
}

type MemberTag struct {
	common.TenantBaseModel
	Name   string `gorm:"type:varchar(64);not null;comment:标签名称" json:"name"`
	Color  string `gorm:"type:varchar(20);default:#409eff;comment:标签颜色" json:"color"`
	Sort   int    `gorm:"comment:排序" json:"sort"`
	Status int8   `gorm:"type:tinyint;default:1;comment:状态" json:"status"`
}

func (MemberTag) TableName() string {
	return "pay_member_tag"
}

type MemberTagRel struct {
	MemberID uint `gorm:"primaryKey;comment:会员ID" json:"memberId"`
	TagID    uint `gorm:"primaryKey;comment:标签ID" json:"tagId"`
}

func (MemberTagRel) TableName() string {
	return "pay_member_tag_rel"
}

type PointsLog struct {
	common.TenantBaseModel
	MemberID uint   `gorm:"comment:会员ID" json:"memberId"`
	Change   int64  `gorm:"column:points_change;comment:变更积分" json:"change"`
	Type     int8   `gorm:"type:tinyint;default:1;comment:类型 1获取 2消费" json:"type"`
	Source   string `gorm:"type:varchar(64);comment:来源" json:"source"`
	OrderNo  string `gorm:"type:varchar(64);comment:关联订单号" json:"orderNo"`
}

func (PointsLog) TableName() string {
	return "pay_points_log"
}
