#!/usr/bin/env bash
# 启用 POSIX 模式并设置严格的错误处理机制
set -o posix errexit -o pipefail

# argocd所在的的命名空间
ARGOCD_NAMESPACE="argocd"
# 角色名称, 用于管理项目
ROLE_NAME="admin"

# 创建argocd的Project(项目)的Role(角色)
cat > deploy/frontend/kubernetes/argocd/project-role.yml <<EOF
#  创建角色
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-cm
  namespace: ${ARGOCD_NAMESPACE}
data:
  accounts.${ROLE_NAME}: "apiKey, login"
# kubectl apply -f argocd-cm.yaml -n ${ARGOCD_NAMESPACE}
EOF

# 给Role分配Project的权限
cat > deploy/frontend/kubernetes/argocd/project-rbac.yml <<EOF
# 分配角色给 frontend-group 前端组 和 backend-group 后端组
# 并具有适当的权限
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-rbac-cm
  namespace: ${ARGOCD_NAMESPACE}
data:
  policy.csv: |
    p, role:admin, applications, *, *, allow
    p, role:${ROLE_NAME}, applications, *, *, allow
    g, admin, role:admin
    g, ${ROLE_NAME}, proj:frontend:${ROLE_NAME}
    g, ${ROLE_NAME}, proj:backend:${ROLE_NAME}
# kubectl apply -f argocd-rbac-cm -n ${ARGOCD_NAMESPACE}
EOF

