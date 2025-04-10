# 第一阶段：构建阶段
FROM golang:latest AS builder

ENV CGO_ENABLED 0
ENV GOOS linux


# 设置工作目录
WORKDIR /app
# 复制源代码
COPY . .

RUN go mod tidy 
RUN go mod download

# 编译 Go 应用，禁用 CGO 以生成静态二进制文件
RUN go build -o h5 main.go

# 第二阶段：运行阶段
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/h5 .

# 暴露端口（假设应用监听 8080）
EXPOSE 8787

# 运行应用
CMD ["./main"]