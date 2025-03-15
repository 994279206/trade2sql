# Trade2SQL Makefile
# 用于构建和打包Mac应用程序

# 变量定义
APP_NAME = Trade2SQL
VERSION = 1.0.0
BUILD_DIR = build
DIST_DIR = dist
RESOURCES_DIR = assets
MAIN_PKG = ./cmd/trade2sql

# Go命令
GO = go
GOBUILD = $(GO) build
GOCLEAN = $(GO) clean
GOTEST = $(GO) test
GOGET = $(GO) get

# Fyne命令
FYNE = fyne

# 默认目标
.PHONY: all
all: clean build package

# 清理构建目录
.PHONY: clean
clean:
	@echo "清理构建目录..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	@mkdir -p $(BUILD_DIR) $(DIST_DIR)
	@$(GOCLEAN)
	@echo "清理完成"

# 构建应用程序
.PHONY: build
build:
	@echo "构建应用程序..."
	@$(GOBUILD) -ldflags="-s -w -extldflags=-Wl,-ld_classic,-no_warn_duplicate_libraries" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PKG)

# 打包为Mac应用程序
.PHONY: package
package: clean build
	@echo "打包为Mac应用程序..."
	@$(FYNE) package -os darwin -icon $(RESOURCES_DIR)/icon.png -name $(APP_NAME) -appID trade2sql -exe $(BUILD_DIR)/$(APP_NAME) -release
	@mv $(APP_NAME).app $(DIST_DIR)/
	@echo "打包完成: $(DIST_DIR)/$(APP_NAME).app"

# 安装依赖
.PHONY: deps
deps:
	@echo "安装依赖..."
	@$(GOGET) fyne.io/fyne/v2/cmd/fyne@latest
	@$(GO) mod tidy

# 运行测试
.PHONY: test
test:
	@echo "运行测试..."
	@$(GOTEST) -v ./...

# 帮助信息
.PHONY: help
help:
	@echo "Trade2SQL Makefile 帮助"
	@echo "make             - 清理、构建并打包应用程序"
	@echo "make clean       - 清理构建目录"
	@echo "make build       - 构建应用程序"
	@echo "make package     - 打包为Mac应用程序"
	@echo "make deps        - 安装依赖"
	@echo "make test        - 运行测试"
	@echo "make help        - 显示帮助信息"