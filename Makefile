.PHONY: help build run test clean deps migrate dev

# 默认目标
help:
	@echo "可用命令:"
	@echo "  make deps     - 安装依赖"
	@echo "  make build    - 编译项目"
	@echo "  make run      - 运行项目"
	@echo "  make dev      - 开发模式运行"
	@echo "  make test     - 运行测试"
	@echo "  make clean    - 清理编译文件"
	@echo "  make migrate  - 运行数据库迁移"

# 安装依赖
deps:
	go mod download
	go mod tidy

# 编译项目
build:
	go build -o bin/server cmd/server/main.go

# 运行项目
run: build
	./bin/server

# 开发模式运行（使用go run）
dev:
	GIN_MODE=debug go run cmd/server/main.go

# 运行测试
test:
	go test -v ./...

# 清理编译文件
clean:
	rm -rf bin/
	rm -f logs/*.log

# 数据库迁移（需要安装migrate工具）
migrate:
	@echo "请手动运行数据库迁移或使用GORM的AutoMigrate功能"

