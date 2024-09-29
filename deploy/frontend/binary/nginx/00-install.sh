#!/usr/bin/env bash
# 启用 POSIX 模式并设置严格的错误处理机制
set -o posix errexit -o pipefail

cd /tmp/ || exit 1
wget https://nginx.org/download/nginx-1.26.0.tar.gz
tar -zxvf nginx-1.26.0.tar.gz -C /usr/local/nginx
