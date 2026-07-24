package service

import (
	"fmt"
	"log"
	"sync"
	"time"

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
}

func NewPaymentService() *PaymentService {
	return &PaymentService{
		orderRepo: paymentRepo.NewPayOrderRepository(),
	}
}

func (s *PaymentService) CreateOrder(tenantID uint, orderNo, subject, body string, amount int64, channel, openID, notifyURL, extra string) (*model.PayOrder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, _ := s.orderRepo.FindByOrderNo(tenantID, orderNo)
	if existing != nil && existing.ID > 0 {
		return existing, nil
	}

	order := &model.PayOrder{
		TenantID:  tenantID,
		OrderNo:   orderNo,
		Subject:   subject,
		Body:      body,
		Amount:    amount,
		Currency:  "CNY",
		Channel:   channel,
		Status:    0,
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
	return s.orderRepo.FindByOrderNo(tenantID, orderNo)
}

func (s *PaymentService) GetOrderByID(tenantID, id uint) (*model.PayOrder, error) {
	return s.orderRepo.FindByID(tenantID, id)
}

func (s *PaymentService) CloseOrder(tenantID uint, orderNo string) error {
	order, err := s.orderRepo.FindByOrderNo(tenantID, orderNo)
	if err != nil {
		return err
	}

	if order.Status == 1 {
		return fmt.Errorf("订单已支付，无法关闭")
	}
	if order.Status == 2 {
		return fmt.Errorf("订单已关闭")
	}

	order.Status = 2
	return s.orderRepo.Update(order)
}

func (s *PaymentService) HandleNotify(channel string, result *PayNotifyResult) error {
	if result == nil || result.OrderNo == "" {
		return fmt.Errorf("invalid notify result")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 回调场景：不带 tenant_id 过滤（无法从外部请求获取租户信息）
	order, err := s.orderRepo.FindByOrderNoForNotify(result.OrderNo)
	if err != nil {
		return fmt.Errorf("order not found: %s", result.OrderNo)
	}

	if order.Status == 1 {
		log.Printf("[payment] order %s already paid, skip", result.OrderNo)
		return nil
	}

	if result.Status == "success" {
		if result.Amount > 0 && result.Amount != order.Amount {
			log.Printf("[payment] order %s amount mismatch: expected %d, got %d",
				result.OrderNo, order.Amount, result.Amount)
			return fmt.Errorf("支付金额不匹配")
		}

		order.Status = 1
		order.TradeNo = result.TradeNo
		order.PaidAt = result.PaidAt
		order.RawNotify = result.RawData
		if err := s.orderRepo.Update(order); err != nil {
			return err
		}
		log.Printf("[payment] order %s paid successfully, trade_no: %s", result.OrderNo, result.TradeNo)
	}

	return nil
}

func (s *PaymentService) RefundOrder(tenantID uint, orderNo string, refundAmt int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, err := s.orderRepo.FindByOrderNo(tenantID, orderNo)
	if err != nil {
		return err
	}

	if order.Status != 1 {
		return fmt.Errorf("订单未支付，无法退款")
	}

	if refundAmt <= 0 {
		return fmt.Errorf("退款金额必须大于0")
	}

	if refundAmt > order.Amount {
		return fmt.Errorf("退款金额不能超过订单金额")
	}

	order.Status = 3
	now := time.Now()
	order.RefundAt = &now
	order.RefundAmt = refundAmt
	return s.orderRepo.Update(order)
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

func (s *PaymentService) RefundOrderWithPayInfo(tenantID uint, orderNo string, refundNo string, refundAmt int64) (*RefundOrderResult, error) {
	order, err := s.GetOrder(tenantID, orderNo)
	if err != nil {
		return nil, err
	}

	if order.Status != 1 {
		statusMap := map[int8]string{0: "待支付", 2: "已关闭", 3: "已退款"}
		return nil, fmt.Errorf("订单状态为%s，无法退款", statusMap[order.Status])
	}

	if refundAmt > order.Amount {
		return nil, fmt.Errorf("退款金额不能超过订单金额")
	}

	result := &RefundOrderResult{
		RefundNo: refundNo,
		Status:   "refunding",
	}

	switch order.Channel {
	case "wechat":
		cfg := LoadWechatPayConfig()
		gw := NewWechatPayGateway(*cfg)
		err = gw.Refund(nil, orderNo, refundNo, order.Amount, refundAmt)
	case "alipay":
		cfg := LoadAlipayConfig()
		gw := NewAlipayGateway(*cfg)
		_, err = gw.Refund(nil, orderNo, refundNo, refundAmt)
	}

	if err != nil {
		result.Error = err
		return result, nil
	}

	if err := s.RefundOrder(tenantID, orderNo, refundAmt); err != nil {
		result.Error = fmt.Errorf("更新退款状态失败: %v", err)
		return result, nil
	}

	return result, nil
}

func loadPayConfig() map[string]string {
	configService := systemService.NewConfigService()
	results, _ := configService.FindByPrefix("pay.")

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
