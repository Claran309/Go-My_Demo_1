# 如果您的 ClaranDemo 需要构建，这里提供一个示例 Dockerfile
FROM golang:1.20-alpine AS builder

WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o claran-demo .

# 运行阶段
FROM alpine:latest

RUN apk add --no-cache tzdata ca-certificates
WORKDIR /app

# 复制二进制文件
COPY --from=builder /app/claran-demo .

# 创建非root用户
RUN addgroup -g 1001 -S appuser && \
    adduser -u 1001 -S appuser -G appuser

USER appuser

EXPOSE 8080

CMD ["./claran-demo"]