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
| ![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat&logo=go&logoColor=white) | 1.25 | Programming Language |
| ![Gin](https://img.shields.io/badge/Gin-v1.10.0-008ECF?style=flat&logo=gin&logoColor=white) | v1.10.0 | Web Framework |
| ![GORM](https://img.shields.io/badge/GORM-v1.25.12-0092D2?style=flat&logo=database&logoColor=white) | v1.25.12 | ORM Framework |
| ![MySQL](https://img.shields.io/badge/MySQL-5.7%2B-4479A1?style=flat&logo=mysql&logoColor=white) | 5.7+ | Relational Database |
| ![Redis](https://img.shields.io/badge/Redis-3.0%2B-DC382D?style=flat&logo=redis&logoColor=white) | 3.0+ | Cache / Token Storage |
| ![JWT](https://img.shields.io/badge/JWT-v5.2.3-000000?style=flat&logo=json-web-tokens&logoColor=white) | v5.2.3 | Authentication |
| ![Casbin](https://img.shields.io/badge/Casbin-v2.103.0-3D9B8F?style=flat&logo=casbin&logoColor=white) | v2.103.0 | RBAC Authorization |
| ![Viper](https://img.shields.io/badge/Viper-v1.19.0-BD3FEB?style=flat&logo=viper.js&logoColor=white) | v1.19.0 | Configuration |
| ![Zap](https://img.shields.io/badge/Zap-v1.27.0-ECBA52?style=flat&logo=go&logoColor=white) | v1.27.0 | Structured Logging |
| ![Swagger](https://img.shields.io/badge/Swagger-v1.6.0-85EA2D?style=flat&logo=swagger&logoColor=white) | v1.6.0 | API Documentation |
| ![OSS](https://img.shields.io/badge/Aliyun_OSS-v3.0.2-FF6A00?style=flat&logo=alibabacloud&logoColor=white) | v3.0.2 | Alibaba Cloud OSS |
| ![COS](https://img.shields.io/badge/Tencent_COS-v0.7.74-006EFF?style=flat&logo=tencentcloud&logoColor=white) | v0.7.74 | Tencent Cloud COS |
| ![MinIO](https://img.shields.io/badge/MinIO-v7.2.1-C72C48?style=flat&logo=minio&logoColor=white) | v7.2.1 | S3 Compatible Storage |

### Frontend

| Technology | Version | Description |
|------------|---------|-------------|
| ![Vue](https://img.shields.io/badge/Vue-3.5.13-42b883?style=flat&logo=vuedotjs&logoColor=white) | 3.5.13 | Progressive Framework |
| ![Vue Router](https://img.shields.io/badge/Vue_Router-4.5.0-42b883?style=flat&logo=vuedotjs&logoColor=white) | 4.5.0 | Routing |
| ![Pinia](https://img.shields.io/badge/Pinia-2.3.0-FCCD2B?style=flat&logo=pinia&logoColor=white) | 2.3.0 | State Management |
| ![Element Plus](https://img.shields.io/badge/Element_Plus-2.9.1-409EFF?style=flat&logo=element&logoColor=white) | 2.9.1 | UI Component Library |
| ![Axios](https://img.shields.io/badge/Axios-1.7.9-5A29E4?style=flat&logo=axios&logoColor=white) | 1.7.9 | HTTP Client |
| ![Vite](https://img.shields.io/badge/Vite-6.0.5-646CFF?style=flat&logo=vite&logoColor=white) | 6.0.5 | Build Tool |
| ![TypeScript](https://img.shields.io/badge/TypeScript-5.7.2-3178C6?style=flat&logo=typescript&logoColor=white) | 5.7.2 | Type System |

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
- Supported formats: images (jpg/png/gif/bmp/svg/webp), video (mp4/mov/avi), audio (mp3/wav), documents (pdf/doc/xls/ppt), archives (zip/rar/7z)

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
