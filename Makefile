.PHONY: all build clean panel node run-panel run-node test

# 默认目标
all: build

# 构建所有组件
build: panel node

# 构建面板端
panel:
	@echo "编译面板端..."
	@cd panel && go build -o ../bin/panel main.go

# 构建节点端
node:
	@echo "编译节点端..."
	@cd node && go build -o ../bin/node main.go

# 运行面板端
run-panel:
	@echo "运行面板端..."
	@cd panel && go run main.go

# 运行节点端
run-node:
	@echo "运行节点端..."
	@cd node && go run main.go

# 清理编译产物
clean:
	@echo "清理编译产物..."
	@rm -rf bin/
	@go clean

# 运行测试
test:
	@echo "运行测试..."
	@go test ./...

# 初始化项目目录
init:
	@echo "初始化项目目录..."
	@mkdir -p bin data panel/data node/data

# 帮助信息
help:
	@echo "使用说明:"
	@echo "  make build     - 构建面板端和节点端"
	@echo "  make panel     - 仅构建面板端"
	@echo "  make node      - 仅构建节点端"
	@echo "  make run-panel - 运行面板端"
	@echo "  make run-node  - 运行节点端"
	@echo "  make clean     - 清理编译产物"
	@echo "  make test      - 运行测试"
	@echo "  make init      - 初始化项目目录"
	@echo "  make help      - 显示帮助信息" 