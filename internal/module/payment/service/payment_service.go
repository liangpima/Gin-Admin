package service

import (
	"fmt"
	"log"
	"sync"
	"time"

	"go-admin/internal/database"
	"go-admin/internal/module/payment/model"
	"go-admin/internal/module/payment/repository"
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
	orderRepo repository.PayOrderRepository
	mu        sync.Mutex
}

func NewPaymentService() *PaymentService {
	return &PaymentService{
		orderRepo: repository.NewPayOrderRepository(),
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

func loadPayConfig() map[string]string {
	var configs []struct {
		Key   string
		Value string
	}
	database.DB.Raw("SELECT `key`, `value` FROM sys_config WHERE `key` LIKE 'pay.%'").Scan(&configs)

	cfgMap := make(map[string]string)
	for _, c := range configs {
		key := c.Key
		if len(key) > 4 && key[:4] == "pay." {
			cfgMap[key[4:]] = c.Value
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
