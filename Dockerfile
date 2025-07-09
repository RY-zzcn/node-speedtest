FROM golang:1.20-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache gcc musl-dev

# 复制go.mod和go.sum
COPY go.mod ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN mkdir -p bin && \
    cd panel && go build -o ../bin/panel main.go && \
    cd ../node && go build -o ../bin/node main.go

# 使用更小的基础镜像
FROM alpine:latest

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建工作目录
WORKDIR /app

# 从构建阶段复制编译好的应用
COPY --from=builder /app/bin/panel /app/bin/panel
COPY --from=builder /app/bin/node /app/bin/node
COPY --from=builder /app/panel/web /app/panel/web
COPY --from=builder /app/start.sh /app/start.sh

# 创建数据目录
RUN mkdir -p /app/data /app/panel/data /app/node/data && \
    chmod +x /app/start.sh

# 暴露端口
EXPOSE 8080 8081

# 设置入口点
ENTRYPOINT ["/app/start.sh"]

# 默认启动面板端
CMD ["panel"] 