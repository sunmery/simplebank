# Self-Documented Makefile see https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

# 生成sql代码
sqlc:
	sqlc generate

# 启动postgres容器
pg-up:
	docker compose -f ./deploy/database/docker/postgres-compose.yml up -d

# 停止postgres容器
pg-down:
	docker compose -f ./deploy/database/docker/postgres-compose.yml down

# 创建simple_bank数据库
pg-create-db:
	docker exec -it postgres17 createdb --username postgres --owner postgres simple_bank
	#docker exec -it postgres17 psql simple_bank --username postgres

# 删除simple_bank数据库
pg-drop-db:
	docker exec -it postgres17 dropdb simple_bank --username postgres

# 升级全部的迁移文件, 先安装https://github.com/golang-migrate/migrate/tree/master
migrate-up:
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

pg-init:
	make pg-down
	make pg-up
	sleep 1
	make pg-create-db
	make migrate-up

server:
	go run .

# go test
test:
	go test -v -cover ./...

# restart
pg-restart:
	make migrate-down && make migrate-up

# Mock DB
mock:
	mockgen -package mockdb -destination db/mock/store.go simple_bank/db/sqlc Store

.PHONY: sqlc pg-up pg-down pg-create-db pg-drop-db pg-init migrate-up migrate-up1 migrate-down migrate-down1 test server mock
