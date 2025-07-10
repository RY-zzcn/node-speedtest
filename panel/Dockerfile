FROM golang:1.21-alpine AS builder

WORKDIR /app

# 复制Go模块文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o panel ./panel/main.go

# 使用轻量级基础镜像
FROM alpine:latest

# 安装必要的依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 复制编译好的应用
COPY --from=builder /app/panel /app/panel

# 创建必要的目录
RUN mkdir -p /app/data /app/logs

# 设置时区
ENV TZ=Asia/Shanghai

# 暴露端口
EXPOSE 8080

# 设置入口点
ENTRYPOINT ["/app/panel"] 