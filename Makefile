# GoExecutors Makefile

.PHONY: build clean test test-verbose test-race test-coverage benchmark lint fmt vet tidy example help

# 变量定义
APP_NAME := goexecutors
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S')
GO_VERSION := $(shell go version | sed 's/go version //')

# 默认目标
.DEFAULT_GOAL := help

# 构建
build: ## 构建项目
	@echo "Building $(APP_NAME)..."
	@go build -ldflags "-X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)' -X 'main.GoVersion=$(GO_VERSION)'" -o bin/$(APP_NAME) .
	@echo "Build completed: bin/$(APP_NAME)"

# 清理
clean: ## 清理构建文件
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf coverage.out
	@rm -rf coverage.html
	@echo "Clean completed"

# 运行测试
test: ## 运行测试
	@echo "Running tests..."
	@go test -v ./...

# 详细测试
test-verbose: ## 运行详细测试
	@echo "Running verbose tests..."
	@go test -v -race ./...

# 竞态条件测试
test-race: ## 运行竞态条件测试
	@echo "Running race tests..."
	@go test -race ./...

# 覆盖率测试
test-coverage: ## 运行覆盖率测试
	@echo "Running coverage tests..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 基准测试
benchmark: ## 运行基准测试
	@echo "Running benchmarks..."
	@go test -v -bench=. -benchmem ./...

# Lint 检查
lint: ## 运行 lint 检查
	@echo "Running lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# 格式化代码
fmt: ## 格式化代码
	@echo "Formatting code..."
	@go fmt ./...

# 静态检查
vet: ## 运行 go vet
	@echo "Running go vet..."
	@go vet ./...

# 整理依赖
tidy: ## 整理 go.mod
	@echo "Tidying dependencies..."
	@go mod tidy

# 运行示例
example: ## 运行示例
	@echo "Running example..."
	@go run main.go

# 运行完整示例
example-full: ## 运行完整示例
	@echo "Running full example..."
	@go run examples/main.go

# 安装依赖
deps: ## 安装依赖
	@echo "Installing dependencies..."
	@go mod download

# 更新依赖
update-deps: ## 更新依赖
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

# 生成文档
docs: ## 生成文档
	@echo "Generating documentation..."
	@if command -v godoc >/dev/null 2>&1; then \
		echo "Documentation server starting at http://localhost:6060"; \
		godoc -http=:6060; \
	else \
		echo "godoc not found, install it with: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

# 检查代码质量
quality: fmt vet lint test-race ## 运行代码质量检查

# 完整的 CI 流程
ci: quality test-coverage benchmark ## 运行 CI 流程

# 发布检查
release-check: ## 发布前检查
	@echo "Running release checks..."
	@go mod verify
	@go mod tidy
	@git diff --exit-code go.mod go.sum
	@$(MAKE) ci

# 创建发布
release: release-check ## 创建发布
	@echo "Creating release..."
	@if [ -z "$(TAG)" ]; then \
		echo "Usage: make release TAG=v1.0.0"; \
		exit 1; \
	fi
	@git tag -a $(TAG) -m "Release $(TAG)"
	@git push origin $(TAG)
	@echo "Release $(TAG) created"

# 环境信息
env: ## 显示环境信息
	@echo "Environment Information:"
	@echo "  Go Version: $(GO_VERSION)"
	@echo "  Version: $(VERSION)"
	@echo "  Build Time: $(BUILD_TIME)"
	@echo "  GOOS: $(shell go env GOOS)"
	@echo "  GOARCH: $(shell go env GOARCH)"

# 性能分析
profile: ## 运行性能分析
	@echo "Running performance profile..."
	@go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./executors/
	@echo "Profile files generated: cpu.prof, mem.prof"
	@echo "View CPU profile: go tool pprof cpu.prof"
	@echo "View Memory profile: go tool pprof mem.prof"

# 安装工具
install-tools: ## 安装开发工具
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/godoc@latest
	@go install github.com/go-delve/delve/cmd/dlv@latest
	@echo "Tools installed"

# 帮助
help: ## 显示帮助信息
	@echo "GoExecutors Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 监控文件变化并运行测试
watch: ## 监控文件变化并运行测试
	@echo "Watching for file changes..."
	@if command -v fswatch >/dev/null 2>&1; then \
		fswatch -o . -e ".*" -i "\\.go$$" | xargs -n1 -I{} make test; \
	else \
		echo "fswatch not found, install it with: brew install fswatch (on macOS)"; \
	fi

# 检查安全漏洞
security: ## 检查安全漏洞
	@echo "Checking for security vulnerabilities..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not found, install it with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# 生成模拟对象
mock: ## 生成模拟对象
	@echo "Generating mocks..."
	@if command -v mockgen >/dev/null 2>&1; then \
		mockgen -source=executors/executors.go -destination=mocks/executors_mock.go; \
	else \
		echo "mockgen not found, install it with: go install github.com/golang/mock/mockgen@latest"; \
	fi

# Docker 构建
docker-build: ## 构建 Docker 镜像
	@echo "Building Docker image..."
	@docker build -t $(APP_NAME):$(VERSION) .

# Docker 运行
docker-run: ## 运行 Docker 容器
	@echo "Running Docker container..."
	@docker run --rm $(APP_NAME):$(VERSION)
