# Self-Documented Makefile see https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

# 生成sql代码
deploy-frontend:
	./deploy/frontend/kubernetes/argocd/init.sh
	chmod +x deploy/frontend/kubernetes/argocd/role.sh
	./deploy/frontend/kubernetes/argocd/role.sh
	./deploy/frontend/kubernetes/argocd/deploy.sh

.PHONY: deploy-frontend
