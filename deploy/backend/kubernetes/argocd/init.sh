#!/usr/bin/env bash

# 启用 POSIX 模式并设置严格的错误处理机制
set -o posix errexit -o pipefail

unset ARGOCD_NAMESPACE
unset ROLE_NAME
unset BRANCH
unset PROJECT_GIT_URL

# argocd所在的的命名空间
export ARGOCD_NAMESPACE="argo"
# 角色名称, 用于管理项目
export ROLE_NAME="admin"
# Kubernetes集群地址
export CLUSTER_SERVER="https://192.168.2.160:6443"
#export CLUSTER_SERVER="https://kubernetes.default.svc"

# 仓库URL
export BRANCH="main"
# 仓库地址
export PROJECT_GIT_URL="https://gitlab.com/lookeke/manifests.git"

# 后端端命名空间, 不需要额外创建命名空间选择default即可
export BACKEND_NAMESPACE="backend"
# argocd中的后端项目名, 用于分配团队人员的操作权限
export BACKEND_PROJECT_NAME="backend"
# 后端应用的名称
export BACKEND_APPLICATION_NAME="go"
# 后端的Kubernetes 资源清单在仓库中的路径, 相对于仓库根目录的路径
export BACKEND_DEPLOY_PATH="full-stack-engineering/backend"

# 获取Git Repo URL
if [ -z "$BRANCH" ]; then
  echo "用户未设置Git Repo URL, 尝试自动获取"
  # 该项目所在的git仓库地址
  export BRANCH=$(git rev-parse --abbrev-ref HEAD)
fi

if [ -z "$PROJECT_GIT_URL" ]; then
  if command -v git &> /dev/null
  then
      echo "Git is installed."
      export PROJECT_GIT_URL=$(git config --get remote."${BRANCH}".url)
      $PROJECT_GIT_URL
      echo "${PROJECT_GIT_URL}"
  else
      echo "Git is not installed."
      exit 1
  fi
fi

cat > create-backend-proj.yml <<EOF
apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: ${BACKEND_NAMESPACE} # 项目名称
  namespace: ${ARGOCD_NAMESPACE} # 项目所在的命名空间，默认为argocd，可根据实际情况调整
spec:
  # 项目描述
  description: "This project is for managing BACKEND applications"

  # 允许应用部署到的命名空间列表
  destinations:
  - namespace: ${BACKEND_NAMESPACE}
    server: ${CLUSTER_SERVER} # 集群API地址，示例值需替换为实际地址

  # 允许的目标 K8s 资源类型
  clusterResourceWhitelist:
    - group: '*'
      #kind: '*'
      kind: Namespace

  # 允许创建所有命名空间范围的资源: ResourceQuota、LimitRange、NetworkPolicy 除外
  namespaceResourceBlacklist:
    - group: '*'
      kind: ResourceQuota
    - group: '*'
      kind: LimitRange
    - group: '*'
      kind: NetworkPolicy
  # 拒绝创建所有名称空间作用域的资源. 但除了以下的Kind除外:
  namespaceResourceWhitelist:
    - group: '*'
      kind: Deployment
    - group: '*'
      kind: StatefulSet
    - group: '*'
      kind: Service
    - group: '*'
      kind: Namespace
  # 源代码仓库配置
  sourceRepos:
  - ${PROJECT_GIT_URL} # 允许使用的Git仓库地址，根据实际情况修改

  # 角色与成员
  roles:
  - name: ${ROLE_NAME} # 角色名称
    description: Access role for ROLE_NAME user
    # 注意这里的subjects配置应当符合Argo CD的RBAC规范，例如使用proj:backend:admin
    # 定义角色权限
    policies:
    # 允许 ROLE_NAME 角色对该命名空间下的项目进行: 获取/创建/同步/删除/操作
    - p, proj:${BACKEND_PROJECT_NAME}:${ROLE_NAME}, applications, get, ${BACKEND_PROJECT_NAME}/*, allow
    - p, proj:${BACKEND_PROJECT_NAME}:${ROLE_NAME}, applications, create, ${BACKEND_PROJECT_NAME}/*, allow
    - p, proj:${BACKEND_PROJECT_NAME}:${ROLE_NAME}, applications, sync, ${BACKEND_PROJECT_NAME}/*, allow
    - p, proj:${BACKEND_PROJECT_NAME}:${ROLE_NAME}, applications, delete, ${BACKEND_PROJECT_NAME}/*, allow
    # 允许查看集群信息
    - p, proj:${BACKEND_PROJECT_NAME}:${ROLE_NAME}, clusters, get, ${BACKEND_PROJECT_NAME}/*, allow
  orphanedResources:
    warn: true
# argocd proj create -f create-backend-proj.yml
# kubectl apply -f create-backend-proj.yml
EOF

cat > create-backend-app.yml <<EOF
apiVersion: argoproj.io/v1alpha1  # 指定 Argo CD API 版本
kind: Application  # 定义资源类型为 Application
metadata: # 元数据部分
  name: ${BACKEND_APPLICATION_NAME}   # 指定 Application 的名称
  namespace: ${ARGOCD_NAMESPACE}  # argocd所属的命名空间
  # 定义资源的 finalizers
  # https://argo-cd.readthedocs.io/en/stable/user-guide/app_deletion/#about-the-deletion-finalizer
  finalizers:
    - resources-finalizer.argocd.argoproj.io  # 删除时行级联删除
    #- resources-finalizer.argocd.argoproj.io/background  # 删除时后台行级联删除
spec: # 规范部分
  project: ${BACKEND_PROJECT_NAME}  # 应用程序将被配置的项目名称，这是在 Argo CD 中应用程序的一种组织方式
  source: # 指定源
    # Kubernetes 资源清单在仓库中的路径
    path: ${BACKEND_DEPLOY_PATH}
    # 指定 Git 仓库的 URL
    repoURL: ${PROJECT_GIT_URL}
    # 使用的 git 分支
    targetRevision: ${BRANCH}
  # 部署应用到Kubernetes 集群中的位置
  destination:
    namespace: ${BACKEND_NAMESPACE}  # 指定应用的命名空间
    server: ${CLUSTER_SERVER}  # 如果部署到同一集群，可以省略
  syncPolicy: # 指定同步策略
    automated: # 自动化同步
      prune: true  # 启用资源清理
      selfHeal: true  # 启用自愈功能
      allowEmpty: false  # 禁止空资源
    syncOptions: # 同步选项
      - Validate=false  # 是否启用验证
      - CreateNamespace=true  # 启用创建命名空间
    retry: # 重试策略
      limit: 5  # 重试次数上限
      backoff: # 重试间隔
        duration: 10s  # 初始重试间隔
        factor: 2  # 重试间隔因子
        maxDuration: 3m  # 最大重试间隔
EOF
