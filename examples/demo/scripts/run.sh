#!/usr/bin/env bash
# demo 示例一键运行脚本：
#   1. 启动 Mock 后端（默认端口 18080）
#   2. 等待后端就绪
#   3. 执行全部测试
#   4. 退出时自动停止后端进程
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PORT="${BACKEND_PORT:-18080}"
BASE_URL="http://localhost:${PORT}"

cd "$ROOT"

echo "==> starting backend on ${BASE_URL}"
BACKEND_PORT="$PORT" go run ./backend &
BACKEND_PID=$!

# 注册清理函数：脚本退出时终止后端进程
cleanup() {
  if kill -0 "$BACKEND_PID" 2>/dev/null; then
    kill "$BACKEND_PID" 2>/dev/null || true
    wait "$BACKEND_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT INT TERM

# 轮询等待后端 /api/system/info 可访问（最多 10 秒）
echo "==> waiting for backend..."
for _ in $(seq 1 50); do
  if curl -sf "${BASE_URL}/api/system/info" >/dev/null 2>&1; then
    echo "==> backend ready"
    break
  fi
  sleep 0.2
done

if ! curl -sf "${BASE_URL}/api/system/info" >/dev/null 2>&1; then
  echo "backend failed to start on ${BASE_URL}" >&2
  exit 1
fi

echo "==> running tests"
go test ./... -v -count=1 "$@"
