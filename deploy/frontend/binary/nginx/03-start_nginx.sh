#!/usr/bin/env bash
# 启用 POSIX 模式并设置严格的错误处理机制
set -o posix errexit -o pipefail

if [ -z "${DOMAIN}" ]; then
    echo "请编写你的域名"
    exit 1
fi

cp /usr/local/nginx/conf/nginx.conf{,.back}

# 判断 NGINX_DIR 变量是否已定义，如果未定义，则设置默认值
if [ -z "${NGINX_DIR}" ]; then
    export NGINX_DIR="/home/nginx"
fi

# 判断 HTML_DIR 变量是否已定义，如果未定义，则设置默认值
if [ -z "${HTML_DIR}" ]; then
    export HTML_DIR="${NGINX_DIR}/html"
fi

# 判断 CONF_DIR 变量是否已定义，如果未定义，则设置默认值
if [ -z "${CONF_DIR}" ]; then
    export CONF_DIR="${NGINX_DIR}/conf"
fi

# 判断 LOG_DIR 变量是否已定义，如果未定义，则设置默认值
if [ -z "${LOG_DIR}" ]; then
    export LOG_DIR="${NGINX_DIR}/logs"
fi

# 判断 SSL_DIR 变量是否已定义，如果未定义，则设置默认值
if [ -z "${SSL_DIR}" ]; then
    export SSL_DIR="${NGINX_DIR}/ssl"
fi

echo "用户定义的变量:"
echo "DOMAIN: ${DOMAIN}"
echo "NGINX_DIR: ${NGINX_DIR}"
echo "HTML_DIR: ${HTML_DIR}"
echo "CONF_DIR: ${CONF_DIR}"
echo "SSL_DIR: ${SSL_DIR}"
echo "LOG_DIR: ${LOG_DIR}"

cp /usr/local/nginx/conf/nginx.conf{,.back}
rm -rf /usr/local/nginx/conf/nginx.conf
cat > /usr/local/nginx/conf/nginx.conf <<EOF
user  nginx;  # 定义运行 Nginx 工作进程的用户
worker_processes  1;  # 设置工作进程的数量

error_log  ${LOG_DIR}/error.log warn;  # 错误日志文件的路径，日志级别设置为 'warn'

pid        /run/nginx.pid;  # 存储主进程 ID 的文件

events {
    worker_connections  1024;  # 默认工作连接数设置为 1024
    #worker_connections  65535;  # 设置每个工作进程的最大连接数，这里为 65535
}

http {
    # MIME 类型文件的包含指令
    include       mime.types;
    # 默认 MIME 类型设置
    default_type  application/octet-stream;

    #log_format  main  ...;  # 默认日志格式定义

    # 自定义的日志格式，包括 QUIC 相关的变量
    log_format quic '$remote_addr - $remote_user [$time_local] '
                    '"$request" $status $body_bytes_sent '
                    '"$http_referer" "$http_user_agent" "$http3"';

    access_log  ${LOG_DIR}/access.log  quic;  # 访问日志的路径和日志格式
    # 启用或禁用 sendfile 模式
    sendfile on;
    # 启用或禁用 gzip 压缩
    gzip  on;
    # 保持连接的超时时间设置
    keepalive_timeout  65;

    include ${CONF_DIR}/*.conf;  # 额外配置文件包含指令
}
EOF

cat > "${CONF_DIR}"/nginx.conf <<EOF
server {
   listen 80;
   # 主机IP或者域名, 多个用空格区分
   server_name ${DOMAIN};
   return 301 https://${DOMAIN};
}

server {
   server_name ${DOMAIN} www.${DOMAIN};  # 服务器名称

   # UDP listener for QUIC+HTTP/3
   #listen 443 quic reuseport so_keepalive=on;  # 为 QUIC+HTTP/3 设置 UDP 监听器
   listen 443 ssl reuseport default_server so_keepalive=on;  # 为 HTTPS 设置监听器

   http2 on;  # 启用 HTTP/2 协议

   # 以下为各种 HTTP 安全相关头部的设置
   add_header Strict-Transport-Security "max-age=63072000; includeSubdomains; preload";
   add_header X-XSS-Protection "1; mode=block" always;
   add_header X-Frame-Options SAMEORIGIN always;
   add_header X-Content-Type-Options nosniff;
   add_header X-Frame-Options "DENY";
   add_header Alt-Svc 'h3=":443"; ma=86400, h3-29=":443"; ma=86400';

   # SSL/TLS 相关配置
   ssl_protocols TLSv1.3 TLSv1.2;  # 设置支持的 SSL 协议版本
   # ssl_ciphers ...;  # 设置 SSL 密码套件
   # 提高 SSL/TLS 连接的性能而设计的一种机制，它允许服务器在 SSL 握手过程中重用之前建立的 SSL 会话信息，从而避免了重复的密钥交换等操作，提高了连接的响应速度。
   ssl_session_cache    shared:SSL:30m;
   ssl_session_timeout  30m;
   ssl_prefer_server_ciphers on;  # 优先使用服务器的密码套件
   ssl_ecdh_curve X25519:P-256:P-384;  # 设置 ECDH 曲线
   ssl_early_data on;  # 启用 TLS 1.3 的 0-RTT 特性
   ssl_stapling on;  # 启用 OCSP Stapling
   ssl_stapling_verify on;  # 启用 OCSP Stapling 的验证
   proxy_set_header Early-Data '$ssl_early_data';  # 设置 Early-Data 头以防止重放攻击

   # SSL 证书路径配置
   ssl_certificate     /home/nginx/ssl/nginx.crt;  # SSL 证书路径
   ssl_certificate_key /home/nginx/ssl/nginx.key;  # SSL 证书密钥路径

   # 将服务器错误页面重定向到静态页面/50x.html
   error_page   500 502 503 504  /50x.html;
   # 错误页面路径
   location = /50x.html {
               root   html;
   }

   location / {
       root   /home/nginx/html;  # 设置根目录路径
       index  index.html index.htm;  # 设置默认index首页文件

       # 添加 HTTP/3 相关的头部
       add_header QUIC-Status '$http3';
       add_header Alt-Svc 'h3=":443"; ma=86400'; # used to advertise the availability of HTTP/3
       #add_header Alt-Svc 'h3-27=":443"; h3-28=":443"; h3-29=":443"; ma=86400; quic=":443"';
   }
}
EOF

echo "清理可能的意外, 保证容器被本shell script控制"
# 查询端口 80 和 443 的进程，并获取它们的 PID
pid_80=$(sudo lsof -t -i:80)
pid_443=$(sudo lsof -t -i:443)

# 如果找到对应的进程，就杀死它们
if [ -n "$pid_80" ]; then
    sudo kill -9 $pid_80
    echo "Killed process on port 80 (PID: $pid_80)"
fi

if [ -n "$pid_443" ]; then
    sudo kill -9 $pid_443
    echo "Killed process on port 443 (PID: $pid_443)"
fi

# systemctl方式:
systemctl start nginx.service 启动
# systemctl stop nginx.service 停止
# systemctl status nginx.service 查看状态
# systemctl restart nginx.service 重新启动
# systemctl reload nginx.service 重新读取配置

# 二进制方式:
# /usr/local/nginx/sbin/nginx
# /usr/local/nginx/sbin/nginx -s reload
