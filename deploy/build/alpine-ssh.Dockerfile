# 使用 alpine:latest 作为基础镜像
FROM alpine:latest

# 修改镜像源为中国科技大学的镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

# 更新安装包索引并安装所需软件包
RUN apk update
RUN apk add --no-cache tar sshpass openssh-client curl

# docker build -t lisa/alpine:git -f alpine-ssh.Dockerfile .
