package model

import (
	"go-admin/internal/common"
)

// SysPost 岗位。
//
// 岗位是**租户内**的基础数据，不是全局表 —— 每个租户各自维护自己的岗位，
// 因此继承 TenantBaseModel，所有查询必须带 tenant_id 过滤。
//
// 唯一性约束是 (tenant_id, code) 复合唯一：不同租户可以存在相同编码的岗位。
//
// 这里没有给 Code 打 `uniqueIndex` 标签：GORM 无法把索引字段指到嵌入结构体里的
// TenantID（定义在 common.TenantBaseModel），若只给 Code 打标签，会建出一个
// **全局**唯一索引，反而与租户内唯一的语义冲突。真正的复合唯一索引由
// sql/init.sql 与 sql/migrations/2026-09-16-post-tenant.sql 的 DDL 建立
// （sys_post 不参与 AutoMigrate，模型标签仅作文档）。
type SysPost struct {
	common.TenantBaseModel
	Code   string `gorm:"type:varchar(64);comment:岗位编码" json:"code"`
	Name   string `gorm:"type:varchar(64);comment:岗位名称" json:"name"`
	Sort   int    `gorm:"type:int;default:0;comment:排序" json:"sort"`
	Status int8   `gorm:"type:tinyint;comment:状态" json:"status"`
}

func (SysPost) TableName() string {
	return "sys_post"
}
