# Self-Documented Makefile see https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

# 生成sql代码
sqlc:
	sqlc generate

# 启动postgres容器
pg-up:
	make migrate-up
	docker compose -f manifests/database/docker/docker-compose.yaml up -d

# 停止postgres容器
pg-down:
	docker compose -f manifests/database/docker/docker-compose.yaml down

# 升级全部的迁移文件, 先安装https://github.com/golang-migrate/migrate/tree/master
migrate-up:
	#export DB_SOURCE="postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable"
	migrate -database "${DB_SOURCE}" -path db/migrate -verbose up

# 向上迁移一个版本, 根据数据库的表schema_migrations的version来决定
migrate-up1:
	migrate -database "${DB_SOURCE}" -path db/migrate -verbose up 1

# 向下全部降级迁移文件, 先安装https://github.com/golang-migrate/migrate/tree/master
migrate-down:
	migrate -database "${DB_SOURCE}" -path db/migrate -verbose down

# 向下降级一个版本, 根据数据库的表schema_migrations的version来决定
migrate-down1:
	migrate -database "${DB_SOURCE}" -path db/migrate -verbose down 1

# 启动redis
redis-up:
	docker-compose -f manifests/cache/redis/docker/compose.yaml up -d

# 关闭redis
redis-down:
	docker-compose -f manifests/cache/redis/docker/compose.yaml down

# 测试redis是否正常
redis-ping:
	docker exec -it redis7 redis-cli ping

# go test
test:
	go test -v -cover ./...

# Mock DB
mock:
	mockgen -package mockdb -destination db/mock/store.go simple_bank/db/sqlc Store

# 生成proto的go代码
proto:
	rm -rf pb/*.go
	protoc -I=proto \
	--go_out=pb --go_opt=paths=source_relative \
	--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
	proto/*.proto

# grpc的命令行测试工具
evans:
	evans -r repl --host localhost -p 30002

# API文档
swagger:
	rm -rf doc/swagger/*.go
	protoc -I=proto \
	--openapiv2_out=doc/swagger \
	--openapiv2_opt=allow_merge=true,merge_file_name=simple_bank \
	--openapiv2_opt=openapi_naming_strategy=fqn \
	proto/*.proto

#
swagger-build:
	docker buildx build . --progress=plain -t actions/swagger -f Dockerfile-swagger --platform linux/amd64
	# docker build . --progress=plain -t actions/swagger -f Dockerfile-swagger

swagger-run:
	docker run --rm \
    -p 8080:8080 \
    actions/swagger

swagger-push:
	docker tag actions/swagger ccr.ccs.tencentyun.com/lisa/swagger:v2.0.3
	docker push ccr.ccs.tencentyun.com/lisa/swagger:v2.0.3

server:
	make proto
	make mock
	make swagger
	go run .

.PHONY: sqlc pg-up pg-down migrate-up migrate-up1 migrate-down migrate-down1 test server mock proto evans swagger swagger-build swagger-run redis-up redis-down redis-ping
