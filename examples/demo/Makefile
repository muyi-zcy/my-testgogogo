.PHONY: backend test test-api test-flow test-report run clean help

BACKEND_PORT ?= 18080
RUN_ID := $(shell date +%Y%m%d-%H%M%S)

help:
	@echo "demo 示例（backend + 测试一体）"
	@echo ""
	@echo "  make backend      启动 mock 后端"
	@echo "  make test         运行全部测试（需 backend 已启动）"
	@echo "  make run          自动启动 backend 并运行测试"
	@echo "  make test-api     单接口测试"
	@echo "  make test-flow    流程测试"
	@echo "  make test-report  全部测试 + Markdown 报告"
	@echo "  make clean        清理缓存与报告"

backend:
	BACKEND_PORT=$(BACKEND_PORT) go run ./backend

test:
	go test ./... -v -count=1

test-api:
	go test ./api/... -v -count=1

test-flow:
	go test ./flow/... -v -count=1

run:
	@chmod +x ./scripts/run.sh
	@./scripts/run.sh

test-report:
	@mkdir -p reports/staging
	@chmod +x ./scripts/run.sh
	@MY_TESTGOGOGO_REPORT_RUN_ID=$(RUN_ID) ./scripts/run.sh -json 2>&1 | tee reports/staging/go-test-$(RUN_ID).json | go run ../../cmd/my-testgogogo report --run-id $(RUN_ID) --command "go test ./... -json -count=1"

clean:
	rm -rf .cache reports/staging reports/**/*.md
