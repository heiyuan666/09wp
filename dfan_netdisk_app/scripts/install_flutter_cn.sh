#!/usr/bin/env bash
# Flutter SDK 国内安装：Gitee 克隆 + 可选镜像（Dart SDK / 引擎 / pub）
#
# 用法：
#   ./scripts/install_flutter_cn.sh [安装目录，默认 ~/flutter_sdk]
#
# 环境变量 FLUTTER_MIRROR（默认 flutter-io）：
#   flutter-io  官方国内合作源（storage / pub，较慢但引擎文件一般最全）— 默认
#   tuna        清华大学（快，但 Gitee 的 stable 有时比镜像同步新 → 可能下到 153B 假 zip）
#   ustc        中国科大（同上风险）
#   sjtu        上海交大
#   google      引擎走 Google 官方存储（需代理）；pub 仍走清华以加速包下载
#
# 若出现「End-of-central-directory / 153 bytes / corrupt zip」：
#   rm -rf ~/flutter_sdk/flutter/bin/cache
#   FLUTTER_MIRROR=flutter-io ./scripts/install_flutter_cn.sh
# 仍失败可试：FLUTTER_MIRROR=google（开系统代理后执行）

set -euo pipefail

INSTALL_DIR="${1:-$HOME/flutter_sdk}"
FLUTTER_DIR="$INSTALL_DIR/flutter"
MIRROR="${FLUTTER_MIRROR:-flutter-io}"

apply_mirror() {
  case "$MIRROR" in
    tuna)
      export PUB_HOSTED_URL="https://mirrors.tuna.tsinghua.edu.cn/dart-pub"
      export FLUTTER_STORAGE_BASE_URL="https://mirrors.tuna.tsinghua.edu.cn/flutter"
      ;;
    ustc)
      export PUB_HOSTED_URL="https://mirrors.ustc.edu.cn/dart-pub"
      export FLUTTER_STORAGE_BASE_URL="https://mirrors.ustc.edu.cn/flutter"
      ;;
    sjtu)
      export PUB_HOSTED_URL="https://mirror.sjtu.edu.cn/dart-pub"
      export FLUTTER_STORAGE_BASE_URL="https://mirror.sjtu.edu.cn/flutter"
      ;;
    flutter-io)
      export PUB_HOSTED_URL="https://pub.flutter-io.cn"
      export FLUTTER_STORAGE_BASE_URL="https://storage.flutter-io.cn"
      ;;
    google)
      export PUB_HOSTED_URL="https://mirrors.tuna.tsinghua.edu.cn/dart-pub"
      unset FLUTTER_STORAGE_BASE_URL || true
      ;;
    *)
      echo "未知 FLUTTER_MIRROR=$MIRROR ，可选: flutter-io | tuna | ustc | sjtu | google"
      exit 1
      ;;
  esac
}

apply_mirror

echo "==> 使用镜像: $MIRROR"
echo "    PUB_HOSTED_URL=${PUB_HOSTED_URL:-}"
if [[ "$MIRROR" == "google" ]]; then
  echo "    FLUTTER_STORAGE_BASE_URL=(未设置，引擎走 Google，需可访问外网或代理)"
else
  echo "    FLUTTER_STORAGE_BASE_URL=$FLUTTER_STORAGE_BASE_URL"
fi
echo "    （持久化可写入 ~/.zprofile；换源改 FLUTTER_MIRROR）"
echo ""
echo "    提示：若用 tuna/ustc 时出现 dart-sdk zip 损坏、仅下载约 153 字节，"
echo "          说明该镜像尚未同步当前 Flutter 版本所需引擎，请改用 FLUTTER_MIRROR=flutter-io"
echo ""

mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"

if [[ -d "$FLUTTER_DIR/.git" ]]; then
  echo "==> 已存在 $FLUTTER_DIR ，执行 git pull 更新 stable …"
  cd "$FLUTTER_DIR"
  git fetch origin
  git checkout stable
  git pull origin stable
else
  echo "==> 从 Gitee 克隆 Flutter（避免 GitHub 不稳定）…"
  git clone https://gitee.com/mirrors/Flutter.git -b stable --depth 1 "$FLUTTER_DIR"
fi

cd "$FLUTTER_DIR/bin"
echo ""
echo "==> 运行 flutter doctor（首次下载 Dart / 引擎）"
./flutter doctor -v

echo ""
echo "==> 请将下列行加入 ~/.zshrc 或 ~/.zprofile（按你选的镜像调整）："
echo ""
echo "export FLUTTER_MIRROR=$MIRROR"
echo "export PUB_HOSTED_URL=${PUB_HOSTED_URL:-}"
if [[ "$MIRROR" == "google" ]]; then
  echo "# 不设置 FLUTTER_STORAGE_BASE_URL，让引擎走 Google"
else
  echo "export FLUTTER_STORAGE_BASE_URL=$FLUTTER_STORAGE_BASE_URL"
fi
echo "export PATH=\"\$PATH:$FLUTTER_DIR/bin\""
echo ""
echo "若下载损坏或失败，先清缓存再换镜像重试："
echo "  rm -rf \"$FLUTTER_DIR/bin/cache\""
echo "  FLUTTER_MIRROR=flutter-io $FLUTTER_DIR/bin/flutter doctor -v"
