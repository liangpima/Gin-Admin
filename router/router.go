package router

import (
	"go-admin/internal/middleware"
	captchaController "go-admin/internal/module/captcha/controller"
	memberController "go-admin/internal/module/member/controller"
	paymentController "go-admin/internal/module/payment/controller"
	"go-admin/internal/module/system/controller"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Setup(mode string) *gin.Engine {
	gin.SetMode(mode)

	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.Cors())
	r.Use(middleware.Tenant())

	fileController := controller.NewFileController()
	payController := paymentController.NewPaymentController()
	memberCtrl := memberController.NewMemberController()
	memberLevelCtrl := memberController.NewMemberLevelController()
	memberTagCtrl := memberController.NewMemberTagController()
	pointsLogCtrl := memberController.NewPointsLogController()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":    "Gin-Admin API",
			"version": "1.0.0",
			"docs":    "/swagger/index.html",
			"health":  "/health",
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")

	authController := controller.NewAuthController()
	userController := controller.NewUserController()
	roleController := controller.NewRoleController()
	menuController := controller.NewMenuController()
	deptController := controller.NewDeptController()
	dashboardController := controller.NewDashboardController()
	postController := controller.NewPostController()
	configController := controller.NewConfigController()
	dictController := controller.NewDictController()
	logController := controller.NewLogController()
	agreementController := controller.NewAgreementController()
	captchaCtrl := captchaController.NewCaptchaController()

	auth := api.Group("/auth")
	{
		auth.POST("/login", authController.Login)
		auth.POST("/refresh", authController.RefreshToken)
	}

	api.GET("/site/info", configController.SiteInfo)

	api.GET("/captcha/generate", captchaCtrl.Generate)
	api.POST("/captcha/verify", captchaCtrl.Verify)

	authorized := api.Group("")
	authorized.Use(middleware.Auth())
	authorized.Use(middleware.CasbinAuth())
	authorized.Use(middleware.OperationLog())
	{
		authorized.POST("/auth/logout", authController.Logout)
		authorized.GET("/auth/userInfo", authController.GetUserInfo)

		authorized.GET("/dashboard/stats", dashboardController.GetStats)

		system := authorized.Group("/system")
		{
			system.POST("/user", userController.Create)
			system.PUT("/user", userController.Update)
			system.DELETE("/user/:id", userController.Delete)
			system.GET("/user/:id", userController.FindByID)
			system.GET("/user/list", userController.FindList)
			system.PUT("/user/status", userController.UpdateStatus)
			system.PUT("/user/roles", userController.UpdateRoles)
			system.PUT("/user/dept", userController.UpdateDept)
			system.PUT("/user/resetPwd", userController.ResetPassword)
			system.PUT("/user/changePwd", userController.ChangePassword)

			system.POST("/role", roleController.Create)
			system.PUT("/role", roleController.Update)
			system.DELETE("/role/:id", roleController.Delete)
			system.GET("/role/:id", roleController.FindByID)
			system.GET("/role/list", roleController.FindList)
			system.PUT("/role/status", roleController.UpdateStatus)
			system.GET("/role/all", roleController.FindAll)

			system.POST("/menu", menuController.Create)
			system.PUT("/menu", menuController.Update)
			system.DELETE("/menu/:id", menuController.Delete)
			system.GET("/menu/:id", menuController.FindByID)
			system.GET("/menu/tree", menuController.FindTree)
			system.GET("/menu/all", menuController.FindAll)

			system.POST("/dept", deptController.Create)
			system.PUT("/dept", deptController.Update)
			system.DELETE("/dept/:id", deptController.Delete)
			system.GET("/dept/:id", deptController.FindByID)
			system.GET("/dept/tree", deptController.FindTree)

			system.POST("/post", postController.Create)
			system.PUT("/post", postController.Update)
			system.DELETE("/post/:id", postController.Delete)
			system.GET("/post/list", postController.FindList)

			system.POST("/config", configController.Create)
			system.PUT("/config", configController.Update)
			system.DELETE("/config/:id", configController.Delete)
			system.GET("/config/list", configController.FindList)
			system.GET("/config/prefix", configController.FindByPrefix)
			system.PUT("/config/batch", configController.BatchSave)
			system.POST("/config/upload", configController.UploadCert)

			system.POST("/dict/type", dictController.CreateType)
			system.DELETE("/dict/type/:id", dictController.DeleteType)
			system.GET("/dict/type/list", dictController.FindTypeList)
			system.POST("/dict/data", dictController.CreateData)
			system.DELETE("/dict/data/:id", dictController.DeleteData)
			system.GET("/dict/data/list", dictController.FindDataList)

			system.GET("/log/operation", logController.FindOperationLogList)
			system.GET("/log/login", logController.FindLoginLogList)

			system.POST("/agreement", agreementController.Create)
			system.PUT("/agreement", agreementController.Update)
			system.DELETE("/agreement/:id", agreementController.Delete)
			system.GET("/agreement/list", agreementController.FindList)
			system.GET("/agreement/type/:type", agreementController.FindByType)

			system.POST("/file/upload", fileController.Upload)
			system.GET("/file/list", fileController.FindList)
			system.GET("/file/:id", fileController.FindByID)
			system.DELETE("/file/:id", fileController.Delete)

			system.POST("/pay/order", payController.CreateOrder)
			system.GET("/pay/order", payController.GetOrder)
			system.POST("/pay/order/close", payController.CloseOrder)
			system.POST("/pay/order/refund", payController.RefundOrder)
			system.GET("/pay/order/list", payController.FindList)
			system.GET("/pay/order/query", payController.QueryOrder)

			member := authorized.Group("/member")
			{
				member.POST("", memberCtrl.Create)
				member.PUT("", memberCtrl.Update)
				member.DELETE("/:id", memberCtrl.Delete)
				member.GET("/:id", memberCtrl.FindByID)
				member.GET("/list", memberCtrl.FindList)
				member.PUT("/status", memberCtrl.UpdateStatus)
				member.PUT("/tags", memberCtrl.UpdateTags)
				member.PUT("/visit", memberCtrl.UpdateLastVisit)
				member.GET("/level/all", memberCtrl.FindAllLevels)
				member.GET("/tag/all", memberCtrl.FindAllTags)

				member.POST("/level", memberLevelCtrl.Create)
				member.PUT("/level", memberLevelCtrl.Update)
				member.DELETE("/level/:id", memberLevelCtrl.Delete)
				member.GET("/level/list", memberLevelCtrl.FindList)

				member.POST("/tag", memberTagCtrl.Create)
				member.PUT("/tag", memberTagCtrl.Update)
				member.DELETE("/tag/:id", memberTagCtrl.Delete)
				member.GET("/tag/list", memberTagCtrl.FindList)

				member.GET("/points/list", pointsLogCtrl.FindList)
			}
		}
	}

	r.POST("/api/v1/pay/notify/wechat", payController.WechatNotify)
	r.POST("/api/v1/pay/notify/alipay", payController.AlipayNotify)

	// Swagger 文档仅在开发环境暴露
	if mode != "release" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// 静态文件服务 - 上传文件
	r.Static("/uploads", "uploads")

	return r
}
