# Gin-Admin

[中文](README.md) | [English](README.en.md)

# Gin-Admin Admin Panel Framework

A full-featured admin panel built with **Go + Gin + GORM + Vue3 + Element Plus**.

Suitable for CRM, agent management platforms, enterprise internal management systems, and more.

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Vue](https://img.shields.io/badge/Vue-3.5-42b883?style=flat&logo=vue.js)](https://vuejs.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## Tech Stack

### Backend

| Technology | Version | Description |
|------------|---------|-------------|
| Go | 1.25 | Programming Language |
| Gin | v1.10.0 | Web Framework |
| GORM | v1.25.12 | ORM Framework |
| MySQL | 5.7+ | Relational Database |
| Redis | 3.0+ | Cache / Token Storage |
| JWT | v5.2.3 | Authentication |
| Casbin | v2.103.0 | RBAC Authorization |
| Viper | v1.19.0 | Configuration |
| Zap | v1.27.0 | Structured Logging |
| Swagger | v1.6.0 | API Documentation |

### Frontend

| Technology | Version | Description |
|------------|---------|-------------|
| Vue | 3.5.13 | Progressive Framework |
| Vue Router | 4.5.0 | Routing |
| Pinia | 2.3.0 | State Management |
| Element Plus | 2.9.1 | UI Component Library |
| Axios | 1.7.9 | HTTP Client |
| Vite | 6.0.5 | Build Tool |
| TypeScript | 5.7.2 | Type System |

## Project Structure

```
go-admin/
├── cmd/server/main.go              # Entry point
├── config/                         # Configuration
│   ├── config.yaml
│   ├── config.go
│   └── casbin/model.conf
├── internal/
│   ├── cache/redis.go              # Redis (optional)
│   ├── common/                     # Response / Errors / Models / Pagination
│   ├── database/mysql.go           # MySQL connection
│   ├── logger/zap.go               # Zap logging
│   ├── middleware/                  # 7 middlewares
│   └── module/
│       ├── system/                 # System management
│       ├── payment/                # Payment module
│       ├── member/                 # Member module
│       └── captcha/                # Captcha
├── pkg/
│   ├── auth/jwt.go                 # JWT utilities
│   ├── upload/                     # Multi-provider file upload
│   ├── excel/excel.go              # Excel import/export
│   ├── task/cron.go                # Cron job scheduler
│   └── utils/                      # Utilities
├── router/router.go                # Route registration
├── sql/init.sql                    # Database initialization
├── web/                            # Frontend (Vue3)
├── AGENTS.md                       # AI Agent dev spec
└── README.md
```

## Architecture

```
Controller (Interface Layer)
  ├── Request binding (ShouldBindJSON / ShouldBindQuery)
  ├── Validation (binding tags)
  └── Unified Response

Service (Business Layer)
  ├── Business logic
  ├── Transaction control (db.Transaction)
  └── Call repository

Repository (Data Layer)
  ├── Database operations (GORM)
  └── Data access only, no business logic
```

### Middleware Order

```
Global: Recovery → Logger → Cors → Tenant
Auth:   Auth → CasbinAuth → OperationLog (protected routes only)
```

### Security Features

| Feature | Implementation |
|---------|----------------|
| Multi-tenant Isolation | `TenantScope` auto-filter, tenant_id from JWT only |
| Login Rate Limiting | Redis IP counter, 5 attempts / 5 min, lockout 15 min |
| Token Revocation | Auto-invalidate on password change / user disable |
| RBAC | Casbin policy control (skip when no policies) |
| CORS Whitelist | `cors.allow_origins` required in production |
| File Upload Validation | Extension whitelist + dangerous extension block |
| Payment Callback Verification | WeChat platform cert RSA + Alipay signature |
| Payment Amount Verification | Callback amount validation against order |
| Password Storage | bcrypt hashing |
| Open Redirect Protection | returnURL protocol and hostname validation |

## Features

### System Management

| Module | Description | API |
|--------|-------------|-----|
| User Management | CRUD, status, password reset | `/api/v1/system/user/*` |
| Role Management | CRUD, menu permission assignment | `/api/v1/system/role/*` |
| Menu Management | Directory / Menu / Button hierarchy | `/api/v1/system/menu/*` |
| Dept Management | Tree structure | `/api/v1/system/dept/*` |
| Post Management | CRUD | `/api/v1/system/post/*` |
| System Config | Key-value configuration | `/api/v1/system/config/*` |
| Data Dictionary | Dictionary types + data | `/api/v1/system/dict/*` |
| Log Management | Operation + login logs | `/api/v1/system/log/*` |
| File Management | Upload, preview, delete | `/api/v1/system/file/*` |

### Payment

| Module | Description | API |
|--------|-------------|-----|
| Create Order | WeChat JSAPI/Native, Alipay H5/Page | `POST /api/v1/system/pay/order` |
| Query Order | By order number | `GET /api/v1/system/pay/order` |
| Close Order | Close pending payment | `POST /api/v1/system/pay/order/close` |
| Refund | WeChat / Alipay refund | `POST /api/v1/system/pay/order/refund` |
| Callback | Async notification | `POST /api/v1/pay/notify/*` |

### Member Management

| Module | Description | API |
|--------|-------------|-----|
| Members | CRUD, status | `/api/v1/member/*` |
| Levels | Level configuration | `/api/v1/member/level/*` |
| Tags | Tag management | `/api/v1/member/tag/*` |
| Points Log | Points change history | `/api/v1/member/points/*` |

### Authentication

- JWT Login + Refresh Token (Access 2h / Refresh 7d)
- Casbin RBAC authorization
- User → Role → Menu/Button permissions
- Click captcha (brute-force protection)
- Login rate limiting (5 attempts / 5 min via Redis)
- Token revocation (immediate on password change / disable)

### File Upload

- Local storage (default)
- Alibaba Cloud OSS
- Tencent Cloud COS
- MinIO (S3 compatible)
- Auto-select via `sys_config` table `oss.*` keys on startup

## Quick Start

### Prerequisites

- Go 1.25+
- MySQL 5.7+
- Redis 3.0+
- Node.js 18+

### 1. Initialize Database

```bash
mysql -u root -p < sql/init.sql
```

### 2. Start Backend

```bash
# Edit config
vim config/config.yaml

# Build and run
go mod tidy
go build -o server.exe ./cmd/server
./server.exe
```

### 3. Start Frontend

```bash
cd web
npm install
npm run dev
```

### 4. One-click Start (Windows)

```powershell
.\start-all.ps1           # Start both
.\start-backend.ps1       # Backend only
.\start-frontend.ps1      # Frontend only
```

### 5. Access

| URL | Description |
|-----|-------------|
| http://localhost:3000 | Frontend |
| http://localhost:8080 | Backend API |
| http://localhost:8080/swagger/index.html | API Docs |

### Default Account

- Username: `admin`
- Password: `admin123`

## Configuration

### Environment Variables

Sensitive configs support env var override (higher priority than config.yaml):

```bash
export DB_PASSWORD="your_db_password"
export JWT_SECRET="your_jwt_secret"
export REDIS_PASSWORD="your_redis_password"
```

### CORS Config

```yaml
cors:
  allow_origins:
    - "http://localhost:3000"
    - "http://localhost:5173"
  allow_methods: ["GET","POST","PUT","DELETE","OPTIONS","PATCH"]
  allow_headers: ["Origin","Content-Type","Accept","Authorization","X-Tenant-Id"]
  allow_credentials: true
```

## Makefile

```bash
make build         # Build
make run           # Run
make test          # Test
make lint          # Static analysis (go vet)
make swagger       # Generate Swagger docs
make deps          # Tidy dependencies
make clean         # Clean build artifacts
```

## Testing

```bash
go test ./...
```

## Support

If this project helps you, consider buying me a coffee~

| WeChat | Alipay |
|:---:|:---:|
| <img src="docs/images/weixin.jpg" width="200"> | <img src="docs/images/zhifubao.jpg" width="200"> |

## License

[MIT License](LICENSE)
