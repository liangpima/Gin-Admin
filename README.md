# Gin-Admin 后台管理框架

基于 **Go + Gin + GORM + Vue3 + Element Plus** 的中大型业务系统后台管理框架。

适用于活动管理、CRM、Agent管理平台、企业内部管理系统等场景。

## 技术栈

### 后端

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
| robfig/cron | v3.0.1 | 定时任务 |
| excelize | v2.9.0 | Excel 导入导出 |

### 前端

| 技术 | 版本 | 说明 |
|------|------|------|
| Vue | 3.5.13 | 渐进式框架 |
| Vue Router | 4.5.0 | 路由管理 |
| Pinia | 2.3.0 | 状态管理 |
| Element Plus | 2.9.1 | UI 组件库 |
| Axios | 1.7.9 | HTTP 客户端 |
| Vite | 6.0.5 | 构建工具 |
| TypeScript | 5.7.2 | 类型系统 |
| WangEditor | 5.1.23 | 富文本编辑器 |

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
│   │   ├── response.go             # 统一 Response
│   │   ├── model.go                # BaseModel / TenantBaseModel
│   │   ├── tenant.go               # TenantScope 租户过滤
│   │   ├── errors.go               # 错误码定义
│   │   ├── constants.go            # 常量
│   │   ├── context.go              # Context 工具
│   │   └── pagination.go           # 分页工具
│   ├── database/mysql.go           # MySQL 连接
│   ├── logger/zap.go               # Zap 日志
│   ├── middleware/                  # 7个中间件
│   │   ├── auth.go                 # JWT 认证
│   │   ├── casbin.go               # RBAC 鉴权
│   │   ├── cors.go                 # CORS
│   │   ├── logger.go               # 请求日志
│   │   ├── operation_log.go        # 操作审计
│   │   ├── recovery.go             # Panic 恢复
│   │   └── tenant.go               # 多租户
│   └── module/
│       ├── system/                 # 系统管理模块
│       ├── payment/                # 支付模块
│       ├── member/                 # 会员模块
│       ├── captcha/                # 验证码
│       └── monitor/                # 监控（占位）
├── pkg/
│   ├── auth/jwt.go                 # JWT 工具
│   ├── upload/                     # 多端文件上传
│   │   ├── upload.go               # 上传入口（自动选择存储方式 + 扩展名校验）
│   │   ├── local.go                # 本地存储
│   │   ├── aliyun_oss.go           # 阿里云 OSS
│   │   ├── tencent_cos.go          # 腾讯云 COS
│   │   └── minio.go                # MinIO
│   ├── excel/excel.go              # Excel 导入导出
│   ├── task/cron.go                # 定时任务调度
│   └── utils/
│       ├── hash.go                 # 哈希工具
│       ├── snowflake.go            # 雪花 ID
│       └── string.go               # 字符串工具
├── router/router.go                # 路由注册
├── sql/
│   └── init.sql                    # 数据库初始化（建表+种子数据）
├── docs/                           # Swagger 文档
├── web/                            # 前端 (Vue3)
│   └── src/
│       ├── api/                    # API 接口（16个）
│       ├── components/             # 公共组件（10个）
│       │   ├── ClickCaptcha/       # 点击验证码
│       │   ├── ImagePicker/        # 图片选择器
│       │   ├── MobileAction/       # 移动端操作折叠
│       │   ├── PageHeader/         # 页头
│       │   ├── Pagination/         # 分页
│       │   ├── RightPanel/         # 右侧面板
│       │   ├── SvgIcon/            # SVG 图标
│       │   ├── TableSkeleton/      # 表格骨架屏
│       │   ├── Upload/             # 文件上传
│       │   └── WangEditor/         # 富文本编辑器
│       ├── hooks/                  # useResponsive, useTheme
│       ├── layout/                 # 布局组件
│       ├── router/                 # 路由配置
│       ├── store/modules/          # app/permission/tagsView/user
│       ├── types/                  # TypeScript 类型
│       ├── utils/                  # auth.ts, format.ts, request.ts
│       └── views/                  # 页面
├── Makefile                        # 构建脚本
├── start-all.ps1                   # 一键启动
├── start-backend.ps1               # 仅启动后端
├── start-frontend.ps1              # 仅启动前端
├── AGENTS.md                       # AI Agent 开发规范
├── go.mod
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

### 网站设置

| 模块 | 说明 |
|------|------|
| 网站设置 | 网站名称、标题、Logo上传、描述、版权、ICP备案 |
| 支付设置 | 微信支付/支付宝参数配置、证书上传 |
| OSS存储设置 | 阿里云OSS/腾讯云COS/MinIO/本地存储切换 |
| 短信设置 | 短信服务商配置 |

### 支付管理

| 模块 | 说明 | API |
|------|------|-----|
| 创建订单 | 微信JSAPI/Native、支付宝H5/Page | `POST /api/v1/system/pay/order` |
| 订单查询 | 按订单号查询 | `GET /api/v1/system/pay/order` |
| 关闭订单 | 关闭待支付订单 | `POST /api/v1/system/pay/order/close` |
| 申请退款 | 微信/支付宝退款 | `POST /api/v1/system/pay/order/refund` |
| 订单列表 | 分页查询、筛选 | `GET /api/v1/system/pay/order/list` |
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

## 数据库设计

### 核心表

| 表名 | 说明 |
|------|------|
| sys_user | 用户表 |
| sys_role | 角色表 |
| sys_menu | 菜单表 |
| sys_dept | 部门表 |
| sys_post | 岗位表 |
| sys_config | 系统配置 |
| sys_dict_type | 字典类型 |
| sys_dict_data | 字典数据 |
| sys_operation_log | 操作日志 |
| sys_login_log | 登录日志 |
| sys_file | 文件表 |
| sys_tenant | 租户表 |
| sys_user_role | 用户角色关联 |
| sys_role_menu | 角色菜单关联 |
| sys_user_post | 用户岗位关联 |
| sys_agreement | 用户协议 |
| pay_order | 支付订单 |
| member | 会员表 |
| member_level | 会员等级 |
| member_tag | 会员标签 |
| member_points_log | 积分日志 |

### ER 关系

```
sys_user ──< sys_user_role >── sys_role
sys_role ──< sys_role_menu >── sys_menu
sys_user ──< sys_user_post >── sys_post
sys_dept ──< sys_user
sys_menu ──< sys_menu (自关联)
sys_dict_type ──< sys_dict_data
member ──< member_points_log
member ──< member_tag (多对多)
pay_order (独立支付订单)
```

## 快速开始

### 环境要求

- Go 1.25+
- MySQL 5.7+
- Redis 3.0+
- Node.js 18+

### 1. 数据库初始化

```bash
# 登录 MySQL 执行初始化脚本（一次执行完成所有建表和数据初始化）
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

### 应用配置 `config/config.yaml`

```yaml
server:
  port: 8080
  mode: debug
  read_timeout: 60
  write_timeout: 60

database:
  host: 127.0.0.1
  port: 3306
  username: root
  password: "your_password" # 生产环境使用 DB_PASSWORD 环境变量
  dbname: go_admin

redis:
  addr: 127.0.0.1:6379
  db: 0

jwt:
  secret: "your-jwt-secret" # 生产环境使用 JWT_SECRET 环境变量
  access_expire: 7200      # 2小时
  refresh_expire: 604800   # 7天

log:
  level: info
  max_size: 200
  max_backups: 7

upload:
  save_path: "./uploads"
  max_size: 10
  allow_exts: ".jpg,.jpeg,.png,.gif,.pdf,.doc,.docx,.xls,.xlsx"

casbin:
  model_path: "config/casbin/model.conf"

cors:
  allow_origins:
    - "http://localhost:3000"
    - "http://localhost:5173"
  allow_methods: ["GET","POST","PUT","DELETE","OPTIONS","PATCH"]
  allow_headers: ["Origin","Content-Type","Accept","Authorization","X-Tenant-Id"]
  allow_credentials: true
```

### 环境变量

敏感配置支持环境变量覆盖（优先级高于 config.yaml）：

```bash
export DB_PASSWORD="your_db_password"     # 数据库密码
export JWT_SECRET="your_jwt_secret"       # JWT 签名密钥
export REDIS_PASSWORD="your_redis_password" # Redis 密码
```

### 支付配置（数据库 sys_config 表）

| Key | 说明 | 示例 |
|-----|------|------|
| pay.wechat_app_id | 微信AppID | wx1234567890 |
| pay.wechat_mch_id | 微信商户号 | 1234567890 |
| pay.wechat_key | 微信API密钥 | 32位密钥 |
| pay.wechat_serial_no | 微信证书序列号 | 7位以上 |
| pay.alipay_app_id | 支付宝AppID | 2021001234567890 |
| pay.alipay_key | 支付宝应用私钥 | RSA私钥 |
| pay.alipay_public_key | 支付宝公钥 | RSA公钥 |
| pay.notify_url | 统一回调地址 | https://domain/api/v1/pay/notify/wechat |
| pay.return_url | 支付宝跳转地址 | https://domain/pay/result |

### OSS 配置（数据库 sys_config 表）

| Key | 说明 | 可选值 |
|-----|------|--------|
| oss.type | 存储类型 | local / aliyun / tencent / minio |
| oss.endpoint | Endpoint | aliyun: oss-cn-hangzhou.aliyuncs.com |
| oss.bucket | Bucket名称 | my-bucket |
| oss.access_key | AccessKey | |
| oss.secret_key | SecretKey | |
| oss.domain | 自定义域名 | https://cdn.example.com |

## API 接口概览

### 认证

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/login` | 登录 |
| POST | `/api/v1/auth/refresh` | 刷新Token |
| POST | `/api/v1/auth/logout` | 退出 |
| GET | `/api/v1/auth/userInfo` | 获取用户信息 |

### 系统管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST/PUT/DELETE | `/api/v1/system/user/*` | 用户 CRUD |
| GET/POST/PUT/DELETE | `/api/v1/system/role/*` | 角色 CRUD |
| GET/POST/PUT/DELETE | `/api/v1/system/menu/*` | 菜单 CRUD |
| GET/POST/PUT/DELETE | `/api/v1/system/dept/*` | 部门 CRUD |
| GET/POST/PUT/DELETE | `/api/v1/system/post/*` | 岗位 CRUD |
| GET/POST/PUT/DELETE | `/api/v1/system/config/*` | 配置 CRUD |
| GET/POST/PUT/DELETE | `/api/v1/system/dict/*` | 字典 CRUD |
| GET | `/api/v1/system/log/*` | 日志查询 |
| POST/GET/DELETE | `/api/v1/system/file/*` | 文件管理 |

### 支付

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/system/pay/order` | 创建订单 |
| GET | `/api/v1/system/pay/order` | 查询订单 |
| POST | `/api/v1/system/pay/order/close` | 关闭订单 |
| POST | `/api/v1/system/pay/order/refund` | 申请退款 |
| GET | `/api/v1/system/pay/order/list` | 订单列表 |
| GET | `/api/v1/system/pay/order/query` | 查询支付状态 |
| POST | `/api/v1/pay/notify/wechat` | 微信回调（无需鉴权） |
| POST | `/api/v1/pay/notify/alipay` | 支付宝回调（无需鉴权） |

### 会员

| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST/PUT/DELETE | `/api/v1/member/*` | 会员 CRUD |
| GET/POST/PUT/DELETE | `/api/v1/member/level/*` | 等级管理 |
| GET/POST/PUT/DELETE | `/api/v1/member/tag/*` | 标签管理 |
| GET | `/api/v1/member/points/*` | 积分日志 |

### 其他

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/captcha/*` | 验证码 |
| GET | `/api/v1/site/info` | 网站信息（公开） |
| GET | `/health` | 健康检查 |
| GET | `/swagger/*any` | API 文档 |

## 响应式适配

| 断点 | 宽度 | 效果 |
|------|------|------|
| Mobile | <768px | 侧边栏折叠、搜索栏堆叠、弹窗92%、操作列折叠（MobileAction） |
| Tablet | 768-1024px | 侧边栏折叠、搜索栏2列、弹窗80% |
| Desktop | >1024px | 完整布局 |

- 使用 `src/hooks/useResponsive.ts` 获取设备状态
- 搜索栏使用 `el-row` + `el-col` 响应式断点

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
# 运行全部测试
go test ./...

# 运行支付模块测试
go test ./internal/module/payment/... -v

# 运行会员模块测试
go test ./internal/module/member/... -v

# 生成 Swagger 文档
swag init -g cmd/server/main.go -o docs
```

## 打赏

如果这个项目对你有帮助，欢迎请作者喝杯咖啡~

| 微信 | 支付宝 |
|:---:|:---:|
| <img src="docs/images/weixin.jpg" width="200"> | <img src="docs/images/zhifubao.jpg" width="200"> |

## 许可证

MIT License
