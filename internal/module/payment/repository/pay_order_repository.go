package repository

import (
	"time"

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
	// MarkPaidIfPending 原子地将"待支付"订单标记为"已支付"，
	// 返回 affected=true 表示由本次调用完成状态流转（即首次处理）。
	// 重复/并发回调只会有一个成功，从而实现幂等。
	MarkPaidIfPending(orderNo, tradeNo string, paidAt *time.Time, rawNotify string) (bool, error)

	// ClaimRefund 原子地把"已支付"订单抢占为"退款中"。
	// 返回 affected=true 表示本次调用抢到了退款权，**只有抢到的请求才允许调支付网关**。
	// 这是防止并发重复退款的关键：网关调用是不可逆的资金操作，必须先抢占再调用。
	ClaimRefund(orderNo string) (bool, error)

	// MarkRefunded 把"退款中"订单落定为"已退款"
	MarkRefunded(orderNo string, refundAmt int64, refundAt time.Time) (bool, error)

	// ReleaseRefundClaim 退款失败时把"退款中"回滚为"已支付"，让用户可以重试
	ReleaseRefundClaim(orderNo string) error
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

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}

// MarkPaidIfPending 以「条件更新 + 影响行数」实现回调幂等：
// 只有仍处于待支付（status=0）的订单才会被更新，重复回调不会重复发货。
func (r *payOrderRepository) MarkPaidIfPending(orderNo, tradeNo string, paidAt *time.Time, rawNotify string) (bool, error) {
	result := database.DB.Model(&model.PayOrder{}).
		Where("order_no = ? AND status = ?", orderNo, model.StatusPending).
		Updates(map[string]interface{}{
			"status":     1,
			"trade_no":   tradeNo,
			"paid_at":    paidAt,
			"raw_notify": rawNotify,
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// ClaimRefund 以「条件更新 + 影响行数」抢占退款权：只有仍处于已支付（status=1）
// 的订单才能被置为退款中（status=4）。
//
// 为什么必须这样做：退款要调用支付网关，那是**不可逆的资金操作**。
// 若先查状态、再调网关、最后改状态，两个并发请求会同时通过状态检查，
// 于是向渠道真实退款两次 —— 数据库状态最终一致，钱却多退了一份。
// 把状态流转提前到网关调用之前，就只有一个请求能拿到退款权。
func (r *payOrderRepository) ClaimRefund(orderNo string) (bool, error) {
	result := database.DB.Model(&model.PayOrder{}).
		Where("order_no = ? AND status = ?", orderNo, model.StatusPaid).
		Update("status", model.StatusRefunding)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// MarkRefunded 把退款中订单落定为已退款
func (r *payOrderRepository) MarkRefunded(orderNo string, refundAmt int64, refundAt time.Time) (bool, error) {
	result := database.DB.Model(&model.PayOrder{}).
		Where("order_no = ? AND status = ?", orderNo, model.StatusRefunding).
		Updates(map[string]interface{}{
			"status":     model.StatusRefunded,
			"refund_amt": refundAmt,
			"refund_at":  refundAt,
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// ReleaseRefundClaim 网关退款失败时回滚状态，让订单回到可再次退款的状态
func (r *payOrderRepository) ReleaseRefundClaim(orderNo string) error {
	return database.DB.Model(&model.PayOrder{}).
		Where("order_no = ? AND status = ?", orderNo, model.StatusRefunding).
		Update("status", model.StatusPaid).Error
}
