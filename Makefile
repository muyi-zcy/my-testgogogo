.PHONY: demo demo-backend demo-api demo-flow demo-test demo-run demo-report library library-run library-api library-flow check tidy fmt clean help

RUN_ID := $(shell date +%Y%m%d-%H%M%S)

help:
	@echo "my-testgogogo API 集成测试框架"
	@echo ""
	@echo "Demo 示例（自定义 demoauth 认证）:"
	@echo "  make demo-run       一键：启动 backend + 跑测试"
	@echo "  make demo-backend   仅启动 backend (localhost:18080)"
	@echo "  make demo-test      全部测试（需 backend 已启动）"
	@echo ""
	@echo "Library 示例（内置 login 认证）:"
	@echo "  make library-run    一键：启动 backend + 跑测试"
	@echo "  make library-api    单接口测试"
	@echo "  make library-flow   流程测试"
	@echo ""
	@echo "  make check          格式化 + 编译 + demo + library 一键测试"
	@echo "  make tidy           整理依赖"
	@echo "  make clean          清理缓存与临时报告"

tidy:
	go mod tidy
	cd examples/demo && go mod tidy
	cd examples/library && go mod tidy

fmt:
	go fmt ./...
	cd examples/demo && go fmt ./...
	cd examples/library && go fmt ./...

build:
	go build ./...
	go build -o bin/my-testgogogo ./cmd/my-testgogogo

demo-backend:
	$(MAKE) -C examples/demo backend

demo-api:
	$(MAKE) -C examples/demo test-api

demo-flow:
	$(MAKE) -C examples/demo test-flow

demo-test:
	$(MAKE) -C examples/demo test

demo-run:
	$(MAKE) -C examples/demo run

demo-report:
	$(MAKE) -C examples/demo test-report

library:
	$(MAKE) -C examples/library run

library-run:
	$(MAKE) -C examples/library run

library-api:
	$(MAKE) -C examples/library test-api

library-flow:
	$(MAKE) -C examples/library test-flow

library-report:
	$(MAKE) -C examples/library test-report

check: fmt build demo-run library-run

clean:
	rm -rf .cache bin
	$(MAKE) -C examples/demo clean
	$(MAKE) -C examples/library clean
