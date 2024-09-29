#!/usr/bin/env bash
# 启用 POSIX 模式并设置严格的错误处理机制
set -o posix errexit -o pipefail

NAME="backend"
NAMESPACE="backend"
SERVER_REPLICAS="1"
SERVER_PORT_TYPE="LoadBalancer"

SERVER_IMAGE1="ccr.ccs.tencentyun.com/lisa/backend:v0.2.7"
SERVER_PORT1="30001"

SERVER_IMAGE2="ccr.ccs.tencentyun.com/lisa/backend:v0.2.7"
SERVER_PORT2="30002"

cat > deploy.yml <<EOF
# 命名空间
apiVersion: v1
kind: Namespace
metadata:
  name: ${NAMESPACE}
---
# 部署清单
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${NAME}
  namespace: ${NAMESPACE}
spec:
  replicas: ${SERVER_REPLICAS}
  selector:
    matchLabels:
      app: ${NAME}
  template:
    metadata:
      labels:
        app: ${NAME}
    spec:
      containers:
        - name: http
          image: ${SERVER_IMAGE1}
          ports:
            - containerPort: ${SERVER_PORT1}
        - name: grpc
          image: ${SERVER_IMAGE2}
          ports:
            - containerPort: ${SERVER_PORT2}
          volumeMounts:
            - name: kratos-config
              mountPath: /data/conf
      volumes:
        - name: kratos-config
          configMap:
            # 提供包含要添加到容器中的文件的 ConfigMap 的名称
            name: kratos-config
      restartPolicy: Always
---
# 服务清单
apiVersion: v1
kind: Service
metadata:
  name: ${NAME}-service
  namespace: ${NAMESPACE}
spec:
  selector:
    app: ${NAME}
  ports:
    - name: http
      protocol: TCP
      port: ${SERVER_PORT1}
      targetPort: ${SERVER_PORT1}
    - name: grpc
      protocol: TCP
      port: ${SERVER_PORT2}
      targetPort: ${SERVER_PORT2}
  type: ${SERVER_PORT_TYPE}
EOF
