#!/usr/bin/env bash
# 在 macOS/Linux 上打包：后端 linux/amd64 + config + 前端 dist
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

# 兼容两种目录结构：
# - monorepo 根：$ROOT/backend
# - 子目录：$ROOT/09wp-master/backend
PROJ="$ROOT"
if [[ ! -f "$PROJ/backend/go.mod" ]]; then
  if [[ -f "$ROOT/09wp-master/backend/go.mod" ]]; then
    PROJ="$ROOT/09wp-master"
  else
    echo "Cannot locate backend/go.mod under ROOT or ROOT/09wp-master" >&2
    exit 1
  fi
fi

OUT="$PROJ/release/linux-amd64"
ZIP="$PROJ/release/09wp-linux-amd64.zip"

echo "==> 输出目录: $OUT"
rm -rf "$OUT"
mkdir -p "$OUT"

echo "==> go build -> 09wp-server"
(
  cd "$PROJ/backend"
  GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "$OUT/09wp-server" ./cmd/server
)

cp "$PROJ/backend/config.json.example" "$OUT/config.json"
echo "==> 已复制 config.json"

echo "==> pnpm run build"
(cd "$PROJ" && pnpm run build)

mkdir -p "$OUT/web"
cp -r "$PROJ/dist/"* "$OUT/web/"
echo "==> 前端已复制到 web/"

cat > "$OUT/部署说明.txt" <<'EOF'
09wp Linux 发布包 (amd64)
========================

目录内容:
  09wp-server   后端可执行文件
  config.json   配置模板（请改 mysql_dsn、jwt_secret 等）
  web/          前端静态资源（Vite dist）

后端: chmod +x 09wp-server && ./09wp-server
前端: Nginx 根目录指向 web/，/api 反代到后端。
EOF

rm -f "$ZIP"
(cd "$PROJ/release" && zip -r "$(basename "$ZIP")" "linux-amd64")
echo "==> 完成: $ZIP"
