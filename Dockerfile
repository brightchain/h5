# 第一阶段：构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 文件并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译 Go 应用，禁用 CGO 以生成静态二进制文件
RUN CGO_ENABLED=0 GOOS=linux go build -o h5 .

# 第二阶段：运行阶段
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/h5 .

# 暴露端口（假设应用监听 8080）
EXPOSE 8080

# 运行应用
CMD ["./h5"]