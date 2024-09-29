#syntax=docker/dockerfile:1

ARG NODE_VERSION=node:20.11.0-alpine
ARG PACK_VERSION=latest
ARG FRONTEND_PORT1=80
ARG FRONTEND_PORT2=443
ARG BUILD_DIR_NAME=dist

FROM $NODE_VERSION AS builder
# FROM node:slim 仅包含运行 node 所需的最小包

WORKDIR /web

# 分离环境
RUN cd /web

# 默认使用生产环境
ENV NODE_ENV production

# 如果出现错误提示: 缺少使用 process.dlopen:
# Alpine v3.18及更早版本:
# RUN apk add --no-cache gcompat
# Alpine v3.19:
# apk add --no-cache libc6-compat

# 缓存.npm依赖
RUN --mount=type=cache,target=/root/.npm \
    npm install -g pnpm@$PACK_VERSION

COPY package.json .

# proxy
# RUN pnpm config set registry https://registry.npmmirror.com

# 下载依赖作为一个单独的步骤，以利用Docker的缓存。
# 利用缓存挂载到/root/.local/share/pnpm/store加速后续构建。
# 利用绑定挂载到 package.json 和 pnpm-lock.yaml 以避免将它们复制到这个layer。

# 安装依赖
# pnpm --prod --frozen-lockfile
# --prod 仅下载生产环境的依赖
# --frozen-lockfile 固定lock
# --mount=type=bind,source=pnpm-lock.yaml,target=pnpm-lock.yaml \
# --mount=type=bind,source=package.json,target=package.json \
RUN --mount=type=cache,target=/root/.local/share/pnpm/store \
    HUSKY=0 pnpm i

COPY . .

# 打包
RUN pnpm build

# 运行清理工作
#RUN rm -rf node_modules

# nginx QUIC + HTTP3 大小约为74.9MB
FROM ghcr.io/macbre/nginx-http3
# 使用nginx代理 web 项目的映射

# 把上一步骤打包好的的 dist目录传递到 nginx 默认的 html 目录作为映射
COPY --from=builder /web/dist/ /usr/share/nginx/html/

# 挂载 Nginx 的必要文件,
# /etc/nginx/ssl/是 SSL 配置,不需要删掉即可
# /var/log/nginx/ 是 nginx 日志,不需要删掉即可
# /var/cache/nginx/是 nginx 缓存, 不需要删掉即可
# /etc/nginx/conf/ 是 nginx 的配置文件目录, 必须保留
# /usr/share/nginx/html/是 nginx 的 html 目录, 必须保留
VOLUME ["/etc/nginx/html/","/etc/nginx/ssl/","/etc/nginx/conf.d/"]

EXPOSE ${FRONTEND_PORT1}
EXPOSE ${FRONTEND_PORT2}

# 运行 nginx 服务
CMD ["nginx", "-g", "daemon off;"]

# 执行打包
# --progress=plain: 构建过程中显示的详细信息的格式
# --no-cache: 不使用缓存
# -t: 标签, 例如: lisa/frontend:v2
# frontend/ : 构建的目录, 相对于Dockerfile的路径, 与Docker相同的目录使用 . 表示当前目录
# -f frontend/Dockerfile: 相对路径, 指定Dockerfile的路径
# docker build --progress=plain --no-cache -t lisa/frontend:v2 frontend/ -f frontend/Dockerfile

# 运行示例
# docker run -itd \
# --name nginx-quic \
# -v /home/nginx/html:/etc/nginx/html \
# -v /home/nginx/conf:/etc/nginx/conf.d \
# -v /home/nginx/ssl:/etc/nginx/ssl \
# -p '443:443/tcp' \
# -p '443:443/udp' \
# -p 80:80 \
# ccr.ccs.tencentyun.com/lisa/frontend:v2
