# Self-Documented Makefile see https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

# 生成sql代码
sqlc:
	sqlc generate

# 启动postgres容器
postgres-up:
	docker compose -f postgres-compose.yml up -d

# 停止postgres容器
postgres-down:
	docker compose -f postgres-compose.yml down

# 创建simple_bank数据库
postgres-create-db:
	docker exec -it postgres15 createdb --username postgres --owner postgres simple_bank
	#docker exec -it postgres15 psql simple_bank --username postgres

# 删除simple_bank数据库
postgres-drop-db:
	docker exec -it postgres15 dropdb simple_bank --username postgres

# 升级全部的迁移文件, 先安装https://github.com/golang-migrate/migrate/tree/master
migrate-up:
	migrate -database "postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable" -path db/migrate -verbose up

# 向上迁移一个版本, 根据数据库的表schema_migrations的version来决定
migrate-up1:
	migrate -database "postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable" -path db/migrate -verbose up 1

# 向下全部降级迁移文件, 先安装https://github.com/golang-migrate/migrate/tree/master
migrate-down:
	migrate -database "postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable" -path db/migrate -verbose down

# 向下降级一个版本, 根据数据库的表schema_migrations的version来决定
migrate-down1:
	migrate -database "postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable" -path db/migrate -verbose down 1

# go test
test:
	go test -v -cover ./...

postgres-first:
	make postgres-up && make postgres-create-db && make migrate-up

# restart
postgres-restart:
	make migrate-down && make migrate-up

# Web Server Start
server:
	go run main.go

# Mock DB
mock:
	mockgen -package mockdb -destination db/mock/store.go simple_bank/db/sqlc Store

.PHONY: sqlc postgres-up postgres-down postgres-create-db postgres-drop-db migrate-up migrate-up1 migrate-down migrate-down1 test server mock
