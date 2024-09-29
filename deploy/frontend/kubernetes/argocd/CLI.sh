#!/usr/bin/env bash
# 启用 POSIX 模式并设置严格的错误处理机制
set -o posix errexit -o pipefail

# https://github.com/argoproj-labs/argocd-operator/releases/tag
VERSION="v0.12.0"
wget https://github.com/argoproj-labs/argocd-operator/archive/refs/tags/${VERSION}.zip

apt install -y unzip
unzip ${VERSION}.zip

cd argocd-operator-v${VERSION} || exit
# 安装操作员, 此 Operator 将安装在 “operators” 命名空间中，并可从集群中的所有命名空间中使用
kubectl create -f https://operatorhub.io/install/argocd-operator.yaml
kubectl get csv -n operators

# 安装argocd实例
export ns="argocd"
kubectl create ns $ns
kubectl create -f argocd-deploy.yaml -n $ns
# 如果需要再argocd-cm添加任何参数, 请编辑argocd.deploy.yaml的spec.extraConfig, 例如
# spec:
#   extraConfig:
#     accounts.admin: "apiKey, login"

# 获取密码
echo "将default-argocd替换成你的argocd的名称"
pwd=$(kubectl -n $ns get secret default-argocd-cluster -o jsonpath='{.data.admin\.password}' | base64 -d)
# tVEO1vRShlfkWysUCuKTw432oXzmB7ZN

# CLI登录
# $lb_ip:port: ip与端口
# --insecure: 忽略TLS验证
# --grpc-web
lb_ip=$(kubectl get service example-argocd-server -o=jsonpath='{.status.loadBalancer.ingress[0].ip}' -n $ns)
argocd login \
$lb_ip \
--username admin \
--password $pwd \
--insecure

# 修改密码
argocd account update-password

# 列出当前集群上下文
kubectl config get-contexts -o name

# 添加集群权限级别的RBAC,即可在任意ns创建/删除应用(需要注意安全性), 示例:
argocd cluster add kubernetes-admin@kubernetes

PROJECT="simplebank"
CLUSTER_URL="https://116.213.43.175:6443"
REPO_URL="https://github.com/sunmery/simplebank.git"
APPLICATION_NAME="bank-frontend1"
NAMESPACE="ban1"

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

# 查看已添加的git仓库
argocd repo list

# 通常 GitOps 使用的仓库是私有仓库，所以添加仓库时一般用 --ssh-private-key-path 指定下 SSH 密钥，
# 以便让 argocd 能够正常拉取到 Git 仓库。
#argocd repo add --ssh-private-key-path $HOME/.ssh/id_rsa --insecure-skip-server-verification git@yourgit.com:your-org/your-repo.git


# 创建APP
argocd app create ${APPLICATION_NAME} \
--project ${PROJECT} \
--repo ${REPO_URL} \
--path deploy/frontend/kubernetes/argocd \
--dest-server ${CLUSTER_URL} \
--dest-namespace ${NAMESPACE} \
--validate

# 查询该应用信息
argocd app get ${APPLICATION_NAME}

# 设置该应用为某个proj
argocd app set ${APPLICATION_NAME} --project ${PROJECT}

# 删除
#argocd app delete ${PROJECT}

# 列出用户
argocd account list

# 获取特定用户信息
argocd account get --account <username>

# 生成token
argocd account generate-token --account admin
# token:
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJhcmdvY2QiLCJzdWIiOiJhZG1pbjphcGlLZXkiLCJuYmYiOjE3MTQ5Mjg0MTYsImlhdCI6MTcxNDkyODQxNiwianRpIjoiYzJiNTAzYzAtNmI0Mi00MzljLTliYTQtNjk1M2E5ZjU5OGZiIn0.t1AjKKWYNBshV5oGFYXOQCfWX-S_u2hX3NcHS3WPMrM

# RBAC权限:
# p, role:lx, applications, *, */*, allow
# p, role:lx, clusters, *, *, allow
# p, role:lx, repositories, *, */*, allow
# p, role:lx, projects, *, */*, allow
# p, role:lx, projects, sync, */*, allow
# p, role:lx, logs, *, */*, allow
# p, role:lx, exec, *, */*, allow
# p, role:admin, applications, *, */*, allow
# p, role:admin, clusters, *, *, allow
# p, role:admin, repositories, *, */*, allow
# p, role:admin, projects, sync, */*, allow
# p, role:admin, logs, *, */*, allow
# p, role:admin, exec, *, */*, allow
# g, admin, role:admin
# g, admin, role:lx
# policy.default: role:admin

# 验证RBAC权限:
# 验证包含rbac的yml或csv文件
argocd admin settings rbac validate --policy-file argocd-rbac-cm.yml
# 命名空间:
argocd admin settings rbac validate --namespace argocd

# 测试策略
# https://argo-cd.readthedocs.io/en/stable/operator-manual/rbac/#testing-a-policy
argocd admin settings rbac can role:org-admin get applications --policy-file argocd-rbac-cm.yaml

