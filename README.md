# Gin-Admin

[中文](README.md) | [English](README.en.md)

# Gin-Admin 后台管理框架

基于 **Go + Gin + GORM + Vue3 + Element Plus** 的中大型业务系统后台管理框架。

适用于活动管理、CRM、Agent管理平台、企业内部管理系统等场景。

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Vue](https://img.shields.io/badge/Vue-3.5-42b883?style=flat&logo=vue.js)](https://vuejs.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## 技术栈

### 后端 Backend

| 技术 | 版本 | 说明 |
|------|------|------|
| Go | 1.25 | 编程语言 |
| Gin | v1.10.0 | Web 框架 |
| GORM | v1.25.12 | ORM 框架 |
| MySQL | 5.7+ | 关系型数据库 |
| Redis | 3.0+ | 缓存/Token存储 |
| JWT | v5.2.3 | 认证授权 |
| Casbin | v2.103.0 | RBAC 权限控制 |
| Viper | v1.19.0 | 配置管理 |
| Zap | v1.27.0 | 结构化日志 |
| Swagger | v1.6.0 | API 文档 |
| Aliyun OSS | v3.0.2 | 阿里云对象存储 |
| Tencent COS | v0.7.74 | 腾讯云对象存储 |
| MinIO | v7.2.1 | S3 兼容存储 |

### 前端 Frontend

| 技术 | 版本 | 说明 |
|------|------|------|
| Vue | 3.5.13 | 渐进式框架 |
| Vue Router | 4.5.0 | 路由管理 |
| Pinia | 2.3.0 | 状态管理 |
| Element Plus | 2.9.1 | UI 组件库 |
| Axios | 1.7.9 | HTTP 客户端 |
| Vite | 6.0.5 | 构建工具 |
| TypeScript | 5.7.2 | 类型系统 |

## 项目结构

```
go-admin/
├── cmd/server/main.go              # 入口
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
│       ├── system/                 # 系统管理模块
│       ├── payment/                # 支付模块
│       ├── member/                 # 会员模块
│       └── captcha/                # 验证码
├── pkg/
│   ├── auth/jwt.go                 # JWT 工具
│   ├── upload/                     # 多端文件上传
│   ├── excel/excel.go              # Excel 导入导出
│   ├── task/cron.go                # 定时任务调度
│   └── utils/                      # 工具函数
├── router/router.go                # 路由注册
├── sql/init.sql                    # 数据库初始化
├── web/                            # 前端 (Vue3)
├── AGENTS.md                       # AI Agent 开发规范
└── README.md
```

## 架构设计

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

### 中间件顺序

```
全局: Recovery → Logger → Cors → Tenant
鉴权: Auth → CasbinAuth → OperationLog（仅受保护路由组）
```

### 安全特性

| 特性 | 实现 |
|------|------|
| 多租户隔离 | `TenantScope` 自动过滤，仅从 JWT 获取 tenant_id |
| 登录限频 | Redis IP 计数，5分钟/5次，超限锁定15分钟 |
| Token 吊销 | 密码修改/用户禁用后自动失效 |
| RBAC 权限 | Casbin 策略控制（空策略时跳过） |
| CORS 白名单 | 生产环境必须配置 `cors.allow_origins` |
| 文件上传校验 | 扩展名白名单 + 危险扩展名拦截 |
| 支付回调验签 | 微信平台证书 RSA + 支付宝签名验证 |
| 支付金额校验 | 回调时比对订单金额，防止篡改 |
| 密码存储 | bcrypt 哈希 |
| 开放重定向防护 | returnURL 协议和主机名校验 |

## 功能模块

### 系统管理

| 模块 | 说明 | API |
|------|------|-----|
| 用户管理 | 增删改查、状态管理、密码重置 | `/api/v1/system/user/*` |
| 角色管理 | 增删改查、菜单权限分配 | `/api/v1/system/role/*` |
| 菜单管理 | 目录/菜单/按钮三级管理 | `/api/v1/system/menu/*` |
| 部门管理 | 树形部门结构 | `/api/v1/system/dept/*` |
| 岗位管理 | 岗位增删改查 | `/api/v1/system/post/*` |
| 系统配置 | 参数键值配置 | `/api/v1/system/config/*` |
| 数据字典 | 字典类型 + 字典数据 | `/api/v1/system/dict/*` |
| 日志管理 | 操作日志 + 登录日志 | `/api/v1/system/log/*` |
| 附件管理 | 图片上传、预览、删除 | `/api/v1/system/file/*` |
| 协议管理 | 用户协议/隐私政策 | `/api/v1/system/agreement/*` |
| 仪表盘 | 数据统计概览 | `/api/v1/system/dashboard/*` |

### 支付管理

| 模块 | 说明 | API |
|------|------|-----|
| 创建订单 | 微信JSAPI/Native、支付宝H5/Page | `POST /api/v1/system/pay/order` |
| 订单查询 | 按订单号查询 | `GET /api/v1/system/pay/order` |
| 关闭订单 | 关闭待支付订单 | `POST /api/v1/system/pay/order/close` |
| 申请退款 | 微信/支付宝退款 | `POST /api/v1/system/pay/order/refund` |
| 支付回调 | 微信/支付宝异步通知 | `POST /api/v1/pay/notify/*` |

### 会员管理

| 模块 | 说明 | API |
|------|------|-----|
| 会员管理 | 会员增删改查、状态管理 | `/api/v1/member/*` |
| 会员等级 | 等级配置、升级规则 | `/api/v1/member/level/*` |
| 会员标签 | 标签管理、批量打标 | `/api/v1/member/tag/*` |
| 积分日志 | 积分变动记录 | `/api/v1/member/points/*` |

### 认证授权

- JWT 登录 + Refresh Token（Access 2h / Refresh 7d）
- Casbin RBAC 权限控制
- 用户 → 角色 → 菜单/按钮权限
- 点击验证码（防暴力破解）
- 登录限频（5分钟/5次，Redis 计数）
- Token 吊销（密码修改/禁用后立即失效）

### 文件上传

- 本地存储（默认）
- 阿里云 OSS
- 腾讯云 COS
- MinIO（S3 兼容）
- 启动时自动读取 `sys_config` 表 `oss.*` 配置选择存储方式

## 快速开始

### 环境要求

- Go 1.25+
- MySQL 5.7+
- Redis 3.0+
- Node.js 18+

### 1. 数据库初始化

```bash
mysql -u root -p < sql/init.sql
```

### 2. 启动后端

```bash
# 修改配置
vim config/config.yaml

# 编译运行
go mod tidy
go build -o server.exe ./cmd/server
./server.exe
```

### 3. 启动前端

```bash
cd web
npm install
npm run dev
```

### 4. 一键启动（Windows）

```powershell
.\start-all.ps1           # 同时启动前后端
.\start-backend.ps1       # 仅启动后端
.\start-frontend.ps1      # 仅启动前端
```

### 5. 访问

| 地址 | 说明 |
|------|------|
| http://localhost:3000 | 前端页面 |
| http://localhost:8080 | 后端 API |
| http://localhost:8080/swagger/index.html | API 文档 |

### 默认账号

- 用户名：`admin`
- 密码：`admin123`

## 配置说明

### 环境变量

敏感配置支持环境变量覆盖（优先级高于 config.yaml）：

```bash
export DB_PASSWORD="your_db_password"
export JWT_SECRET="your_jwt_secret"
export REDIS_PASSWORD="your_redis_password"
```

### CORS 配置

```yaml
cors:
  allow_origins:
    - "http://localhost:3000"
    - "http://localhost:5173"
  allow_methods: ["GET","POST","PUT","DELETE","OPTIONS","PATCH"]
  allow_headers: ["Origin","Content-Type","Accept","Authorization","X-Tenant-Id"]
  allow_credentials: true
```

### 支付配置（数据库 sys_config 表）

| Key | 说明 |
|-----|------|
| pay.wechat_app_id | 微信AppID |
| pay.wechat_mch_id | 微信商户号 |
| pay.wechat_key | 微信API密钥 |
| pay.alipay_app_id | 支付宝AppID |
| pay.alipay_key | 支付宝应用私钥 |
| pay.alipay_public_key | 支付宝公钥 |
| pay.notify_url | 统一回调地址 |

### OSS 配置（数据库 sys_config 表）

| Key | 说明 | 可选值 |
|-----|------|--------|
| oss.type | 存储类型 | local / aliyun / tencent / minio |
| oss.endpoint | Endpoint | |
| oss.bucket | Bucket名称 | |
| oss.access_key | AccessKey | |
| oss.secret_key | SecretKey | |
| oss.domain | 自定义域名 | |

## Makefile

```bash
make build         # 编译
make run           # 运行
make test          # 测试
make lint          # 静态检查 (go vet)
make swagger       # 生成 Swagger 文档
make deps          # 整理依赖 (go mod tidy)
make clean         # 清理编译产物
```

## 测试

```bash
go test ./...
```

## 打赏

如果这个项目对你有帮助，欢迎请作者喝杯咖啡~

| 微信 | 支付宝 |
|:---:|:---:|
| <img src="docs/images/weixin.jpg" width="200"> | <img src="docs/images/zhifubao.jpg" width="200"> |

## 许可证

[MIT License](LICENSE)
