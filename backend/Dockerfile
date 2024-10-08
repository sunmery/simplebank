# syntax=docker/dockerfile:1
# https://docs.docker.com/go/dockerfile-reference/

# 版本号
ARG VERSION=latest
# 定义基础镜像的 Golang 版本
ARG GO_IMAGE=golang:1.23.1-alpine3.20
# 构建的目标平台
ARG GOOS=linux
# 构建的目标架构
ARG ARCH=amd64
# Go的环境变量, 例如alpine镜像不内置gcc,则关闭CGO很有效
ARG CGO_ENABLED=0

FROM --platform=$BUILDPLATFORM ${GO_IMAGE} AS build
WORKDIR /src

COPY . .
RUN ls

RUN ls
RUN pwd
RUN ls /src

# 设置环境变量
# RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go env -w GOPROXY=https://proxy.golang.com.cn,direct

# 利用 Docker 层缓存机制，单独下载依赖项，提高后续构建速度。
# 使用缓存挂载和绑定挂载技术，避免不必要的文件复制到容器中。
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

# RUN go mod tidy

# 获取代码版本号，用于编译时标记二进制文件
RUN VERSION=$(git describe --tags --always) || true
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    CGO_ENABLED=$CGO_ENABLED GOOS=$GOOS GOARCH=$ARCH \
    go build -o /bin/main .
   # 带版本的形式: go build -ldflags="-X main.Version=${VERSION}" -o /bin/main .
   # 多个服务的形式: go build -o /bin/ ./...

FROM alpine:latest AS final

# 用户进程ID
ARG UID=10001
# 后端程序的gRPC端口(如果有), 例如30001
ARG PORT=8080

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
COPY --from=build /bin/main /bin/

# 复制配置文件
COPY app.env .

# 指定容器对外暴露的端口号
EXPOSE $PORT

# 设置容器启动时执行的命令
CMD ["/bin/main"]

# 打包, 进到当前的backend目录
# export version=v2.0.0
# export repository="lisa/backend"
# docker build \
#   --progress=plain \
#   -t $repository:$version . \
#   --build-arg GOOS=darwin \
#   --build-arg ARCH=arm64 \
#   --build-arg CGO_ENABLED=0 \
#   --build-arg GO_IMAGE=golang:1.23.1-alpine3.20 \
#   --build-arg version=$version \
#   --build-arg PORT=8080

# 推送
# export register="ccr.ccs.tencentyun.com"
# docker tag $repository:$version $register/$repository:$version
# docker push $register/$repository:$version

# 拉取
# docker pull $register/$repository:$version

# 运行
# docker run \
# --rm \
# -p 8080:8080 \
# $register/container:$version
