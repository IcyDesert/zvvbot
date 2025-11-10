# --- Build Stage ---
FROM golang:1.25.3-alpine AS builder

WORKDIR /app

ARG http_proxy
ARG https_proxy

# 预先复制 go.mod 和 go.sum 并下载依赖项，以利用 Docker 的层缓存
COPY go.mod go.sum ./
RUN echo "Using proxy: $https_proxy" && go mod download

# 复制所有源代码
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o zvvbot .

# --- Final Stage ---
FROM alpine:latest

WORKDIR /app

# 从构建阶段复制编译好的二进制文件和配置文件
COPY --from=builder /app/zvvbot .
COPY --from=builder /app/config.yml .

# 暴露应用监听的端口
EXPOSE 6432

# 设置容器启动时执行的命令
CMD ["./zvvbot"]