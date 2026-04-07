# 打包 Linux 后端 + config 模板 + 前端 dist，输出到 release/linux-amd64/，并生成 zip
$ErrorActionPreference = "Stop"
$RepoRoot = Split-Path -Parent $PSScriptRoot

# 兼容两种目录结构：
# - monorepo 根：$RepoRoot/backend
# - 子目录：$RepoRoot/09wp-master/backend
$ProjectRoot = $RepoRoot
if (!(Test-Path (Join-Path $ProjectRoot "backend/go.mod"))) {
  $alt = Join-Path $RepoRoot "09wp-master"
  if (Test-Path (Join-Path $alt "backend/go.mod")) {
    $ProjectRoot = $alt
  } else {
    throw "Cannot locate backend/go.mod under RepoRoot or RepoRoot/09wp-master"
  }
}

$OutDir = Join-Path $ProjectRoot "release/linux-amd64"
$ZipPath = Join-Path $ProjectRoot "release/09wp-linux-amd64.zip"

Write-Host "==> Output: $OutDir"

if (Test-Path $OutDir) {
  Remove-Item $OutDir -Recurse -Force
}
New-Item -ItemType Directory -Path $OutDir -Force | Out-Null

# 1) Linux amd64 binary (no CGO)
Push-Location (Join-Path $ProjectRoot "backend")
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
$bin = Join-Path $OutDir "09wp-server"
Write-Host "==> go build -> 09wp-server"
go build -trimpath -ldflags="-s -w" -o $bin ./cmd/server
Pop-Location

Copy-Item (Join-Path $ProjectRoot "backend/config.json.example") (Join-Path $OutDir "config.json") -Force
Write-Host "==> config.json copied from config.json.example"

# 3) Frontend
Push-Location $ProjectRoot
Write-Host "==> pnpm run build"
pnpm run build
$web = Join-Path $OutDir "web"
New-Item -ItemType Directory -Path $web -Force | Out-Null
Copy-Item (Join-Path $ProjectRoot "dist/*") $web -Recurse -Force
Pop-Location
Write-Host "==> dist -> web/"

# 4) Deploy notes (single-quoted here-string avoids & parsing)
$readme = @'
09wp Linux bundle (amd64)
===========================

Contents:
  09wp-server   API binary
  config.json   Template from config.json.example (edit mysql_dsn, jwt_secret)
  web/          Frontend static files (Vite dist)

Run backend:
  chmod +x 09wp-server
  ./09wp-server

Optional env:
  CONFIG_PATH=/opt/09wp/config.json
  HTTP_PORT=8080
  MYSQL_DSN
  JWT_SECRET

Frontend:
  Point Nginx root to web/ and proxy /api/ to the backend (default :8080).
'@

$readmePath = Join-Path $OutDir "DEPLOY.txt"
$utf8NoBom = New-Object System.Text.UTF8Encoding $false
[System.IO.File]::WriteAllText($readmePath, $readme, $utf8NoBom)

if (Test-Path $ZipPath) {
  Remove-Item $ZipPath -Force
}
# 避免 Compress-Archive 在部分 WinPS 上触发 Write-Progress 索引越界
Add-Type -AssemblyName System.IO.Compression.FileSystem
[System.IO.Compression.ZipFile]::CreateFromDirectory(
  $OutDir,
  $ZipPath,
  [System.IO.Compression.CompressionLevel]::Optimal,
  $false
)
Write-Host "==> Done: $ZipPath"
Write-Host "    Folder: $OutDir"
