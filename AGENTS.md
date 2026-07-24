# AGENTS.md — Gin-Admin 项目开发规范

## 项目概述

Gin-Admin 是一套基于 Go + Gin + GORM + Vue3 + Element Plus 的后台管理框架。
本文件定义了 AI Agent 在本项目中生成代码时必须遵守的规则和规范。

## 核心架构

采用 **Controller → Service → Repository** 三层架构。

```
Controller (接口层)
  ├── 参数接收 (ShouldBindJSON/ShouldBindQuery)
  ├── 参数校验 (binding tag)
  └── 返回统一 Response

Service (业务层)
  ├── 业务逻辑
  ├── 事务控制 (db.Transaction)
  └── 调用本模块 Repository

Repository (数据层)
  ├── 数据库操作 (GORM)
  └── 只做数据访问，不含业务逻辑
```

## 项目技术栈

- **后端**: Go 1.25, Gin v1.10, GORM v1.25, MySQL 5.7+, Redis 3.0+
- **认证**: JWT (golang-jwt/v5) + Casbin RBAC
- **前端**: Vue 3.5, Element Plus 2.9, Vite 6, TypeScript 5.7, Pinia 2.3
- **其他**: Zap 日志, Viper 配置, Swagger 文档, 1Password/robfig/cron

## 目录结构

```
go-admin/
├── cmd/server/main.go              # 唯一入口
├── config/
│   ├── config.yaml                 # 应用配置
│   ├── config.go                   # 配置加载
│   └── casbin/model.conf           # RBAC 模型
├── internal/
│   ├── cache/redis.go              # Redis 封装（可选）
│   ├── common/                     # 统一响应/错误码/模型/分页
│   ├── database/mysql.go           # MySQL 连接
│   ├── logger/zap.go               # Zap 日志
│   ├── middleware/                  # 7个中间件
│   └── module/
│       ├── system/                 # 系统管理（用户/角色/菜单/部门/岗位/配置/字典/日志/文件/协议）
│       ├── payment/                # 支付模块
│       ├── member/                 # 会员模块（会员/等级/标签/积分）
│       ├── captcha/                # 验证码
│       └── monitor/                # 监控（占位，未实现）
├── pkg/
│   ├── auth/jwt.go                 # JWT 工具
│   ├── upload/                     # 多端文件上传（本地/OSS/COS/MinIO）
│   │   ├── upload.go               # 上传入口（自动选择存储方式 + 扩展名校验）
│   │   ├── local.go                # 本地存储
│   │   ├── aliyun_oss.go           # 阿里云 OSS
│   │   ├── tencent_cos.go          # 腾讯云 COS
│   │   └── minio.go                # MinIO
│   ├── excel/excel.go              # Excel 导入导出
│   ├── task/cron.go                # 定时任务调度
│   └── utils/                      # Hash/Snowflake/字符串工具
├── router/router.go               # 路由注册（含 /uploads 静态服务）
├── sql/                            # 数据库脚本
├── docs/                           # Swagger 文档
├── web/                            # 前端 (Vue3)
│   └── src/
│       ├── api/                    # API 接口定义（16个）
│       ├── components/             # 公共组件（10个）
│       ├── hooks/                  # useResponsive, useTheme
│       ├── layout/                 # 布局组件
│       ├── router/                 # 路由配置
│       ├── store/modules/          # app/permission/tagsView/user
│       ├── utils/                  # auth/format/request
│       └── views/                  # 页面（7个目录）
├── Makefile
├── start-all.ps1                   # 一键启动脚本
└── AGENTS.md
```

## 开发规则（12条铁律）

### 规则1: Controller 只负责参数接收与返回

- **允许**: 参数接收 (`ShouldBindJSON`/`ShouldBindQuery`)、参数校验 (`binding` tag)、返回结果 (`common.Success`/`common.Error`)
- **禁止**: 在 Controller 中编写业务逻辑、直接操作数据库、调用 Repository

### 规则2: Service 负责业务逻辑与事务控制

- **允许**: 业务判断、数据组装、事务管理 (`db.Transaction`)、调用本模块 Repository
- **禁止**: 直接返回 HTTP 响应、操作 `gin.Context`、引入 `net/http` 相关依赖

### 规则3: Repository 只负责数据库操作

- **允许**: GORM 查询、CRUD 操作、SQL 构建
- **禁止**: 业务逻辑判断、跨表关联查询（应通过 Service 组合）、返回 HTTP 响应

### 规则4: 禁止跨模块访问 Repository

每个 Service 只能访问自己模块的 Repository。

```
✅ 允许:
UserService  → UserRepository
RoleService  → RoleRepository

❌ 禁止:
PaymentService → UserRepository   (跨模块)
OrderService   → UserRepository   (跨模块)
```

如需跨模块数据，通过调用对应 Service 实现。

### 规则5: 所有接口必须返回统一 Response 结构

```go
// 成功
common.Success(c, data)
common.SuccessWithPage(c, list, total, page, pageSize)

// 失败
common.Error(c, common.CodeBadRequest, "错误信息")
common.Unauthorized(c, "未登录")
common.Forbidden(c, "无权限")
```

禁止直接使用 `c.JSON()` 返回业务数据。

### 规则6: 所有表必须包含基础字段

```go
type BaseModel struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    CreateBy  uint           `gorm:"comment:创建者ID" json:"createBy"`
    UpdateBy  uint           `gorm:"comment:更新者ID" json:"updateBy"`
    CreatedAt time.Time      `gorm:"comment:创建时间" json:"createdAt"`
    UpdatedAt time.Time      `gorm:"comment:更新时间" json:"updatedAt"`
    DeletedAt gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
    Remark    string         `gorm:"type:varchar(500);comment:备注" json:"remark"`
}
```

多租户表继承 `TenantBaseModel`（额外包含 `TenantID`）。

### 规则7: 多租户必须自动附加 tenant_id

- 使用 `common.TenantScope(db, tenantID)` 在 Repository 层过滤租户数据
- 禁止无条件查询全表
- Repository 方法必须接收 `tenantID uint` 参数
- Controller 从 JWT claims 中提取 tenantID，透传给 Service → Repository
- 支付回调等外部接口可使用 `FindByXxxForNotify()`（无 tenant 过滤）

```go
// ✅ 允许:
func (r *memberRepository) FindByID(tenantID, id uint) (*model.Member, error) {
    return common.TenantScope(database.DB, tenantID).First(&member, id).Error
}

// ❌ 禁止:
func (r *memberRepository) FindByID(id uint) (*model.Member, error) {
    return database.DB.First(&member, id).Error  // 缺少 tenant 过滤
}
```

### 规则8: 禁止直接使用 db.Where()

所有数据库查询必须通过 Repository 封装。

```
❌ 禁止:
db.Where("username = ?", username).First(&user)

✅ 允许:
userRepository.FindByUsername(username)
```

### 规则9: 新增功能优先复用已有 Service

- 优先调用已有 Service 获取数据，禁止重复实现
- 新模块应依赖现有 Service，而非直接访问其 Repository
- Service 间互相调用通过接口，禁止直接 `NewXxxRepository()` 绕过 Service 层

### 规则10: 生成代码前先分析项目现有结构

在编写新代码前，必须：
1. 查看现有模块结构（model/repository/service/controller）
2. 遵循现有命名规范和代码风格
3. 复用已有的公共组件和工具函数（如 `pkg/utils/`、`internal/common/`）
4. 确保与现有架构一致

### 规则11: 前端页面必须支持响应式

所有前端页面必须适配多端显示：

- **搜索栏**: 使用 `el-row` + `el-col` 响应式断点（`:xs="24" :sm="12" :md="8" :lg="6"`）
- **表格**: 小屏启用横向滚动
- **弹窗**: 小屏（<768px）自动切换为 92% 宽度
- **统计卡片**: 使用响应式列数（`:xs="24" :sm="12" :md="6"`）
- **操作列**: 小屏使用 `MobileAction` 组件折叠操作按钮
- **侧边栏**: 移动端自动折叠为图标模式
- **断点规范**: mobile(<768px) / tablet(768-1024px) / desktop(>1024px)

使用 `src/hooks/useResponsive.ts` 获取设备状态，使用 `src/assets/styles/responsive.scss` 的 mixin。

### 规则12: 支付回调接口不做鉴权

支付回调接口（微信/支付宝通知）不经过 Auth 中间件，直接在 `router/router.go` 顶层注册：

```go
r.POST("/api/v1/pay/notify/wechat", payController.WechatNotify)
r.POST("/api/v1/pay/notify/alipay", payController.AlipayNotify)
```

回调接口必须自行验签，防止伪造请求。

## 中间件使用顺序

路由组注册时的中间件应用顺序（`router/router.go`）：

```
全局中间件: Recovery → Logger → Cors → Tenant
鉴权中间件: Auth → CasbinAuth → OperationLog（仅受保护路由组）
```

## 后端目录规范

```
internal/module/<模块名>/
├── controller/    # 接口层（1个文件 = 1个功能域）
├── service/       # 业务层（与 controller 一一对应）
├── repository/    # 数据层（与 controller 一一对应）
├── model/         # 数据模型（继承 BaseModel 或 TenantBaseModel）
├── dto/           # 请求 DTO（binding tag 校验）
└── vo/            # 响应 VO
```

## 前端目录规范

```
web/src/
├── api/<模块名>.ts      # API 接口定义（与后端路由一一对应）
├── views/<模块名>/      # 页面组件（支持响应式）
├── components/          # 公共组件（11个：ClickCaptcha, FormDialog, ImagePicker, MobileAction, PageHeader, Pagination, RightPanel, SvgIcon, TableSkeleton, Upload, WangEditor）
├── hooks/               # useResponsive, useTheme
├── store/modules/       # app/permission/tagsView/user
├── utils/               # auth.ts, format.ts, request.ts
└── layout/              # 布局组件
```

## 已实现模块

### system 模块（系统管理）

| 功能 | API 路径 |
|------|----------|
| 用户管理 | `/api/v1/system/user/*` |
| 角色管理 | `/api/v1/system/role/*` |
| 菜单管理 | `/api/v1/system/menu/*` |
| 部门管理 | `/api/v1/system/dept/*` |
| 岗位管理 | `/api/v1/system/post/*` |
| 系统配置 | `/api/v1/system/config/*` |
| 数据字典 | `/api/v1/system/dict/*` |
| 日志管理 | `/api/v1/system/log/*` |
| 文件管理 | `/api/v1/system/file/*` |
| 协议管理 | `/api/v1/system/agreement/*` |
| 仪表盘 | `/api/v1/system/dashboard/*` |

### payment 模块（支付管理）

- 支付订单 CRUD
- 微信支付 V3（JSAPI/Native/退款/查询）
- 支付宝开放平台（H5/App/Page/退款/查询）
- 微信/支付宝回调处理（独立验签，无鉴权）

### member 模块（会员管理）

- 会员管理、会员等级、会员标签、积分日志

### captcha 模块（验证码）

- 验证码生成与验证（无 Repository 层）

## 共享包（pkg/）

| 包 | 用途 |
|----|------|
| `pkg/auth` | JWT Token 创建与验证 |
| `pkg/upload` | 多端文件上传（本地/OSS/COS/MinIO），启动时从 sys_config 读取 oss.* 配置自动选择 |
| `pkg/excel` | Excel 导入导出（excelize） |
| `pkg/task` | 定时任务调度（robfig/cron） |
| `pkg/utils` | Hash/Snowflake ID/字符串工具 |

## 代码风格

- Go 代码遵循 `gofmt` 标准格式
- 错误处理必须显式检查，禁止 `_` 忽略关键错误
- 所有公开函数必须有注释（Swagger 格式）
- DTO 使用 `binding` tag 进行参数校验
- Model 使用 `gorm` tag 定义数据库字段
- JSON 字段使用小驼峰命名
- 日志统一使用 `internal/logger` 的 Zap 实例，禁止使用 `log.Printf`
- 前端 API 文件与后端路由模块一一对应
- 前端组件优先使用 Element Plus 内置组件

## 常用命令

```bash
# 后端
go mod tidy                    # 整理依赖
go build -o server.exe ./cmd/server  # 编译
go test ./...                  # 运行测试
go test ./internal/module/payment/... -v  # 模块测试
swag init -g cmd/server/main.go -o docs  # 生成 Swagger 文档
go vet ./...                   # 静态检查

# 前端
cd web && npm install          # 安装依赖
npm run dev                    # 开发服务器
npm run build                  # 构建（含类型检查）

# Makefile
make build                     # 编译
make run                       # 运行
make test                      # 测试
make lint                      # 静态检查
make swagger                   # 生成文档
make deps                      # 整理依赖

# 一键启动（Windows）
.\run.bat                     # 杀旧进程 + 编译 + 启动后端
.\start-all.ps1               # 同时启动前后端
.\start-backend.ps1           # 仅启动后端（不杀旧进程）
.\start-frontend.ps1          # 仅启动前端
```

## 安全规范

### 认证与授权

- JWT Secret 通过环境变量 `JWT_SECRET` 注入，禁止硬编码
- Access Token 有效期 2 小时，Refresh Token 7 天
- 登录限频：同一 IP 5 分钟内最多 5 次失败，超限锁定 15 分钟
- 密码修改/用户禁用后自动吊销所有 Token（Redis 黑名单）
- 退出登录时将 Access Token 加入 Redis 黑名单（`cache.RevokeToken`），Auth 中间件检查 `IsTokenRevoked`
- Casbin RBAC 中间件已挂载，空策略时跳过检查（开发兼容）

### 密码安全

- 密码使用 bcrypt 哈希（`bcrypt.DefaultCost`）
- 密码强度校验：至少包含大写字母、小写字母、数字中的两种，禁止包含空格
- 适用场景：用户创建（`Create`）、密码重置（`ResetPassword`）、密码修改（`ChangePassword`）

### 多租户隔离

- 所有 Repository 查询必须通过 `common.TenantScope(db, tenantID)` 过滤
- tenant_id 仅从 JWT claims 获取，禁止从 Header/Query 参数读取
- 关联表操作（如 `ReplaceTags`、`FindTagIDsByMemberID`）也必须使用 `TenantScope`
- 支付回调等外部接口使用 `FindByXxxForNotify()`（无 tenant 过滤）

### 文件上传

- 扩展名白名单校验（jpg/png/gif/bmp/svg/webp/mp4/mov/mp3/pdf/doc/xls/ppt/zip 等）
- 危险扩展名拦截（php/exe/sh/bat/js/vbs 等）
- 文件大小限制从 `config.yaml` 的 `upload.max_size` 读取（单位 MB），默认 10MB

### CORS 配置

- 生产环境必须在 `config.yaml` 的 `cors.allow_origins` 配置白名单
- 开发环境（mode: debug）允许所有来源
- `AllowCredentials: true` 必须配合明确的 Origin 白名单

### 支付安全

- 微信支付回调验签已实现（平台证书获取 + RSA 验签）
- 支付宝回调验签已实现
- 回调金额校验（防止金额篡改）
- returnURL 开放重定向防护（协议和主机名校验）

### 敏感配置

```bash
# 环境变量覆盖（优先级高于 config.yaml）
export DB_PASSWORD="your_db_password"
export JWT_SECRET="your_jwt_secret"
export REDIS_PASSWORD="your_redis_password"
```

### 前端安全

- Token 存储在 Cookie 中，设置 `sameSite: Lax` + `secure`（HTTPS 时）
- 无 `v-html` / `innerHTML` 使用（XSS 安全）
- Swagger 文档仅在非 release 模式暴露

## 新业务模块接入清单

1. 在 `internal/module/<模块名>/model/` 创建数据模型（继承 BaseModel 或 TenantBaseModel）
2. 在 `internal/module/<模块名>/repository/` 创建数据访问层
3. 在 `internal/module/<模块名>/service/` 创建业务逻辑层
4. 在 `internal/module/<模块名>/dto/` 创建请求 DTO
5. 在 `internal/module/<模块名>/vo/` 创建响应 VO（如需要）
6. 在 `internal/module/<模块名>/controller/` 创建接口层
7. 在 `router/router.go` 注册路由（鉴权路由放 Auth 组内）
8. 在 `sql/init.sql` 添加建表语句和初始数据
11. 在 `web/src/api/` 创建前端 API 文件
12. 在 `web/src/views/` 创建前端页面（支持响应式）
13. 在 controller 方法上添加 Swagger 注解
14. 运行 `swag init -g cmd/server/main.go -o docs` 生成文档
15. 页面必须支持响应式（搜索栏用 `el-row`/`el-col` 断点，弹窗小屏适配）
