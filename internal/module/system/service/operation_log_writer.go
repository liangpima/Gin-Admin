package service

import (
	"go-admin/internal/middleware"
	"go-admin/internal/module/system/model"
	"go-admin/internal/module/system/repository"
)

// OperationLogWriter 实现 middleware.OperationLogWriter，把审计条目落库。
//
// 适配器放在 system 模块的原因与 rbac_resolver.go 相同：
// 中间件只声明「我需要写一条审计记录」这一能力，
// 具体写到哪张表、字段怎么映射属于系统模块的知识。
// 这样 middleware 不再 import system 的 model 与 repository。
//
// 依赖 middleware 仅用于编译期断言接口实现。
type OperationLogWriter struct{}

var _ middleware.OperationLogWriter = (*OperationLogWriter)(nil)

func NewOperationLogWriter() *OperationLogWriter {
	return &OperationLogWriter{}
}

// WriteOperationLog 把中间件采集的审计条目转换为落库模型。
//
// 这里是 middleware.OperationLogEntry 与 model.SysOperationLog 之间
// **唯一**的映射点：将来审计表增减字段，只需要改这一处。
func (w *OperationLogWriter) WriteOperationLog(entry *middleware.OperationLogEntry) error {
	if entry == nil {
		return nil
	}

	log := &model.SysOperationLog{
		TenantID:      entry.TenantID,
		Title:         entry.Title,
		Action:        entry.Action,
		RequestMethod: entry.RequestMethod,
		RequestURL:    entry.RequestURL,
		RequestParam:  entry.RequestParam,
		Status:        entry.Status,
		IP:            entry.IP,
		UserAgent:     entry.UserAgent,
		OperatorID:    entry.OperatorID,
		OperatorName:  entry.OperatorName,
		CostTime:      entry.CostTime,
		ErrorMsg:      entry.ErrorMsg,
	}

	return repository.NewLogRepository().CreateOperationLog(log)
}
