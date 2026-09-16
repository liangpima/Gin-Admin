# syntax=docker/dockerfile:1

# ============================================================
# 后端镜像（多阶段构建）
#
# 构建： docker build -t gin-admin:latest .
# 运行： 见 docker-compose.yml（推荐）或 deploy/README.md
#
# 注意：本镜像**只包含后端**。Go 服务不提供前端页面，
#      前端由 web/Dockerfile 产出的 nginx 镜像承载，并反代 /api 与 /uploads。
# ============================================================

# ---------- 构建阶段 ----------
FROM golang:1.25-alpine AS builder

WORKDIR /src

# 先只拷贝依赖清单：依赖没变时这一层能命中缓存，避免每次改代码都重新拉依赖
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 产出静态二进制，才能跑在无 glibc 的 alpine 上。
# -trimpath 去掉构建机绝对路径；-s -w 去掉符号表与调试信息，显著减小体积。
RUN CGO_ENABLED=0 GOOS=linux go build \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/go-admin ./cmd/server

# 同时构建迁移执行器。放进镜像是为了让**容器部署场景**也能用它升级数据库 ——
# 否则升级只能退回「手工逐条 mysql < 文件」，漏执行不会有任何提示。
RUN CGO_ENABLED=0 GOOS=linux go build \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/migrate ./cmd/migrate

# ---------- 运行阶段 ----------
FROM alpine:3.20

# ca-certificates：调用支付宝/微信/OSS/COS 等 HTTPS 接口必需，缺了会报 x509 错误。
# tzdata：DSN 使用 loc=Local、日志用本地时间，没有时区库会全部退化成 UTC。
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 10001 app \
    && adduser -S -u 10001 -G app app

WORKDIR /app

COPY --from=builder /out/go-admin /app/go-admin
COPY --from=builder /out/migrate /app/migrate

# 配置与 casbin 模型随镜像分发（casbin 加载失败会**拒绝启动**，必须存在）。
# 运行时用 volume 覆盖 config/config.yaml 注入容器环境（见 deploy/config.docker.yaml）。
COPY config/ /app/config/

# sql/ 一并放入，便于在容器内执行初始化与迁移脚本
COPY sql/ /app/sql/

# 路径类配置（log.filename / upload.save_path / casbin.model_path）都是
# **相对工作目录**的（见 AGENTS.md「路径」一节），所以必须从 /app 启动。
# 这几个目录要预先建好并授权，否则写日志、上传文件会失败。
RUN mkdir -p /app/logs /app/uploads /app/runtime/certs \
    && chown -R app:app /app

USER app

EXPOSE 8080

# liveness：只判断进程与 HTTP 服务是否存活（不探测依赖，避免依赖抖动导致容器被反复重启）
HEALTHCHECK --interval=30s --timeout=3s --start-period=15s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/health || exit 1

# main.go 从 os.Args[1] 读取配置路径，显式传入便于用 volume 覆盖
ENTRYPOINT ["/app/go-admin"]
CMD ["config/config.yaml"]
