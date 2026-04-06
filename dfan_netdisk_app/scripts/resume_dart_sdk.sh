#!/usr/bin/env bash
# 断点续传下载 Dart SDK（解决 flutter doctor 时 curl 中断、Connection reset、partial file）
#
# 用法：
#   chmod +x scripts/resume_dart_sdk.sh
#   ./scripts/resume_dart_sdk.sh [Flutter 根目录，默认 ~/flutter_sdk/flutter]
#
# 依赖环境变量（与 flutter doctor 一致，默认国内官方镜像）：
#   FLUTTER_STORAGE_BASE_URL  默认 https://storage.flutter-io.cn
#
# 完成后执行：flutter doctor -v

set -euo pipefail

FLUTTER_ROOT="${1:-$HOME/flutter_sdk/flutter}"
STORAGE="${FLUTTER_STORAGE_BASE_URL:-https://storage.flutter-io.cn}"
CACHE="$FLUTTER_ROOT/bin/cache"

resolve_engine_rev() {
  local f
  for f in \
    "$FLUTTER_ROOT/bin/internal/engine.version" \
    "$CACHE/engine.stamp" \
    "$FLUTTER_ROOT/bin/cache/engine.stamp"
  do
    if [[ -f "$f" ]]; then
      tr -d ' \n\r\t' < "$f"
      return 0
    fi
  done
  echo ""
}

REV="$(resolve_engine_rev)"
if [[ -z "$REV" ]]; then
  echo "无法读取 engine 版本，请确认 Flutter 目录正确: $FLUTTER_ROOT"
  exit 1
fi

OS="$(uname -s)"
ARCH="$(uname -m)"
PLAT=""
case "$OS" in
  Darwin)
    case "$ARCH" in
      arm64)  PLAT="darwin-arm64" ;;
      x86_64) PLAT="darwin-x64" ;;
      *) echo "未支持的 Mac 架构: $ARCH"; exit 1 ;;
    esac
    ;;
  Linux)
    case "$ARCH" in
      aarch64|arm64) PLAT="linux-arm64" ;;
      x86_64)        PLAT="linux-x64" ;;
      *) echo "未支持的 Linux 架构: $ARCH"; exit 1 ;;
    esac
    ;;
  *)
    echo "未支持的系统: $OS"
    exit 1
    ;;
esac

ZIP_NAME="dart-sdk-${PLAT}.zip"
URL="${STORAGE}/flutter_infra_release/flutter/${REV}/${ZIP_NAME}"
OUT="$CACHE/${ZIP_NAME}"

mkdir -p "$CACHE"

echo "==> Engine: $REV"
echo "==> 下载（支持断点续传 -C -，多次重试）"
echo "    $URL"
echo "    -> $OUT"
echo ""

# -C - 续传；-f 失败时不把错误页当 zip；--retry 应对偶发 reset（旧版 curl 无 retry-all-errors）
curl -fL \
  --connect-timeout 30 \
  --retry 20 \
  --retry-delay 5 \
  -C - \
  -o "$OUT" \
  "$URL"

echo ""
echo "==> 校验 zip 大小（应约 200MB 级，若仅几 KB 请换 FLUTTER_STORAGE_BASE_URL 或检查网络）"
ls -lh "$OUT"

echo ""
echo "==> 解压到 $CACHE"
rm -rf "$CACHE/dart-sdk"
unzip -q -o "$OUT" -d "$CACHE"

# 与官方脚本一致：标记已安装该 engine 的 Dart SDK
echo "$REV" > "$CACHE/engine-dart-sdk.stamp"

echo ""
echo "==> 完成。请执行:"
echo "    export PATH=\"\$PATH:$FLUTTER_ROOT/bin\""
echo "    flutter doctor -v"
