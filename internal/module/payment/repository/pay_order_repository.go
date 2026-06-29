package repository

import (
	"go-admin/internal/common"
	"go-admin/internal/database"
	"go-admin/internal/module/payment/model"
)

type PayOrderRepository interface {
	Create(order *model.PayOrder) error
	Update(order *model.PayOrder) error
	FindByOrderNo(tenantID uint, orderNo string) (*model.PayOrder, error)
	FindByTradeNo(tenantID uint, tradeNo string) (*model.PayOrder, error)
	FindByID(tenantID, id uint) (*model.PayOrder, error)
	FindList(tenantID uint, subject string, status int8, channel string, page, pageSize int) ([]model.PayOrder, int64, error)
	// 用于支付回调，不带 tenant 过滤（回调无法获取 tenant_id）
	FindByOrderNoForNotify(orderNo string) (*model.PayOrder, error)
}

type payOrderRepository struct{}

func NewPayOrderRepository() PayOrderRepository {
	return &payOrderRepository{}
}

func (r *payOrderRepository) Create(order *model.PayOrder) error {
	return database.DB.Create(order).Error
}

func (r *payOrderRepository) Update(order *model.PayOrder) error {
	return database.DB.Save(order).Error
}

func (r *payOrderRepository) FindByOrderNo(tenantID uint, orderNo string) (*model.PayOrder, error) {
	var order model.PayOrder
	err := common.TenantScope(database.DB, tenantID).Where("order_no = ?", orderNo).First(&order).Error
	return &order, err
}

func (r *payOrderRepository) FindByTradeNo(tenantID uint, tradeNo string) (*model.PayOrder, error) {
	var order model.PayOrder
	err := common.TenantScope(database.DB, tenantID).Where("trade_no = ?", tradeNo).First(&order).Error
	return &order, err
}

func (r *payOrderRepository) FindByID(tenantID, id uint) (*model.PayOrder, error) {
	var order model.PayOrder
	err := common.TenantScope(database.DB, tenantID).First(&order, id).Error
	return &order, err
}

func (r *payOrderRepository) FindByOrderNoForNotify(orderNo string) (*model.PayOrder, error) {
	var order model.PayOrder
	err := database.DB.Where("order_no = ?", orderNo).First(&order).Error
	return &order, err
}

func (r *payOrderRepository) FindList(tenantID uint, subject string, status int8, channel string, page, pageSize int) ([]model.PayOrder, int64, error) {
	var orders []model.PayOrder
	var total int64

	query := common.TenantScope(database.DB.Model(&model.PayOrder{}), tenantID)

	if subject != "" {
		query = query.Where("subject LIKE ?", "%"+subject+"%")
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	if channel != "" {
		query = query.Where("channel = ?", channel)
	}

	query.Count(&total)
	err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}
