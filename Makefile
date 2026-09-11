# B2B Marketing Agent Harness
.DEFAULT_GOAL := help
GO      ?= go
BIN_DIR ?= bin
CONFIG  ?= configs/config.yaml

.PHONY: help
help: ## 显示帮助
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

.PHONY: tidy
tidy: ## 整理依赖
	$(GO) mod tidy

.PHONY: fmt
fmt: ## 格式化
	gofmt -w .

.PHONY: vet
vet: ## 静态检查
	$(GO) vet ./...

.PHONY: test
test: ## 单元测试
	$(GO) test ./...

.PHONY: build
build: ## 构建全部命令
	@mkdir -p $(BIN_DIR)
	@for c in api worker scheduler harness; do \
		echo ">> building $$c"; \
		$(GO) build -o $(BIN_DIR)/$$c ./cmd/$$c || exit 1; \
	done

.PHONY: run-api
run-api: ## 启动 API 服务
	$(GO) run ./cmd/api -config $(CONFIG)

.PHONY: run-worker
run-worker: ## 启动 Worker
	$(GO) run ./cmd/worker -config $(CONFIG)

.PHONY: run-scheduler
run-scheduler: ## 启动调度器
	$(GO) run ./cmd/scheduler -config $(CONFIG)

.PHONY: demo
demo: ## 离线跑通半导体营销参考流水线（mock）
	$(GO) run ./cmd/harness -config $(CONFIG) demo

.PHONY: clean
clean: ## 清理产物
	rm -rf $(BIN_DIR)
