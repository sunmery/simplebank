#!/usr/bin/env bash
# 启用 POSIX 模式并设置严格的错误处理机制
set -o posix errexit -o pipefail

# 判断 NGINX_DIR 变量是否已定义，如果未定义，则设置默认值
if [ -z "${NGINX_DIR}" ]; then
    export NGINX_DIR="/home/nginx"
    #exit 1
fi

# 判断 HTML_DIR 变量是否已定义，如果未定义，则设置默认值
if [ -z "${HTML_DIR}" ]; then
    export HTML_DIR="${NGINX_DIR}/html"
    #exit 1
fi

# 判断 CONF_DIR 变量是否已定义，如果未定义，则设置默认值
if [ -z "${CONF_DIR}" ]; then
    export CONF_DIR="${NGINX_DIR}/conf"
    #exit 1
fi

# 判断 LOG_DIR 变量是否已定义，如果未定义，则设置默认值
if [ -z "${LOG_DIR}" ]; then
    export LOG_DIR="${NGINX_DIR}/logs"
    #exit 1
fi

# 判断 SSL_DIR 变量是否已定义，如果未定义，则设置默认值
if [ -z "${SSL_DIR}" ]; then
    export SSL_DIR="${NGINX_DIR}/ssl"
    #exit 1
fi

# export DOMAIN="example.com"
# export NGINX_DIR=""
# export HTML_DIR=""
# export CONF_DIR=""
# export SSL_DIR=""

echo "用户定义的变量:"
echo "DOMAIN: ${DOMAIN}"
echo "NGINX_DIR: ${NGINX_DIR}"
echo "HTML_DIR: ${HTML_DIR}"
echo "CONF_DIR: ${CONF_DIR}"
echo "SSL_DIR: ${SSL_DIR}"

echo "创建一个系统用户 nginx，并将其 shell 设置为 /sbin/nologin，以确保该用户无法登录系统。"
sudo useradd -r -s /sbin/nologin nginx

echo "创建 /home/nginx 目录，并将其所有者设置为 root，组设置为 nginx，并给予 root 读、写和执行权限，而给予 nginx 组读和执行权限。"
sudo mkdir -p ${NGINX_DIR}
sudo chown -R root:nginx ${NGINX_DIR}
sudo chmod -R 750 ${NGINX_DIR}

echo "创建 /home/nginx/logs 目录，并将其所有者和组都设置为 nginx，并给予 nginx 用户和组读、写和执行权限。"
sudo mkdir -p ${LOG_DIR}
sudo chown -R nginx:nginx ${LOG_DIR}
sudo chmod -R 770 ${LOG_DIR}

echo "创建 /var/run/nginx.pid 文件，并将其所有者和组都设置为 nginx，并给予 nginx 用户和组读和写权限。"
sudo touch /var/run/nginx.pid
sudo chown nginx:nginx /var/run/nginx.pid
sudo chmod 660 /var/run/nginx.pid
