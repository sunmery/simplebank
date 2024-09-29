# syntax=docker/dockerfile:1
# https://docs.docker.com/go/dockerfile-reference/

# 版本号
ARG VERSION=latest
# 定义基础镜像的 Golang 版本
ARG GO_IMAGE=golang:1.22.2-alpine3.19
# 构建的目标平台
ARG GOOS=linux
# 构建的目标架构
ARG ARCH=amd64
# Go的环境变量, 例如alpine镜像不内置gcc,则关闭CGO很有效
ARG CGO_ENABLED=0

FROM --platform=$BUILDPLATFORM ${GO_IMAGE} AS build
WORKDIR /src

# 设置环境变量
# RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go env -w GOPROXY=https://proxy.golang.com.cn,direct

# 利用 Docker 层缓存机制，单独下载依赖项，提高后续构建速度。
# 使用缓存挂载和绑定挂载技术，避免不必要的文件复制到容器中。
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

# 获取代码版本号，用于编译时标记二进制文件
#RUN VERSION=$(git describe --tags --always)
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    CGO_ENABLED=$CGO_ENABLED GOOS=$GOOS GOARCH=$ARCH \
    go build -ldflags="-X main.Version=${VERSION}" -o /bin/ ./...

FROM alpine:latest AS final
# 用户进程ID
ARG UID=10001
# 后端程序的gRPC端口(如果有), 例如30001
ARG PORT1=30001
# 后端程序的HTTP端口(如果有), 例如30002
ARG PORT1=30002

# 修改镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

# 安装应用运行必需的系统证书和时区数据包
# RUN --mount=type=cache,target=/var/cache/apk \
#    apk --update add ca-certificates tzdata && update-ca-certificates

# 创建一个非特权用户来运行应用，增强容器安全性
RUN adduser --disabled-password --gecos "" --home "/nonexistent" --shell "/sbin/nologin" --no-create-home --uid "${UID}" appuser

# 设置时区为上海
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

USER appuser

# 从构建阶段复制编译好的 Go 应用程序到运行阶段
COPY --from=build /bin/backend /bin/

# 指定容器对外暴露的端口号
EXPOSE $PORT1
EXPOSE $PORT2

# 设置容器启动时执行的命令
ENTRYPOINT ["/bin/backend", "-conf", "/data/conf"]

# 执行打包
# --progress=plain: 构建过程中显示的详细信息的格式
# --no-cache: 不使用缓存
# -t: 标签, 例如: lisa/frontend:v2
# frontend/ : 构建的目录, 相对于Dockerfile的路径, 与Docker相同的目录使用 . 表示当前目录
# -f frontend/Dockerfile: 相对路径, 指定Dockerfile的路径
# docker build --progress=plain --no-cache -t lisa/backend:v2 backend/ -f backend/Dockerfile

# 运行示例
# docker run -itd \
# --name backend \
# -v /home/backend/conf:/data/conf \
# -p '30001:30001/tcp' \
# -p '30002:30002/udp' \
# ccr.ccs.tencentyun.com/lisa/backend:v2
