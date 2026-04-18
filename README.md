<div align="center">
  <img src="./public/logo.svg" alt="Logo" width="120" />

# 09 网盘资源导航

</div>

基于 **Vue 3 + Vite + Element Plus / Semi Design** 的管理后台与前台站点，配套 **Go（Gin）+ MySQL** 后端 API。提供网盘资源收录、搜索、多网盘转存、TG/RSS 同步、全网搜聚合、游戏/软件子站点等能力；默认 **关闭 MSW**，直连真实后端开发调试。

## 功能概览

| 模块 | 说明 |
| :--- | :--- |
| 资源与分类 | 资源 CRUD、分类、链接检测、转存日志 |
| 前台搜索 | 关键词搜索、筛选、TMDB/豆瓣卡片、**全网搜**（多线路聚合、链接检测） |
| 网盘转存 | 夸克 / 百度 / UC / 115 / 天翼 / 123 / 阿里云盘 / 迅雷等（依赖后台凭证配置） |
| 系统配置 | 站点信息、SEO、号卡、Meilisearch、全网搜站点与线路、详情页转存策略等 |
| 可选能力 | Redis 搜索缓存、Meilisearch、Telegram 频道同步、RSS、用户提交与反馈等 |

## 技术栈

- **前端**：Vue 3、Vite、TypeScript、Element Plus、Pinia、Vue Router；部分页面使用 React + Semi Design（如前台搜索页）。
- **后端**：Go 1.25+、Gin、GORM、MySQL；可选 Redis、Meilisearch。

## 环境要求

- **Node.js**：`^20.19.0` 或 `>=22.12.0`
- **Go**：`1.25+`（以 `backend/go.mod` 为准）
- **MySQL**：`5.7+` / `8.x`，字符集建议 `utf8mb4`

可选：**Redis**（搜索缓存等）、**Meilisearch**（加速搜索）。

## 快速开始

### 1. 克隆与依赖

```bash
git clone <你的仓库地址> 09wp
cd 09wp
```

前端依赖（任选其一包管理器）：

```bash
npm install
# 或
pnpm install
```

### 2. 配置数据库与后端

在 `backend` 目录通过 **环境变量** 或 **`config.json`** 指定 MySQL 等配置（环境变量优先生效）。

常用环境变量示例：

| 变量 | 说明 |
| :--- | :--- |
| `MYSQL_DSN` | 例如 `user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local` |
| `HTTP_PORT` | HTTP 监听端口，默认 `8080` |
| `JWT_SECRET` | JWT 密钥，生产环境务必修改 |
| `CONFIG_PATH` | 指向自定义 JSON 配置文件路径（可选） |
| `REDIS_ENABLED` | `1` / `true` 开启 Redis（需同时配置 `REDIS_HOST` 等） |
| `MEILI_ENABLED` | `1` / `true` 开启 Meilisearch |

也可在 `backend` 下放置 `config.json`（或 `configs/config.json`），字段见 `backend/internal/config/config.go` 中的 `fileConfig`。

首次启动会自动 **迁移表结构** 并 **种子数据**（含默认管理员，见下方「默认账号」）。

### 3. 启动后端 API

在仓库根目录执行：

```bash
cd backend
go run ./cmd/server
```

默认监听 **`http://localhost:8080`**（若未改端口）。调试网盘转存日志可加：

```bash
go run ./cmd/server -debug
# 或设置环境变量 DEBUG=1 / APP_DEBUG=1
```

### 4. 启动前端（开发）

在仓库根目录新建或编辑 **`.env.development`**（若不存在可从团队模板复制），至少保证 API 指向后端，例如：

```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

然后：

```bash
npm run dev
# 或
pnpm dev
```

开发服务器默认 **`http://localhost:3007`**（见 `vite.config.ts`），并已配置将 `/api`、`/public` 等代理到后端根地址。

### 5. 生产构建

```bash
npm run build
# 或
pnpm build
```

产物在 `dist/`，需配合静态资源服务器与反向代理，将 **`/api/v1`** 指到 Go 服务。

## 默认账号（种子数据）

后端首次初始化后，默认管理员账号一般为 **`admin` / `123456`**（请以 `backend/internal/database` 中种子逻辑为准，生产环境登录后立即修改密码）。

## 目录结构（摘要）

```text
09wp-master/
├── backend/                 # Go API（cmd/server 入口）
│   ├── cmd/server/          # main 启动
│   ├── internal/            # 业务、路由、配置、数据库
│   └── docs/                # Swagger 文档（若已生成）
├── src/                     # Vue 管理端 + 部分前台（Vue/React）
├── public/                  # 静态资源
├── b_jF6ffPlZxzW/           # Next.js 子项目（若使用）
├── vite.config.ts
└── package.json
```

## 相关文档

- 后端 Swagger：服务运行后访问项目内配置的 docs 地址（若已集成 `swag` 生成）。
- 前台 API 基址：`src/config/app.config.ts` 中的 `API_BASE_URL` 与 `VITE_API_BASE_URL`。

## 许可证

本项目采用 [MIT License](./LICENSE) 开源协议。

**允许二次开发/修改/自用部署，但禁止商业用途**（包括但不限于：售卖、收费部署/代运营、作为商业产品/服务的一部分提供、以营利为目的的分发与推广等）。

如需商业授权，请联系作者获得书面许可。
