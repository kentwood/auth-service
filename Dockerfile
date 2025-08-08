# Dockerfile
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 创建应用目录
WORKDIR /app

# 创建非root用户（安全考虑）
RUN addgroup -g 1000 appgroup && \
    adduser -D -s /bin/sh -u 1000 -G appgroup appuser

# 创建必要的目录
RUN mkdir -p /app/configs && \
    chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 设置默认的配置文件路径环境变量（可选）
ENV CONFIG_PATH=/app/configs/config.yaml

# 启动命令（假设您的二进制文件名为 auth-service）
CMD ["/app/auth-service"]