package controller

import (
	"fmt"
	"io"
	"log"
	"time"

	"go-admin/internal/common"
	"go-admin/internal/module/payment/service"

	"github.com/gin-gonic/gin"
)

type PaymentController struct {
	paymentService *service.PaymentService
}

func NewPaymentController() *PaymentController {
	return &PaymentController{paymentService: service.NewPaymentService()}
}

func (ctl *PaymentController) CreateOrder(c *gin.Context) {
	var req struct {
		Subject  string `json:"subject" binding:"required"`
		Body     string `json:"body"`
		Amount   int64  `json:"amount" binding:"required"`
		Channel  string `json:"channel" binding:"required"`
		OpenID   string `json:"openId"`
		Extra    string `json:"extra"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}

	if req.Amount <= 0 {
		common.Error(c, common.CodeBadRequest, "金额必须大于0")
		return
	}

	if req.Channel != "wechat" && req.Channel != "alipay" {
		common.Error(c, common.CodeBadRequest, "不支持的支付渠道")
		return
	}

	tenantID := common.GetTenantID(c)
	orderNo := fmt.Sprintf("PAY%d%04d", time.Now().UnixMilli(), time.Now().Nanosecond()%10000)

	result, err := ctl.paymentService.CreateOrderWithPayInfo(
		tenantID, orderNo, req.Subject, req.Body,
		req.Amount, req.Channel, req.OpenID, req.Extra,
	)
	if err != nil {
		log.Printf("[payment] create order failed: %v", err)
		common.Error(c, common.CodeInternalError, "创建订单失败")
		return
	}

	if result.PayError != nil {
		result.PayInfo["payError"] = result.PayError.Error()
	}

	common.Success(c, result.PayInfo)
}

func (ctl *PaymentController) GetOrder(c *gin.Context) {
	orderNo := c.Query("orderNo")
	if orderNo == "" {
		common.Error(c, common.CodeBadRequest, "订单号不能为空")
		return
	}

	tenantID := common.GetTenantID(c)
	order, err := ctl.paymentService.GetOrder(tenantID, orderNo)
	if err != nil {
		common.Error(c, common.CodeNotFound, "订单不存在")
		return
	}

	common.Success(c, order)
}

func (ctl *PaymentController) CloseOrder(c *gin.Context) {
	var req struct {
		OrderNo string `json:"orderNo" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}

	tenantID := common.GetTenantID(c)
	if err := ctl.paymentService.CloseOrder(tenantID, req.OrderNo); err != nil {
		common.Error(c, common.CodeInternalError, "关闭订单失败")
		return
	}

	common.Success(c, nil)
}

func (ctl *PaymentController) FindList(c *gin.Context) {
	subject := c.Query("subject")
	channel := c.Query("channel")
	status := -1
	if s := c.Query("status"); s != "" {
		fmt.Sscanf(s, "%d", &status)
	}
	page, pageSize := common.GetPageInfo(c)
	tenantID := common.GetTenantID(c)

	list, total, err := ctl.paymentService.FindList(tenantID, subject, int8(status), channel, page, pageSize)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.SuccessWithPage(c, list, total, page, pageSize)
}

func (ctl *PaymentController) WechatNotify(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("[pay-notify] wechat read body failed: %v", err)
		c.JSON(200, gin.H{"code": "FAIL", "message": "read body failed"})
		return
	}
	defer c.Request.Body.Close()

	log.Printf("[pay-notify] wechat received body length: %d", len(body))

	cfg := service.LoadWechatPayConfig()
	gw := service.NewWechatPayGateway(*cfg)

	result, err := gw.ParseNotify(body, c.Request.Header)
	if err != nil {
		log.Printf("[pay-notify] wechat parse failed: %v", err)
		c.JSON(200, gin.H{"code": "FAIL", "message": err.Error()})
		return
	}

	log.Printf("[pay-notify] wechat order_no=%s trade_no=%s status=%s", result.OrderNo, result.TradeNo, result.Status)

	if err := ctl.paymentService.HandleNotify("wechat", result); err != nil {
		log.Printf("[pay-notify] wechat handle failed: %v", err)
		c.JSON(200, gin.H{"code": "FAIL", "message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"code": "SUCCESS", "message": "成功"})
}

func (ctl *PaymentController) AlipayNotify(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("[pay-notify] alipay read body failed: %v", err)
		c.String(200, "fail")
		return
	}
	defer c.Request.Body.Close()

	log.Printf("[pay-notify] alipay received body length: %d", len(body))

	cfg := service.LoadAlipayConfig()
	gw := service.NewAlipayGateway(*cfg)

	result, err := gw.ParseNotify(body)
	if err != nil {
		log.Printf("[pay-notify] alipay parse failed: %v", err)
		c.String(200, "fail")
		return
	}

	log.Printf("[pay-notify] alipay order_no=%s trade_no=%s status=%s", result.OrderNo, result.TradeNo, result.Status)

	if err := ctl.paymentService.HandleNotify("alipay", result); err != nil {
		log.Printf("[pay-notify] alipay handle failed: %v", err)
		c.String(200, "fail")
		return
	}

	c.String(200, "success")
}

func (ctl *PaymentController) QueryOrder(c *gin.Context) {
	orderNo := c.Query("orderNo")
	if orderNo == "" {
		common.Error(c, common.CodeBadRequest, "订单号不能为空")
		return
	}

	tenantID := common.GetTenantID(c)
	order, err := ctl.paymentService.GetOrder(tenantID, orderNo)
	if err != nil {
		common.Error(c, common.CodeNotFound, "订单不存在")
		return
	}

	common.Success(c, gin.H{
		"orderNo": order.OrderNo,
		"status":  order.Status,
		"paidAt":  order.PaidAt,
	})
}

func (ctl *PaymentController) RefundOrder(c *gin.Context) {
	var req struct {
		OrderNo   string `json:"orderNo" binding:"required"`
		RefundAmt int64  `json:"refundAmt" binding:"required"`
		RefundNo  string `json:"refundNo"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeBadRequest, err.Error())
		return
	}

	if req.RefundAmt <= 0 {
		common.Error(c, common.CodeBadRequest, "退款金额必须大于0")
		return
	}

	if req.RefundNo == "" {
		req.RefundNo = fmt.Sprintf("REF%d%04d", time.Now().UnixMilli(), time.Now().Nanosecond()%10000)
	}

	tenantID := common.GetTenantID(c)
	result, err := ctl.paymentService.RefundOrderWithPayInfo(tenantID, req.OrderNo, req.RefundNo, req.RefundAmt)
	if err != nil {
		common.Error(c, common.CodeNotFound, err.Error())
		return
	}

	if result.Error != nil {
		log.Printf("[payment] refund failed: %v", result.Error)
		common.Error(c, common.CodeInternalError, result.Error.Error())
		return
	}

	common.Success(c, gin.H{
		"refundNo": result.RefundNo,
		"status":   result.Status,
	})
}
