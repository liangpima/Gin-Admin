package router

import (
	"net/http"

	"go-admin/internal/middleware"
	captchaController "go-admin/internal/module/captcha/controller"
	memberController "go-admin/internal/module/member/controller"
	paymentController "go-admin/internal/module/payment/controller"
	"go-admin/internal/module/system/controller"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// 权限码与 sys_menu.permission 一一对应。
// 「角色 -> 菜单」的授权关系由 sys_role_menu 维护，中间件据此自动生成 Casbin 策略，
// 因此新增接口只需在注册时声明权限码（或传空串表示仅需登录态），
// 无需再手工维护 casbin_rule 表。
const (
	permUserList   = "system:user:list"
	permUserAdd    = "system:user:add"
	permUserEdit   = "system:user:edit"
	permUserDelete = "system:user:delete"

	permRoleList   = "system:role:list"
	permRoleAdd    = "system:role:add"
	permRoleEdit   = "system:role:edit"
	permRoleDelete = "system:role:delete"

	permMenuList   = "system:menu:list"
	permMenuAdd    = "system:menu:add"
	permMenuEdit   = "system:menu:edit"
	permMenuDelete = "system:menu:delete"

	permDeptList   = "system:dept:list"
	permDeptAdd    = "system:dept:add"
	permDeptEdit   = "system:dept:edit"
	permDeptDelete = "system:dept:delete"

	permPostList   = "system:post:list"
	permPostAdd    = "system:post:add"
	permPostEdit   = "system:post:edit"
	permPostDelete = "system:post:delete"

	permDictList   = "system:dict:list"
	permDictAdd    = "system:dict:add"
	permDictDelete = "system:dict:delete"

	permConfigList   = "system:config:list"
	permConfigAdd    = "system:config:add"
	permConfigEdit   = "system:config:edit"
	permConfigDelete = "system:config:delete"

	permLogList   = "system:log:list"
	permLogDelete = "system:log:delete"

	permFileList   = "system:file:list"
	permFileDelete = "system:file:delete"

	permAgreementList   = "system:agreement:list"
	permAgreementAdd    = "system:agreement:add"
	permAgreementEdit   = "system:agreement:edit"
	permAgreementDelete = "system:agreement:delete"

	permMemberList   = "member:list"
	permMemberAdd    = "member:add"
	permMemberEdit   = "member:edit"
	permMemberDelete = "member:delete"

	permMemberLevelList   = "member:level:list"
	permMemberLevelAdd    = "member:level:add"
	permMemberLevelEdit   = "member:level:edit"
	permMemberLevelDelete = "member:level:delete"

	permMemberTagList   = "member:tag:list"
	permMemberTagAdd    = "member:tag:add"
	permMemberTagEdit   = "member:tag:edit"
	permMemberTagDelete = "member:tag:delete"

	permMemberPointsList = "member:points:list"

	permPayOrderList   = "payment:order:list"
	permPayOrderCreate = "payment:order:create"
	permPayOrderRefund = "payment:order:refund"
	permPayOrderClose  = "payment:order:close"
)

// protected 注册一条受保护路由，并登记其所需权限码。
// code 为空表示仅要求登录态（自助接口，如 userInfo / changePwd）。
// 组级鉴权中间件会依据登记表校验；未登记的路由会被拒绝。
func protected(g *gin.RouterGroup, method, path, code string, h gin.HandlerFunc) {
	middleware.RegisterPermission(method, g.BasePath()+path, code)
	g.Handle(method, path, h)
}

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

	// 公开接口：无需登录
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
		// 自助接口：仅要求登录态（权限码传空串）
		protected(authorized, http.MethodPost, "/auth/logout", "", authController.Logout)
		protected(authorized, http.MethodGet, "/auth/userInfo", "", authController.GetUserInfo)
		protected(authorized, http.MethodGet, "/dashboard/stats", "", dashboardController.GetStats)
		protected(authorized, http.MethodPut, "/system/user/changePwd", "", userController.ChangePassword)

		system := authorized.Group("/system")
		{
			protected(system, http.MethodPost, "/user", permUserAdd, userController.Create)
			protected(system, http.MethodPut, "/user", permUserEdit, userController.Update)
			protected(system, http.MethodDelete, "/user/:id", permUserDelete, userController.Delete)
			protected(system, http.MethodGet, "/user/:id", permUserList, userController.FindByID)
			protected(system, http.MethodGet, "/user/list", permUserList, userController.FindList)
			protected(system, http.MethodPut, "/user/status", permUserEdit, userController.UpdateStatus)
			protected(system, http.MethodPut, "/user/roles", permUserEdit, userController.UpdateRoles)
			protected(system, http.MethodPut, "/user/dept", permUserEdit, userController.UpdateDept)
			protected(system, http.MethodPut, "/user/resetPwd", permUserEdit, userController.ResetPassword)

			protected(system, http.MethodPost, "/role", permRoleAdd, roleController.Create)
			protected(system, http.MethodPut, "/role", permRoleEdit, roleController.Update)
			protected(system, http.MethodDelete, "/role/:id", permRoleDelete, roleController.Delete)
			protected(system, http.MethodGet, "/role/:id", permRoleList, roleController.FindByID)
			protected(system, http.MethodGet, "/role/list", permRoleList, roleController.FindList)
			protected(system, http.MethodPut, "/role/status", permRoleEdit, roleController.UpdateStatus)
			protected(system, http.MethodGet, "/role/all", permRoleList, roleController.FindAll)

			protected(system, http.MethodPost, "/menu", permMenuAdd, menuController.Create)
			protected(system, http.MethodPut, "/menu", permMenuEdit, menuController.Update)
			protected(system, http.MethodDelete, "/menu/:id", permMenuDelete, menuController.Delete)
			protected(system, http.MethodGet, "/menu/:id", permMenuList, menuController.FindByID)
			protected(system, http.MethodGet, "/menu/tree", permMenuList, menuController.FindTree)
			protected(system, http.MethodGet, "/menu/all", permMenuList, menuController.FindAll)

			protected(system, http.MethodPost, "/dept", permDeptAdd, deptController.Create)
			protected(system, http.MethodPut, "/dept", permDeptEdit, deptController.Update)
			protected(system, http.MethodDelete, "/dept/:id", permDeptDelete, deptController.Delete)
			protected(system, http.MethodGet, "/dept/:id", permDeptList, deptController.FindByID)
			protected(system, http.MethodGet, "/dept/tree", permDeptList, deptController.FindTree)

			protected(system, http.MethodPost, "/post", permPostAdd, postController.Create)
			protected(system, http.MethodPut, "/post", permPostEdit, postController.Update)
			protected(system, http.MethodDelete, "/post/:id", permPostDelete, postController.Delete)
			protected(system, http.MethodGet, "/post/list", permPostList, postController.FindList)

			protected(system, http.MethodPost, "/config", permConfigAdd, configController.Create)
			protected(system, http.MethodPut, "/config", permConfigEdit, configController.Update)
			protected(system, http.MethodDelete, "/config/:id", permConfigDelete, configController.Delete)
			protected(system, http.MethodGet, "/config/list", permConfigList, configController.FindList)
			protected(system, http.MethodGet, "/config/prefix", permConfigList, configController.FindByPrefix)
			protected(system, http.MethodPut, "/config/batch", permConfigEdit, configController.BatchSave)
			protected(system, http.MethodPost, "/config/upload", permConfigEdit, configController.UploadCert)

			protected(system, http.MethodPost, "/dict/type", permDictAdd, dictController.CreateType)
			protected(system, http.MethodDelete, "/dict/type/:id", permDictDelete, dictController.DeleteType)
			protected(system, http.MethodGet, "/dict/type/list", permDictList, dictController.FindTypeList)
			protected(system, http.MethodPost, "/dict/data", permDictAdd, dictController.CreateData)
			protected(system, http.MethodDelete, "/dict/data/:id", permDictDelete, dictController.DeleteData)
			protected(system, http.MethodGet, "/dict/data/list", permDictList, dictController.FindDataList)

			protected(system, http.MethodGet, "/log/operation", permLogList, logController.FindOperationLogList)
			protected(system, http.MethodGet, "/log/login", permLogList, logController.FindLoginLogList)
			protected(system, http.MethodDelete, "/log/operation", permLogDelete, logController.ClearOperationLogs)
			protected(system, http.MethodDelete, "/log/login", permLogDelete, logController.ClearLoginLogs)

			protected(system, http.MethodPost, "/agreement", permAgreementAdd, agreementController.Create)
			protected(system, http.MethodPut, "/agreement", permAgreementEdit, agreementController.Update)
			protected(system, http.MethodDelete, "/agreement/:id", permAgreementDelete, agreementController.Delete)
			protected(system, http.MethodGet, "/agreement/list", permAgreementList, agreementController.FindList)
			protected(system, http.MethodGet, "/agreement/type/:type", permAgreementList, agreementController.FindByType)

			protected(system, http.MethodPost, "/file/upload", permFileList, fileController.Upload)
			protected(system, http.MethodGet, "/file/list", permFileList, fileController.FindList)
			protected(system, http.MethodGet, "/file/:id", permFileList, fileController.FindByID)
			protected(system, http.MethodDelete, "/file/:id", permFileDelete, fileController.Delete)

			protected(system, http.MethodPost, "/pay/order", permPayOrderCreate, payController.CreateOrder)
			protected(system, http.MethodGet, "/pay/order", permPayOrderList, payController.GetOrder)
			protected(system, http.MethodPost, "/pay/order/close", permPayOrderClose, payController.CloseOrder)
			protected(system, http.MethodPost, "/pay/order/refund", permPayOrderRefund, payController.RefundOrder)
			protected(system, http.MethodGet, "/pay/order/list", permPayOrderList, payController.FindList)
			protected(system, http.MethodGet, "/pay/order/query", permPayOrderList, payController.QueryOrder)

			member := authorized.Group("/member")
			{
				protected(member, http.MethodPost, "", permMemberAdd, memberCtrl.Create)
				protected(member, http.MethodPut, "", permMemberEdit, memberCtrl.Update)
				protected(member, http.MethodDelete, "/:id", permMemberDelete, memberCtrl.Delete)
				protected(member, http.MethodGet, "/:id", permMemberList, memberCtrl.FindByID)
				protected(member, http.MethodGet, "/list", permMemberList, memberCtrl.FindList)
				protected(member, http.MethodPut, "/status", permMemberEdit, memberCtrl.UpdateStatus)
				protected(member, http.MethodPut, "/tags", permMemberEdit, memberCtrl.UpdateTags)
				protected(member, http.MethodPut, "/visit", permMemberEdit, memberCtrl.UpdateLastVisit)
				protected(member, http.MethodGet, "/level/all", permMemberLevelList, memberCtrl.FindAllLevels)
				protected(member, http.MethodGet, "/tag/all", permMemberTagList, memberCtrl.FindAllTags)

				protected(member, http.MethodPost, "/level", permMemberLevelAdd, memberLevelCtrl.Create)
				protected(member, http.MethodPut, "/level", permMemberLevelEdit, memberLevelCtrl.Update)
				protected(member, http.MethodDelete, "/level/:id", permMemberLevelDelete, memberLevelCtrl.Delete)
				protected(member, http.MethodGet, "/level/list", permMemberLevelList, memberLevelCtrl.FindList)

				protected(member, http.MethodPost, "/tag", permMemberTagAdd, memberTagCtrl.Create)
				protected(member, http.MethodPut, "/tag", permMemberTagEdit, memberTagCtrl.Update)
				protected(member, http.MethodDelete, "/tag/:id", permMemberTagDelete, memberTagCtrl.Delete)
				protected(member, http.MethodGet, "/tag/list", permMemberTagList, memberTagCtrl.FindList)

				protected(member, http.MethodGet, "/points/list", permMemberPointsList, pointsLogCtrl.FindList)
			}
		}
	}

	// 支付回调：由支付平台发起，自行验签，不做登录鉴权
	r.POST("/api/v1/pay/notify/wechat", payController.WechatNotify)
	r.POST("/api/v1/pay/notify/alipay", payController.AlipayNotify)

	// Swagger 文档仅在开发环境暴露
	if mode != "release" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// 静态文件服务 - 上传文件。
	// 该目录对匿名访问开放（前端需直链引用），因此补一层安全响应头，
	// 防止 .svg 等同源可解析文件被当作文档打开而触发存储型 XSS。
	uploads := r.Group("/uploads", middleware.UploadSecurity())
	uploads.Static("/", "uploads")

	return r
}
