package service

import (
	"fmt"
	"log"
	"sync"
	"time"

	"go-admin/internal/common"
	"go-admin/internal/module/payment/model"
	paymentRepo "go-admin/internal/module/payment/repository"
	systemModel "go-admin/internal/module/system/model"
	systemService "go-admin/internal/module/system/service"
)

type PayNotifyResult struct {
	OrderNo  string
	TradeNo  string
	Status   string
	Amount   int64
	PaidAt   *time.Time
	RawData  string
}

type PaymentService struct {
	orderRepo paymentRepo.PayOrderRepository
	mu        sync.Mutex

	// gatewayRefund 实际调用支付渠道退款的函数，默认为 refundVia。
	// 留出这个接缝是因为真实渠道调用依赖线上配置与网络，无法在单测中执行，
	// 而「并发退款只允许一个请求打到渠道」正是最需要回归保护的行为。
	gatewayRefund func(order *model.PayOrder, refundNo string, refundAmt int64) error
}

func NewPaymentService() *PaymentService {
	return &PaymentService{
		orderRepo: paymentRepo.NewPayOrderRepository(),
	}
}

// refund 调用支付渠道发起退款，可通过 gatewayRefund 替换以便测试
func (s *PaymentService) refund(order *model.PayOrder, refundNo string, refundAmt int64) error {
	if s.gatewayRefund != nil {
		return s.gatewayRefund(order, refundNo, refundAmt)
	}
	return s.refundVia(order, refundNo, refundAmt)
}

// refundVia 按渠道发起真实退款请求
func (s *PaymentService) refundVia(order *model.PayOrder, refundNo string, refundAmt int64) error {
	switch order.Channel {
	case "wechat":
		cfg := LoadWechatPayConfig()
		gw := NewWechatPayGateway(*cfg)
		return gw.Refund(nil, order.OrderNo, refundNo, order.Amount, refundAmt)
	case "alipay":
		cfg := LoadAlipayConfig()
		gw := NewAlipayGateway(*cfg)
		_, err := gw.Refund(nil, order.OrderNo, refundNo, refundAmt)
		return err
	default:
		return common.NewBizErrorf("不支持的支付渠道: %s", order.Channel)
	}
}

func (s *PaymentService) CreateOrder(tenantID uint, orderNo, subject, body string, amount int64, channel, openID, notifyURL, extra string) (*model.PayOrder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 订单号重复时，只有「标题 + 金额 + 渠道」全部一致才视为幂等重放并复用原单；
	// 否则说明订单号被复用（撞号或调用方传了重复单号），
	// 此时**绝不能把已存在的订单返回给调用方** —— 那是别人的单子，
	// 调用方却会拿新参数去调渠道，造成订单与支付参数错配。
	if existing, err := s.orderRepo.FindByOrderNo(tenantID, orderNo); err == nil && existing != nil && existing.ID > 0 {
		if existing.Subject == subject && existing.Amount == amount && existing.Channel == channel {
			return existing, nil
		}
		return nil, common.NewBizError("订单号已存在，请更换订单号后重试")
	}

	order := &model.PayOrder{
		TenantID:  tenantID,
		OrderNo:   orderNo,
		Subject:   subject,
		Body:      body,
		Amount:    amount,
		Currency:  "CNY",
		Channel:   channel,
		Status:    model.StatusPending,
		OpenID:    openID,
		NotifyURL: notifyURL,
		Extra:     extra,
	}

	if err := s.orderRepo.Create(order); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *PaymentService) GetOrder(tenantID uint, orderNo string) (*model.PayOrder, error) {
	order, err := s.orderRepo.FindByOrderNo(tenantID, orderNo)
	if err != nil {
		return nil, common.NotFoundOrErr(err, "订单不存在")
	}
	return order, nil
}

func (s *PaymentService) GetOrderByID(tenantID, id uint) (*model.PayOrder, error) {
	order, err := s.orderRepo.FindByID(tenantID, id)
	if err != nil {
		return nil, common.NotFoundOrErr(err, "订单不存在")
	}
	return order, nil
}

func (s *PaymentService) CloseOrder(tenantID uint, orderNo string) error {
	// 走 GetOrder 而非直接查 repository：前者会把「记录不存在」转成 404 业务错误，
	// 否则 gorm.ErrRecordNotFound 一路透出会变成 500「服务器内部错误」
	order, err := s.GetOrder(tenantID, orderNo)
	if err != nil {
		return err
	}

	if order.Status == model.StatusPaid {
		return common.NewBizError("订单已支付，无法关闭")
	}
	if order.Status == model.StatusClosed {
		return common.NewBizError("订单已关闭")
	}

	order.Status = model.StatusClosed
	return s.orderRepo.Update(order)
}

func (s *PaymentService) HandleNotify(channel string, result *PayNotifyResult) error {
	if result == nil || result.OrderNo == "" {
		return fmt.Errorf("invalid notify result")
	}

	// 回调场景：不带 tenant_id 过滤（无法从外部请求获取租户信息）
	order, err := s.orderRepo.FindByOrderNoForNotify(result.OrderNo)
	if err != nil {
		return fmt.Errorf("order not found: %s", result.OrderNo)
	}

	if order.Status == model.StatusPaid {
		log.Printf("[payment] order %s already paid, skip", result.OrderNo)
		return nil
	}

	if result.Status == "success" {
		if result.Amount > 0 && result.Amount != order.Amount {
			log.Printf("[payment] order %s amount mismatch: expected %d, got %d",
				result.OrderNo, order.Amount, result.Amount)
			return fmt.Errorf("支付金额不匹配")
		}

		paidAt := result.PaidAt
		if paidAt == nil {
			now := time.Now()
			paidAt = &now
		}

		// 以数据库条件更新保证幂等：并发/重复回调中只有一个能完成状态流转，
		// 避免因实例级锁不共享（每次请求新建 Service）导致的重复发货。
		affected, err := s.orderRepo.MarkPaidIfPending(result.OrderNo, result.TradeNo, paidAt, result.RawData)
		if err != nil {
			return err
		}
		if !affected {
			log.Printf("[payment] order %s already processed by concurrent notify, skip", result.OrderNo)
			return nil
		}

		log.Printf("[payment] order %s paid successfully, trade_no: %s", result.OrderNo, result.TradeNo)
	}

	return nil
}

// validateRefund 退款前的纯校验（不产生任何副作用），供退款流程复用
func validateRefund(order *model.PayOrder, refundAmt int64) error {
	if order.Status == model.StatusRefunding {
		return common.NewBizError("退款正在处理中，请勿重复提交")
	}
	if order.Status == model.StatusRefunded {
		return common.NewBizError("订单已退款")
	}
	if order.Status != model.StatusPaid {
		return common.NewBizError("订单未支付，无法退款")
	}
	if refundAmt <= 0 {
		return common.NewBizError("退款金额必须大于0")
	}
	if refundAmt > order.Amount {
		return common.NewBizError("退款金额不能超过订单金额")
	}
	return nil
}

func (s *PaymentService) FindList(tenantID uint, subject string, status int8, channel string, page, pageSize int) ([]model.PayOrder, int64, error) {
	return s.orderRepo.FindList(tenantID, subject, status, channel, page, pageSize)
}

type CreateOrderResult struct {
	Order    *model.PayOrder
	PayInfo  map[string]interface{}
	PayError error
}

func (s *PaymentService) CreateOrderWithPayInfo(tenantID uint, orderNo, subject, body string, amount int64, channel, openID, extra string) (*CreateOrderResult, error) {
	configs := LoadWechatPayConfig()
	notifyURL := configs.NotifyURL

	order, err := s.CreateOrder(tenantID, orderNo, subject, body, amount, channel, openID, notifyURL, extra)
	if err != nil {
		return nil, err
	}

	result := &CreateOrderResult{
		Order: order,
		PayInfo: map[string]interface{}{
			"orderNo": order.OrderNo,
			"amount":  order.Amount,
			"status":  order.Status,
		},
	}

	switch channel {
	case "wechat":
		cfg := LoadWechatPayConfig()
		gw := NewWechatPayGateway(*cfg)
		payInfo, payErr := gw.Prepay(nil, orderNo, subject, body, amount, openID)
		if payErr != nil {
			result.PayError = payErr
		} else {
			for k, v := range payInfo {
				result.PayInfo[k] = v
			}
		}
	case "alipay":
		cfg := LoadAlipayConfig()
		gw := NewAlipayGateway(*cfg)
		payInfo, payErr := gw.Prepay(nil, orderNo, subject, amount, cfg.ReturnURL)
		if payErr != nil {
			result.PayError = payErr
		} else {
			for k, v := range payInfo {
				result.PayInfo[k] = v
			}
		}
	}

	return result, nil
}

type RefundOrderResult struct {
	RefundNo string
	Status   string
	Error    error
}

// RefundOrderWithPayInfo 发起退款。
//
// 关键：**先抢占状态，再调用支付网关**。
//
// 早期实现是「查状态 → 调网关 → 改状态」，三者之间没有互斥。
// 两个并发请求会同时通过状态检查，于是向渠道真实退款两次 ——
// 数据库最终状态看起来正常，钱却多退了一份。
// 实例级 mutex 也救不了：它只覆盖最后那步改状态，而资金操作在锁外。
//
// 现在的顺序是：
//  1. 校验（纯读，无副作用）
//  2. ClaimRefund 用条件更新把 已支付 → 退款中，只有一个请求能成功
//  3. 抢到的人才调网关
//  4. 网关失败 → 回滚为已支付（允许重试）；成功 → 落定为已退款
func (s *PaymentService) RefundOrderWithPayInfo(tenantID uint, orderNo string, refundNo string, refundAmt int64) (*RefundOrderResult, error) {
	order, err := s.GetOrder(tenantID, orderNo)
	if err != nil {
		return nil, err
	}

	if err := validateRefund(order, refundAmt); err != nil {
		return nil, err
	}

	// 渠道校验放在抢占之前：不支持时直接拒绝，不会留下卡在「退款中」的订单
	if order.Channel != "wechat" && order.Channel != "alipay" {
		return nil, common.NewBizErrorf("不支持的支付渠道: %s", order.Channel)
	}

	// 抢占退款权。抢不到说明已有并发请求在处理，直接拒绝，绝不重复调网关。
	claimed, err := s.orderRepo.ClaimRefund(orderNo)
	if err != nil {
		return nil, err
	}
	if !claimed {
		return nil, common.NewBizError("退款正在处理中或已退款，请勿重复提交")
	}

	result := &RefundOrderResult{
		RefundNo: refundNo,
		Status:   "refunding",
	}

	if err := s.refund(order, refundNo, refundAmt); err != nil {
		// 渠道侧失败：回滚状态，让用户可以重新发起退款
		if releaseErr := s.orderRepo.ReleaseRefundClaim(orderNo); releaseErr != nil {
			log.Printf("[payment] 退款失败后回滚状态也失败, order=%s: %v", orderNo, releaseErr)
		}
		result.Error = err
		return result, nil
	}

	affected, err := s.orderRepo.MarkRefunded(orderNo, refundAmt, time.Now())
	if err != nil {
		result.Error = fmt.Errorf("更新退款状态失败: %v", err)
		return result, nil
	}
	if !affected {
		// 抢到了退款权却更新失败，属异常状态，记日志便于排查
		log.Printf("[payment] 退款已完成但状态落定失败, order=%s", orderNo)
	}

	return result, nil
}

func loadPayConfig() map[string]string {
	configService := systemService.NewConfigService()
	// 需要真实密钥用于签名与验签，因此读取原始值（接口侧会打码）
	results, _ := configService.FindByPrefixRaw("pay.")

	cfgMap := make(map[string]string)
	for _, r := range results {
		if cfg, ok := r.(systemModel.SysConfig); ok {
			key := cfg.ConfigKey
			if len(key) > 4 && key[:4] == "pay." {
				cfgMap[key[4:]] = cfg.Value
			}
		}
	}
	return cfgMap
}

func LoadWechatPayConfig() *WechatPayConfig {
	cfgMap := loadPayConfig()
	return &WechatPayConfig{
		AppID:     cfgMap["wechat_app_id"],
		MchID:     cfgMap["wechat_mch_id"],
		Key:       cfgMap["wechat_key"],
		APIv3Key:  cfgMap["wechat_apiv3_key"],
		SerialNo:  cfgMap["wechat_serial_no"],
		NotifyURL: cfgMap["notify_url"],
	}
}

func LoadAlipayConfig() *AlipayConfig {
	cfgMap := loadPayConfig()
	return &AlipayConfig{
		AppID:       cfgMap["alipay_app_id"],
		PrivateKey:  cfgMap["alipay_key"],
		NotifyURL:   cfgMap["notify_url"],
		ReturnURL:   cfgMap["return_url"],
		PublicKeyID: cfgMap["alipay_public_key"],
	}
}
