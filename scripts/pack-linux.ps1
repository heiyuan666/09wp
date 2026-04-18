# Pack: Linux amd64 Go binary + Vite dist -> release/*.zip
# Run from repo root: pnpm run pack:linux
$ErrorActionPreference = 'Stop'
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

$stamp = Get-Date -Format 'yyyyMMdd-HHmmss'
$outName = "dfan-netdisk-linux-amd64-$stamp"
$stage = Join-Path $Root "release\$outName"
$zipPath = Join-Path $Root "release\$outName.zip"

Write-Host '==> frontend: pnpm run build-only' -ForegroundColor Cyan
pnpm run build-only
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

$indexHtml = Join-Path $Root 'dist\index.html'
if (-not (Test-Path $indexHtml)) {
  Write-Error 'Missing dist/index.html. Frontend build failed.'
}

Write-Host '==> backend: GOOS=linux GOARCH=amd64 go build' -ForegroundColor Cyan
Push-Location (Join-Path $Root 'backend')
$env:GOOS = 'linux'
$env:GOARCH = 'amd64'
$env:CGO_ENABLED = '0'
go build -trimpath -ldflags '-s -w' -o 'dfan-netdisk-server' ./cmd/server
if ($LASTEXITCODE -ne 0) { Pop-Location; exit $LASTEXITCODE }
Pop-Location

New-Item -ItemType Directory -Force -Path (Split-Path $stage -Parent) | Out-Null
if (Test-Path $stage) { Remove-Item -Recurse -Force $stage }
New-Item -ItemType Directory -Force -Path $stage | Out-Null

$binSrc = Join-Path $Root 'backend\dfan-netdisk-server'
$binDst = Join-Path $stage 'dfan-netdisk-server'
Copy-Item $binSrc $binDst -Force
Copy-Item (Join-Path $Root 'dist') (Join-Path $stage 'dist') -Recurse -Force

$ts = Get-Date -Format 'yyyy-MM-dd HH:mm:ss'
$readmeLines = @(
  'DFAN Netdisk - Linux amd64 bundle',
  '==================================',
  '',
  'Contents:',
  '  dfan-netdisk-server  - Go backend (chmod +x on Linux)',
  '  dist/                - Vite frontend static assets',
  '',
  'Deploy:',
  '  1. Backend: set MySQL/JWT/Redis env vars, run ./dfan-netdisk-server',
  '  2. Frontend: serve dist/ with nginx; reverse-proxy /api to backend',
  '  3. Rebuild frontend with correct VITE_* then pnpm run pack:linux',
  '',
  "Generated: $ts"
)
$readme = $readmeLines -join [Environment]::NewLine
Set-Content -Path (Join-Path $stage 'README.txt') -Value $readme -Encoding utf8

if (Test-Path $zipPath) { Remove-Item -Force $zipPath }
Compress-Archive -Path $stage -DestinationPath $zipPath -Force

Remove-Item -Recurse -Force $stage
Remove-Item -Force $binSrc -ErrorAction SilentlyContinue

Write-Host "==> done: $zipPath" -ForegroundColor Green
