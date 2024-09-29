#!/usr/bin/env bash
# 启用 POSIX 模式并设置严格的错误处理机制
set -o posix errexit -o pipefail

PROJECT="test1"
CLUSTER_URL="https://116.213.43.175:6443"
REPO_URL="https://github.com/sunmery/simplebank.git"
APPLICATION_NAME="bank-frontend12"
NAMESPACE="test1"
M_PATH="deploy/frontend/kubernetes/argocd/test"

# 创建项目
argocd proj create ${PROJECT}

# 删除
# argocd proj delete frontend

# 添加仓库到项目
#  argocd proj add-source <PROJECT> <REPO>
argocd proj add-source ${PROJECT} ${REPO_URL}

# 删除
# argocd proj remove-source <PROJECT> <REPO>

# 排除项目
# argocd proj add-source <PROJECT> !<REPO>

# 添加集群与命名空间
# argocd proj add-destination <PROJECT> <CLUSTER>,<NAMESPACE>
# argocd proj remove-destination <PROJECT> <CLUSTER>,<NAMESPACE>
argocd proj add-destination ${PROJECT} ${CLUSTER_URL} ${PROJECT}

# 创建仓库秘钥
#cat > gitlab-secret.yml <<EOF
#apiVersion: v1
#kind: Secret
#metadata:
#  name: argocd-example-apps
#  labels:
#    argocd.argoproj.io/secret-type: repository
#type: Opaque
#stringData:
#  # Project scoped
#  project: my-project1
#  name: argocd-example-apps
#  url: https://github.com/argoproj/argocd-example-apps.git
#  username: ****
#  password: ****
#EOF

# 创建APP

argocd app create ${APPLICATION_NAME} \
--project ${PROJECT} \
--repo ${REPO_URL} \
--path ${M_PATH} \
--dest-server ${CLUSTER_URL} \
--dest-namespace ${NAMESPACE} \
--validate

# 查询该应用信息
argocd app get ${APPLICATION_NAME}

# 设置该应用为某个proj
#argocd app set ${APPLICATION_NAME} --project ${PROJECT}

# 删除
#argocd app delete ${PROJECT}

# 列出用户
#argocd account list
