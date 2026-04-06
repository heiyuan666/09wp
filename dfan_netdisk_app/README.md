# DFAN 网盘导航 — Flutter 客户端

对接仓库内 Go 服务：`09wp-master/backend`（路由前缀 **`/api/v1`**）。

## 后端约定

- 成功响应：`{ "code": 200, "message": "…", "data": … }`（注意：部分业务错误仍可能 **HTTP 200** + `code != 200`）。
- 鉴权：登录后请求头 `Authorization: Bearer <jwt>`（普通用户接口 `adminOnly=false`）。
- 默认端口：`8080`（见 `config.json` 的 `http_port`）。

## 本地运行

### 国内网络安装 Flutter（推荐）

从 GitHub 克隆常出现 `RPC failed / Connection reset`，可用 **Gitee 克隆 SDK**，再用**镜像**下载 Dart / 引擎（`flutter doctor` 时约 200MB+）。

**镜像说明**（脚本默认 **`flutter-io`**：官方国内存储，引擎/Dart 一般最全；清华等可能滞后导致 zip 损坏）

| `FLUTTER_MIRROR` | 说明 |
|------------------|------|
| `flutter-io`（默认） | `storage.flutter-io.cn` + `pub.flutter-io.cn`，适合避免「假 zip / 153 字节」 |
| `tuna` / `ustc` / `sjtu` | 教育网镜像，**快**但可能未同步当前 Gitee `stable` 所需引擎 → 易报 `End-of-central-directory` |
| `google` | **不**设 `FLUTTER_STORAGE_BASE_URL`，引擎走 Google（需代理）；`pub` 仍走清华 |

一键脚本（默认安装到 `~/flutter_sdk/flutter`）：

```bash
cd dfan_netdisk_app
chmod +x scripts/install_flutter_cn.sh
./scripts/install_flutter_cn.sh
# 想用清华镜像（若 doctor 报 zip 损坏再改回 flutter-io）：
# FLUTTER_MIRROR=tuna ./scripts/install_flutter_cn.sh
```

**报错 `153 bytes` / `not a zipfile` / `corrupt`**：多为镜像上还没有对应 Dart SDK。先删缓存再换 **`flutter-io`**：

```bash
rm -rf ~/flutter_sdk/flutter/bin/cache
export FLUTTER_MIRROR=flutter-io
export PUB_HOSTED_URL=https://pub.flutter-io.cn
export FLUTTER_STORAGE_BASE_URL=https://storage.flutter-io.cn
export PATH="$PATH:$HOME/flutter_sdk/flutter/bin"
flutter doctor -v
```

**有代理**时可用引擎官方源 + 国内 pub：

```bash
rm -rf ~/flutter_sdk/flutter/bin/cache
FLUTTER_MIRROR=google ./scripts/install_flutter_cn.sh
```

手动克隆（不配脚本时，建议先用 `flutter-io`）：

```bash
mkdir -p ~/flutter_sdk && cd ~/flutter_sdk
export PUB_HOSTED_URL=https://pub.flutter-io.cn
export FLUTTER_STORAGE_BASE_URL=https://storage.flutter-io.cn
git clone https://gitee.com/mirrors/Flutter.git -b stable --depth 1 flutter
export PATH="$PATH:$HOME/flutter_sdk/flutter/bin"
flutter doctor
```

若 Gitee 仍失败，可换网络/VPN 后重试，或去掉 `--depth 1` 再克隆；半克隆失败时请删除不完整目录后重新执行。

#### `flutter doctor` 下载 Dart 中断（Connection reset / partial file / curl 18）

大文件约 200MB，弱网易断。不要用半坏的 zip，先删掉再**断点续传**：

```bash
rm -f ~/flutter_sdk/flutter/bin/cache/dart-sdk-darwin-arm64.zip   # Intel 用 dart-sdk-darwin-x64.zip
cd /Users/pangzinan/Desktop/gomod/dfan_netdisk_app
chmod +x scripts/resume_dart_sdk.sh
export FLUTTER_STORAGE_BASE_URL=https://storage.flutter-io.cn
./scripts/resume_dart_sdk.sh ~/flutter_sdk/flutter
flutter doctor -v
```

脚本用 `curl -C -` 续传，可**反复执行**直到下完。若 `storage.flutter-io.cn` 仍不稳，可开代理后：

```bash
unset FLUTTER_STORAGE_BASE_URL
./scripts/resume_dart_sdk.sh ~/flutter_sdk/flutter
```

（此时 URL 需改脚本里的默认 STORAGE 为 `https://storage.googleapis.com` — 更简单是代理全局后直接 `flutter doctor -v`。）

---

1. 安装 Flutter（见上；官方文档：<https://docs.flutter.dev/get-started/install>）。
2. 若本目录缺少 `android/`、`ios/` 等平台文件夹，在项目根执行：

   ```bash
   cd dfan_netdisk_app
   flutter create .
   ```

3. 安装依赖并运行：

   ```bash
   flutter pub get
   flutter run
   ```

## 配置 API 地址

应用内 **设置** 页可修改「API 根地址」，需包含 **`/api/v1`**，例如：

| 场景 | 示例 |
|------|------|
| 本机 iOS 模拟器 / 桌面 | `http://127.0.0.1:8080/api/v1` |
| Android 模拟器访问本机后端 | `http://10.0.2.2:8080/api/v1` |
| 真机访问电脑局域网 IP | `http://192.168.x.x:8080/api/v1` |

确保 Go 服务已开启 CORS（项目已含 `CORSMiddleware`），移动端调试时后端需监听 `0.0.0.0` 或对应网卡。

## 已实现接口

- `GET /home`、`GET /categories`、`GET /resources`、`GET /search`
- `GET /resources/:id`、`POST /resources/:id/access-link`
- `POST /auth/login`、`GET /user/profile`
- `GET /public/config`（站点标题等）

可在 `lib/services/netdisk_api.dart` 中继续扩展收藏、投稿等接口。

## 搜索与网盘 App 唤起

- 底部导航含 **「搜索」** 标签，对接 `GET /search`。
- 资源详情页 **「网盘链接」** 下列出主链与附加链；点击会按网盘类型依次尝试 **URL Scheme**（如百度 `baiduwangpan://`、夸克 `quark://` 等），失败则回退 **https**，由系统决定打开浏览器或已关联的 App。
- 右上角 **同步** 或 **获取最新** 会调用 `POST /resources/:id/access-link`，与 Web 端一致（含转存成功后的链接合并字段 `links`）。

若 iOS 上 Scheme 无法跳转，可在 `ios/Runner/Info.plist` 的 `LSApplicationQueriesSchemes` 中增加可能用到的 scheme（如 `baiduwangpan`、`quark`、`alipans`、`smartdrive`、`xunlei` 等）。Android 12+ 若需显式查询第三方 App，可在 `AndroidManifest.xml` 的 `<queries>` 中声明对应 `scheme`。

## 扫码登录（Web + App）

流程：**Web 调接口生成二维码 → 手机 App 扫码 → 在 App 内输入同一套用户账号密码 → 后端签发 JWT，Web 轮询拿到 token。**

| 接口 | 说明 |
|------|------|
| `POST /auth/qr/create` | 创建会话，返回 `sid`、`qr_payload`（建议生成二维码）、`expires_at` |
| `GET /auth/qr/status/:sid` | Web 轮询：`pending` / `confirmed`（含 `token`、`user`）/ `expired` |
| `POST /auth/qr/confirm` | App 提交 `sid` + `username` + `password`，与普通登录相同响应 |

二维码内容使用返回的 **`qr_payload`**（`dfannetdisk://qr-login?sid=...`）或 JSON `qr_payload_alt` 均可，App 已做解析。

Vue 侧已封装：`siteQrLoginCreate`、`siteQrLoginStatus`（见 `09wp-master/src/api/netdisk.ts`）。需在登录页展示二维码并定时轮询 `status` 直至 `confirmed` 后写入前台 token。

### App 权限（`flutter create .` 之后）

- **Android**：在 `AndroidManifest.xml` 增加 `<uses-permission android:name="android.permission.CAMERA" />`。
- **iOS**：在 `Info.plist` 增加 `NSCameraUsageDescription`（相册扫码说明文案）。

依赖：`mobile_scanner`（见 `pubspec.yaml`）。
