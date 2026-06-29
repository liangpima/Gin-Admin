package service

import (
	"fmt"
	"testing"
	"time"

	"go-admin/internal/module/payment/model"
)

type mockOrderRepo struct {
	orders   map[string]*model.PayOrder
	nextID   uint
	createFn func(order *model.PayOrder) error
	updateFn func(order *model.PayOrder) error
}

func newMockRepo() *mockOrderRepo {
	return &mockOrderRepo{
		orders: make(map[string]*model.PayOrder),
		nextID: 1,
	}
}

func (m *mockOrderRepo) Create(order *model.PayOrder) error {
	if m.createFn != nil {
		return m.createFn(order)
	}
	order.ID = m.nextID
	m.nextID++
	m.orders[order.OrderNo] = order
	return nil
}

func (m *mockOrderRepo) Update(order *model.PayOrder) error {
	if m.updateFn != nil {
		return m.updateFn(order)
	}
	m.orders[order.OrderNo] = order
	return nil
}

func (m *mockOrderRepo) FindByOrderNo(tenantID uint, orderNo string) (*model.PayOrder, error) {
	if o, ok := m.orders[orderNo]; ok {
		return o, nil
	}
	return nil, fmt.Errorf("record not found")
}

func (m *mockOrderRepo) FindByTradeNo(tenantID uint, tradeNo string) (*model.PayOrder, error) {
	for _, o := range m.orders {
		if o.TradeNo == tradeNo {
			return o, nil
		}
	}
	return nil, fmt.Errorf("record not found")
}

func (m *mockOrderRepo) FindByID(tenantID, id uint) (*model.PayOrder, error) {
	for _, o := range m.orders {
		if o.ID == id {
			return o, nil
		}
	}
	return nil, fmt.Errorf("record not found")
}

func (m *mockOrderRepo) FindByOrderNoForNotify(orderNo string) (*model.PayOrder, error) {
	if o, ok := m.orders[orderNo]; ok {
		return o, nil
	}
	return nil, fmt.Errorf("record not found")
}

func (m *mockOrderRepo) FindList(tenantID uint, subject string, status int8, channel string, page, pageSize int) ([]model.PayOrder, int64, error) {
	var result []model.PayOrder
	for _, o := range m.orders {
		if subject != "" && o.Subject != subject {
			continue
		}
		if status >= 0 && o.Status != status {
			continue
		}
		if channel != "" && o.Channel != channel {
			continue
		}
		result = append(result, *o)
	}
	return result, int64(len(result)), nil
}

func newTestService(repo *mockOrderRepo) *PaymentService {
	return &PaymentService{orderRepo: repo}
}

const testTenantID uint = 1

func TestCreateOrder(t *testing.T) {
	t.Run("正常创建订单", func(t *testing.T) {
		repo := newMockRepo()
		svc := newTestService(repo)

		order, err := svc.CreateOrder(testTenantID, "ORDER001", "测试商品", "描述", 100, "wechat", "", "https://notify.url", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if order.ID == 0 {
			t.Fatal("expected order ID to be set")
		}
		if order.OrderNo != "ORDER001" {
			t.Errorf("expected OrderNo ORDER001, got %s", order.OrderNo)
		}
		if order.Amount != 100 {
			t.Errorf("expected Amount 100, got %d", order.Amount)
		}
		if order.Status != 0 {
			t.Errorf("expected Status 0, got %d", order.Status)
		}
		if order.Currency != "CNY" {
			t.Errorf("expected Currency CNY, got %s", order.Currency)
		}
	})

	t.Run("重复订单号返回已存在订单", func(t *testing.T) {
		repo := newMockRepo()
		svc := newTestService(repo)

		first, _ := svc.CreateOrder(testTenantID, "ORDER001", "商品1", "", 100, "wechat", "", "", "")
		second, err := svc.CreateOrder(testTenantID, "ORDER001", "商品2", "", 200, "alipay", "", "", "")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if first.ID != second.ID {
			t.Error("expected same order returned for duplicate orderNo")
		}
	})

	t.Run("创建失败时返回错误", func(t *testing.T) {
		repo := newMockRepo()
		repo.createFn = func(order *model.PayOrder) error {
			return fmt.Errorf("db error")
		}
		svc := newTestService(repo)

		_, err := svc.CreateOrder(testTenantID, "ORDER001", "商品", "", 100, "wechat", "", "", "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestCloseOrder(t *testing.T) {
	t.Run("关闭待支付订单", func(t *testing.T) {
		repo := newMockRepo()
		svc := newTestService(repo)

		svc.CreateOrder(testTenantID, "ORDER001", "商品", "", 100, "wechat", "", "", "")
		err := svc.CloseOrder(testTenantID, "ORDER001")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		order, _ := svc.GetOrder(testTenantID, "ORDER001")
		if order.Status != 2 {
			t.Errorf("expected status 2, got %d", order.Status)
		}
	})

	t.Run("已支付订单不能关闭", func(t *testing.T) {
		repo := newMockRepo()
		svc := newTestService(repo)

		svc.CreateOrder(testTenantID, "ORDER001", "商品", "", 100, "wechat", "", "", "")
		repo.orders["ORDER001"].Status = 1

		err := svc.CloseOrder(testTenantID, "ORDER001")
		if err == nil {
			t.Fatal("expected error for paid order")
		}
	})

	t.Run("已关闭订单不能再关闭", func(t *testing.T) {
		repo := newMockRepo()
		svc := newTestService(repo)

		svc.CreateOrder(testTenantID, "ORDER001", "商品", "", 100, "wechat", "", "", "")
		repo.orders["ORDER001"].Status = 2

		err := svc.CloseOrder(testTenantID, "ORDER001")
		if err == nil {
			t.Fatal("expected error for already closed order")
		}
	})

	t.Run("不存在的订单返回错误", func(t *testing.T) {
		repo := newMockRepo()
		svc := newTestService(repo)

		err := svc.CloseOrder(testTenantID, "NOT_EXIST")
		if err == nil {
			t.Fatal("expected error for non-existent order")
		}
	})
}

func TestHandleNotify(t *testing.T) {
	t.Run("正常回调更新订单为已支付", func(t *testing.T) {
		repo := newMockRepo()
		svc := newTestService(repo)

		svc.CreateOrder(testTenantID, "ORDER001", "商品", "", 100, "wechat", "", "", "")
		paidAt := time.Now()
		err := svc.HandleNotify("wechat", &PayNotifyResult{
			OrderNo: "ORDER001",
			TradeNo: "WX_TRADE_001",
			Status:  "success",
			PaidAt:  &paidAt,
			RawData: `{"raw":"data"}`,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		order, _ := svc.GetOrder(testTenantID, "ORDER001")
		if order.Status != 1 {
			t.Errorf("expected status 1, got %d", order.Status)
		}
		if order.TradeNo != "WX_TRADE_001" {
			t.Errorf("expected TradeNo WX_TRADE_001, got %s", order.TradeNo)
		}
	})

	t.Run("已支付订单重复回调不报错", func(t *testing.T) {
		repo := newMockRepo()
		svc := newTestService(repo)

		svc.CreateOrder(testTenantID, "ORDER001", "商品", "", 100, "wechat", "", "", "")
		repo.orders["ORDER001"].Status = 1

		err := svc.HandleNotify("wechat", &PayNotifyResult{
			OrderNo: "ORDER001",
			Status:  "success",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("nil回调返回错误", func(t *testing.T) {
		repo := newMockRepo()
		svc := newTestService(repo)

		err := svc.HandleNotify("wechat", nil)
		if err == nil {
			t.Fatal("expected error for nil result")
		}
	})

	t.Run("空订单号返回错误", func(t *testing.T) {
		repo := newMockRepo()
		svc := newTestService(repo)

		err := svc.HandleNotify("wechat", &PayNotifyResult{OrderNo: ""})
		if err == nil {
			t.Fatal("expected error for empty orderNo")
		}
	})

	t.Run("不存在的订单返回错误", func(t *testing.T) {
		repo := newMockRepo()
		svc := newTestService(repo)

		err := svc.HandleNotify("wechat", &PayNotifyResult{
			OrderNo: "NOT_EXIST",
			Status:  "success",
		})
		if err == nil {
			t.Fatal("expected error for non-existent order")
		}
	})

	t.Run("非success状态不更新订单", func(t *testing.T) {
		repo := newMockRepo()
		svc := newTestService(repo)

		svc.CreateOrder(testTenantID, "ORDER001", "商品", "", 100, "wechat", "", "", "")
		err := svc.HandleNotify("wechat", &PayNotifyResult{
			OrderNo: "ORDER001",
			Status:  "pending",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		order, _ := svc.GetOrder(testTenantID, "ORDER001")
		if order.Status != 0 {
			t.Errorf("expected status unchanged (0), got %d", order.Status)
		}
	})

	t.Run("金额不匹配拒绝回调", func(t *testing.T) {
		repo := newMockRepo()
		svc := newTestService(repo)

		svc.CreateOrder(testTenantID, "ORDER001", "商品", "", 100, "wechat", "", "", "")
		paidAt := time.Now()
		err := svc.HandleNotify("wechat", &PayNotifyResult{
			OrderNo: "ORDER001",
			Status:  "success",
			Amount:  999,
			PaidAt:  &paidAt,
		})
		if err == nil {
			t.Fatal("expected error for amount mismatch")
		}
	})
}

func TestRefundOrder(t *testing.T) {
	t.Run("正常退款", func(t *testing.T) {
		repo := newMockRepo()
		svc := newTestService(repo)

		svc.CreateOrder(testTenantID, "ORDER001", "商品", "", 1000, "wechat", "", "", "")
		repo.orders["ORDER001"].Status = 1

		err := svc.RefundOrder(testTenantID, "ORDER001", 500)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		order, _ := svc.GetOrder(testTenantID, "ORDER001")
		if order.Status != 3 {
			t.Errorf("expected status 3, got %d", order.Status)
		}
		if order.RefundAmt != 500 {
			t.Errorf("expected RefundAmt 500, got %d", order.RefundAmt)
		}
		if order.RefundAt == nil {
			t.Error("expected RefundAt to be set")
		}
	})

	t.Run("全额退款", func(t *testing.T) {
		repo := newMockRepo()
		svc := newTestService(repo)

		svc.CreateOrder(testTenantID, "ORDER001", "商品", "", 1000, "wechat", "", "", "")
		repo.orders["ORDER001"].Status = 1

		err := svc.RefundOrder(testTenantID, "ORDER001", 1000)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		order, _ := svc.GetOrder(testTenantID, "ORDER001")
		if order.RefundAmt != 1000 {
			t.Errorf("expected RefundAmt 1000, got %d", order.RefundAmt)
		}
	})

	t.Run("未支付订单不能退款", func(t *testing.T) {
		repo := newMockRepo()
		svc := newTestService(repo)

		svc.CreateOrder(testTenantID, "ORDER001", "商品", "", 1000, "wechat", "", "", "")

		err := svc.RefundOrder(testTenantID, "ORDER001", 500)
		if err == nil {
			t.Fatal("expected error for unpaid order")
		}
	})

	t.Run("退款金额为0返回错误", func(t *testing.T) {
		repo := newMockRepo()
		svc := newTestService(repo)

		svc.CreateOrder(testTenantID, "ORDER001", "商品", "", 1000, "wechat", "", "", "")
		repo.orders["ORDER001"].Status = 1

		err := svc.RefundOrder(testTenantID, "ORDER001", 0)
		if err == nil {
			t.Fatal("expected error for zero refund amount")
		}
	})

	t.Run("退款金额为负返回错误", func(t *testing.T) {
		repo := newMockRepo()
		svc := newTestService(repo)

		svc.CreateOrder(testTenantID, "ORDER001", "商品", "", 1000, "wechat", "", "", "")
		repo.orders["ORDER001"].Status = 1

		err := svc.RefundOrder(testTenantID, "ORDER001", -100)
		if err == nil {
			t.Fatal("expected error for negative refund amount")
		}
	})

	t.Run("退款金额超过订单金额返回错误", func(t *testing.T) {
		repo := newMockRepo()
		svc := newTestService(repo)

		svc.CreateOrder(testTenantID, "ORDER001", "商品", "", 1000, "wechat", "", "", "")
		repo.orders["ORDER001"].Status = 1

		err := svc.RefundOrder(testTenantID, "ORDER001", 2000)
		if err == nil {
			t.Fatal("expected error for refund amount exceeding order amount")
		}
	})

	t.Run("不存在的订单返回错误", func(t *testing.T) {
		repo := newMockRepo()
		svc := newTestService(repo)

		err := svc.RefundOrder(testTenantID, "NOT_EXIST", 100)
		if err == nil {
			t.Fatal("expected error for non-existent order")
		}
	})
}

func TestGetOrder(t *testing.T) {
	t.Run("正常获取订单", func(t *testing.T) {
		repo := newMockRepo()
		svc := newTestService(repo)

		svc.CreateOrder(testTenantID, "ORDER001", "商品", "", 100, "wechat", "", "", "")
		order, err := svc.GetOrder(testTenantID, "ORDER001")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if order.OrderNo != "ORDER001" {
			t.Errorf("expected OrderNo ORDER001, got %s", order.OrderNo)
		}
	})

	t.Run("不存在的订单返回错误", func(t *testing.T) {
		repo := newMockRepo()
		svc := newTestService(repo)

		_, err := svc.GetOrder(testTenantID, "NOT_EXIST")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestGetOrderByID(t *testing.T) {
	t.Run("正常获取订单", func(t *testing.T) {
		repo := newMockRepo()
		svc := newTestService(repo)

		svc.CreateOrder(testTenantID, "ORDER001", "商品", "", 100, "wechat", "", "", "")
		order, err := svc.GetOrderByID(testTenantID, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if order.ID != 1 {
			t.Errorf("expected ID 1, got %d", order.ID)
		}
	})

	t.Run("不存在的ID返回错误", func(t *testing.T) {
		repo := newMockRepo()
		svc := newTestService(repo)

		_, err := svc.GetOrderByID(testTenantID, 999)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestFindList(t *testing.T) {
	repo := newMockRepo()
	svc := newTestService(repo)

	svc.CreateOrder(testTenantID, "O1", "商品A", "", 100, "wechat", "", "", "")
	svc.CreateOrder(testTenantID, "O2", "商品B", "", 200, "alipay", "", "", "")
	svc.CreateOrder(testTenantID, "O3", "商品A", "", 300, "wechat", "", "", "")
	repo.orders["O2"].Status = 1

	t.Run("按标题搜索", func(t *testing.T) {
		list, total, err := svc.FindList(testTenantID, "商品A", -1, "", 1, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 2 {
			t.Errorf("expected total 2, got %d", total)
		}
		if len(list) != 2 {
			t.Errorf("expected 2 items, got %d", len(list))
		}
	})

	t.Run("按状态筛选", func(t *testing.T) {
		list, total, err := svc.FindList(testTenantID, "", 1, "", 1, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
		if len(list) != 1 {
			t.Errorf("expected 1 item, got %d", len(list))
		}
	})

	t.Run("按渠道筛选", func(t *testing.T) {
		list, total, err := svc.FindList(testTenantID, "", -1, "alipay", 1, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
		if len(list) != 1 {
			t.Errorf("expected 1 item, got %d", len(list))
		}
	})

	t.Run("无筛选条件返回全部", func(t *testing.T) {
		_, total, err := svc.FindList(testTenantID, "", -1, "", 1, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 3 {
			t.Errorf("expected total 3, got %d", total)
		}
	})
}

func TestConcurrentCreateOrder(t *testing.T) {
	repo := newMockRepo()
	svc := newTestService(repo)

	const goroutines = 50
	errs := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			_, err := svc.CreateOrder(testTenantID, "SAME_ORDER", "商品", "", 100, "wechat", "", "", "")
			errs <- err
		}()
	}

	for i := 0; i < goroutines; i++ {
		<-errs
	}

	if len(repo.orders) != 1 {
		t.Errorf("expected 1 order (dedup), got %d", len(repo.orders))
	}

	order, _ := svc.GetOrder(testTenantID, "SAME_ORDER")
	if order == nil {
		t.Fatal("order not found")
	}
}
