#!/usr/bin/env bash
#
# start_nohup.sh
# 用 nohup 后台启动 tcp-collector（避免终端断开导致进程退出）
#
# 说明：
# - 默认使用 ./tcp-collector-linux-amd64（当前目录）
# - 默认使用 config.yaml（位于项目根目录）
# - 输出日志到 logs/nohup.out（与程序自身日志配置独立）
#
# 用法：
#   ./start_nohup.sh
#   ./start_nohup.sh -config config.yaml
#   BIN=./tcp-collector ./start_nohup.sh -config config.yaml
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

LOG_DIR="${LOG_DIR:-$SCRIPT_DIR/logs}"
mkdir -p "$LOG_DIR"

NOHUP_LOG="${NOHUP_LOG:-$LOG_DIR/nohup.out}"
PID_FILE="${PID_FILE:-$LOG_DIR/tcp-collector.pid}"

# 旧逻辑（保留注释，便于回滚）：默认使用 bin/ 目录产物
# DEFAULT_BIN="$SCRIPT_DIR/bin/tcp-collector-linux-amd64"
#
# 新逻辑：默认使用当前目录的二进制；也允许外部覆盖 BIN
DEFAULT_BIN="$SCRIPT_DIR/tcp-collector-linux-amd64"
BIN="${BIN:-$DEFAULT_BIN}"

if [[ ! -x "$BIN" ]]; then
  echo "[ERROR] 找不到可执行文件或不可执行: $BIN" >&2
  echo "        你可以：chmod +x \"$BIN\"，或用 BIN=... 指定正确路径" >&2
  exit 1
fi

if [[ -f "$PID_FILE" ]]; then
  OLD_PID="$(cat "$PID_FILE" || true)"
  if [[ -n "${OLD_PID:-}" ]] && kill -0 "$OLD_PID" >/dev/null 2>&1; then
    echo "[WARN] 已在运行 (pid=$OLD_PID)。如果你确认要重启，请先停止该进程或删掉 $PID_FILE" >&2
    exit 0
  fi
  rm -f "$PID_FILE"
fi

# 默认参数：如果用户没传 -config，就补一个 -config config.yaml
ARGS=("$@")
HAS_CONFIG=0
for ((i=0; i<${#ARGS[@]}; i++)); do
  if [[ "${ARGS[$i]}" == "-config" ]]; then
    HAS_CONFIG=1
    break
  fi
done
if [[ "$HAS_CONFIG" -eq 0 ]]; then
  ARGS+=("-config" "config.yaml")
fi

echo "[INFO] 启动中..."
echo "       BIN=$BIN"
echo "       ARGS=${ARGS[*]}"
echo "       NOHUP_LOG=$NOHUP_LOG"

nohup "$BIN" "${ARGS[@]}" >>"$NOHUP_LOG" 2>&1 &
NEW_PID="$!"
echo "$NEW_PID" > "$PID_FILE"

sleep 0.2
if kill -0 "$NEW_PID" >/dev/null 2>&1; then
  echo "[OK] 已启动 (pid=$NEW_PID). tail -f \"$NOHUP_LOG\" 查看输出"
else
  echo "[ERROR] 启动失败：进程未存活。请查看 $NOHUP_LOG" >&2
  exit 1
fi
