#!/usr/bin/env bash
# 启用 POSIX 模式并设置严格的错误处理机制
set -o posix errexit -o pipefail

NAME="frontend"
NAMESPACE="bank"
PORT1="80"
PORT2="443"
IMAGE="ccr.ccs.tencentyun.com/lisa/frontend:v0.6.0"
PORT_TYPE="NodePort"
DOMAIN="lookeke.com"

cat > deploy/frontend/kubernetes/argocd/secret.yml <<EOF
# 存储TLS证书和密钥
apiVersion: v1
kind: Secret
metadata:
  name: nginx-ssl
  namespace: frontend
type: kubernetes.io/tls
data:
  tls.crt: base64编码的证书数据
  tls.key: base64编码的密钥数据
# kubectl create secret tls domain --cert  tls.crt --key tls.key -n frontend
EOF

cat > deploy/frontend/kubernetes/argocd/pv.yml <<EOF
# 定义PV用于HTML存储
apiVersion: v1
kind: PersistentVolume
metadata:
  name: nfs-claim1
spec:
  capacity:
    storage: 100Mi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  # storageClassName: nfs-csi # 如果不使用默认的SC, 则需要手动编写, 根据你使用的SC不同, 这份PV清单参数仅供参考
  nfs:
    path: /mnt/data/full/kubernetes/ci/nginx/pv  # NFS共享的路径
    server: 192.168.2.160  # NFS服务器地址
---
EOF

cat > deploy/frontend/kubernetes/argocd/pvc.yml <<EOF
# 声明PVC用于HTML存储
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: html-volume-claim
  namespace: frontend
spec:
  accessModes:
    - ReadWriteOnce
  # storageClassName: nfs-csi # 如果不使用默认的SC, 则需要手动编写
  resources:
    requests:
      storage: 100Mi
---
EOF

cat > deploy/frontend/kubernetes/argocd/deploy.yml <<EOF
# 命名空间
apiVersion: v1
kind: Namespace
metadata:
  name: frontend
---
# 储Nginx配置
apiVersion: v1
kind: ConfigMap
metadata:
  name: nginx-conf
  namespace: frontend
data:
  nginx.conf: |
    server {
      listen 80;
      server_name ${DOMAIN}; # server_name
      return 301 https://${DOMAIN}; # webside
    }
    server {
      server_name ${DOMAIN} www.${DOMAIN};  # 服务器名称

      # UDP listener for QUIC+HTTP/3
      # http/3
      listen 443 quic reuseport;

      # http/2 and http/1.1
      listen 443 ssl;
      http2 on;

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
      ssl_prefer_server_ciphers on;  # 优先使用服务器的密码套件
      ssl_ecdh_curve X25519:P-256:P-384;  # 设置 ECDH 曲线
      ssl_early_data on;  # 启用 TLS 1.3 的 0-RTT 特性
      ssl_stapling on;  # 启用 OCSP Stapling
      ssl_stapling_verify on;  # 启用 OCSP Stapling 的验证

      # SSL 证书路径配置
      ssl_certificate     /etc/nginx/ssl/nginx.crt;  # SSL 证书路径
      ssl_certificate_key /etc/nginx/ssl/nginx.key;  # SSL 证书密钥路径

      location / {
      root   /etc/nginx/html;  # 设置根目录路径
      index  index.html index.htm default.html default.htm;  # 设置默认index首页文件
      }
    }
# kubectl create cm --from-file nginx.conf -n frontend
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: frontend
  labels:
    app: nginx
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
        - name: nginx-quic
          image: ${IMAGE}  # 替换成您的镜像
          ports:
            - containerPort: 80
            - containerPort: 443
          volumeMounts:
            - name: html-volume
              mountPath: /etc/nginx/html
            - name: ssl-volume
              mountPath: /etc/nginx/ssl
            - name: conf-volume
              mountPath: /etc/nginx/conf.d
      volumes:
        - name: html-volume
          persistentVolumeClaim:
            claimName: html-volume-claim
        - name: ssl-volume
          secret:
            secretName: nginx-ssl
        - name: conf-volume
          configMap:
            name: nginx-conf
---
apiVersion: v1
kind: Service
metadata:
  name: frontend-service
  namespace: frontend
spec:
  type: LoadBalancer
  ports:
    - port: 80
      targetPort: 80
      protocol: TCP
      name: http
    - port: 443
      targetPort: 443
      protocol: TCP
      name: https
  selector:
    app: nginx
EOF

cat > deploy/frontend/kubernetes/argocd/nginx.conf <<EOF
server {
    listen 80;
    server_name ${DOMAIN}; # server_name
    return 301 https://${DOMAIN}; # webside
}

server {
    server_name ${DOMAIN} www.${DOMAIN};  # 服务器名称

    # UDP listener for QUIC+HTTP/3
    # http/3
    listen 443 quic reuseport;

    # http/2 and http/1.1
    listen 443 ssl;
    http2 on;

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
    ssl_prefer_server_ciphers on;  # 优先使用服务器的密码套件
    ssl_ecdh_curve X25519:P-256:P-384;  # 设置 ECDH 曲线
    ssl_early_data on;  # 启用 TLS 1.3 的 0-RTT 特性
    ssl_stapling on;  # 启用 OCSP Stapling
    ssl_stapling_verify on;  # 启用 OCSP Stapling 的验证

    # SSL 证书路径配置
    ssl_certificate     /etc/nginx/ssl/nginx.crt;  # SSL 证书路径
    ssl_certificate_key /etc/nginx/ssl/nginx.key;  # SSL 证书密钥路径

    location / {
        root   /etc/nginx/html;  # 设置根目录路径
        index  index.html index.htm default.html default.htm;  # 设置默认index首页文件
    }
}
EOF
