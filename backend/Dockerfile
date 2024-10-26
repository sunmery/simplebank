# syntax=docker/dockerfile:1
# https://docs.docker.com/go/dockerfile-reference/

# 定义基础镜像的 Golang 版本
ARG GOIMAGE=golang:1.23.1-alpine3.20

FROM --platform=$BUILDPLATFORM ${GOIMAGE} AS build
WORKDIR /src

# 版本号
ARG VERSION=latest

# Go的环境变量, 例如alpine镜像不内置gcc,则关闭CGO很有效
ARG GOOS=linux
ARG GOARCH=amd64
ARG CGOENABLED=0

# Go的环境变量, 例如alpine镜像不内置gcc,则关闭CGO很有效
ARG GO_PROXY=https://proxy.golang.com.cn,direct

COPY . .

# 设置环境变量
# RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go env -w GOPROXY=$GO_PROXY

# 利用 Docker 层缓存机制，单独下载依赖项，提高后续构建速度。
# 使用缓存挂载和绑定挂载技术，避免不必要的文件复制到容器中。
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

# 获取代码版本号，用于编译时标记二进制文件
RUN echo "$a: $GOOS $GOARCH $CGOENABLED"
RUN VERSION=$(git describe --tags --always) || true
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    GOOS=$GOOS \
    GOARCH=$GOARCH \
    CGOENABLED=$CGOENABLED \
    go build -o /bin/main .
   # 带版本的形式: go build -ldflags="-X main.Version=${VERSION}" -o /bin/main .
   # 多个服务的形式: go build -o /bin/ ./...

FROM alpine:latest AS final
# 从构建阶段复制编译好的 Go 应用程序到运行阶段
COPY --from=build /bin/main /bin/

# 用户进程ID
ARG UID=10001

# 后端程序的HTTP/gRPC端口
ARG HTTP_PORT=30001
ARG GRPC_PORT=30002

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

# 复制配置文件
COPY app.env .

# 指定容器对外暴露的端口号
EXPOSE $HTTP_PORT
EXPOSE $GRPC_PORT

# 设置容器启动时执行的命令
CMD ["/bin/main"]

# 构建Docker所属的当前平台与架构的二进制文件, 进到当前的backend目录
# export VERSION=v2.0.0
# export REPOSITORY="lisa/backend"
# export GOOS=linux
# export GOARCH=amd64
# export HTTP_PORT=30001
# export GRPC_PORT=30002
# docker build . \
#   --progress=plain \
#   -t $REPOSITORY:$VERSION \
#   --build-arg CGOENABLED=0 \
#   --build-arg GOIMAGE=golang:1.23.1-alpine3.20 \
#   --build-arg GOOS=$GOOS \
#   --build-arg GOARCH=$GOARCH \
#   --build-arg VERSION=$VERSION \
#   --build-arg HTTP_PORT=$HTTP_PORT
#   --build-arg GRPC_PORT=$GRPC_PORT
#   --load

# 构建多架构的二进制文件, 需要在Docker Desktop 启用 containerd 映像存储
# https://docs.docker.com/desktop/containerd/#enable-the-containerd-image-store
# export VERSION=v2.0.0
# export REPOSITORY="lisa/backend"
# export GOOS=linux
# export GOARCH=amd64
# export HTTP_PORT=30001
# export GRPC_PORT=30002
# docker buildx build . \
#   --progress=plain \
#   -t $REPOSITORY:$VERSION \
#   --build-arg CGOENABLED=0 \
#   --build-arg GOIMAGE=golang:1.23.1-alpine3.20 \
#   --build-arg VERSION=$VERSION \
#   --build-arg HTTP_PORT=$PORT \
#   --build-arg GRPC_PORT=$PORT \
#   --build-arg GOOS=$GOOS \
#   --build-arg GOARCH=$GOARCH \
#   --platform $GOOS/$GOARCH

# 推送
# export register="ccr.ccs.tencentyun.com"
# docker tag $REPOSITORY:$VERSION $register/$REPOSITORY:$VERSION
# docker push $register/$REPOSITORY:$VERSION

# 拉取
# export GOOS=linux
# export GOARCH=amd64
# docker pull $register/$REPOSITORY:$VERSION --platform $GOOS/GOARCH

# 运行
# docker run \
# --rm \
# -p 30001:30001 \
# -p 30002:30002 \
# -e GIN_MODE=release \
# -e DB_SOURCE="postgresql://postgres:postgres@159.75.231.54:5432/simple_bank?sslmode=disable" \
# $register/container:$VERSION
