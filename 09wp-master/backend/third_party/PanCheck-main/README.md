# PanCheck - 网盘链接检测系统

PanCheck 是一个功能强大的网盘链接有效性检测系统，支持批量检测多种主流网盘平台的分享链接是否有效。

## ✨ 特性

- 🔍 **多平台支持**：支持检测 9 种主流网盘平台的链接
- ⚡ **高性能检测**：支持并发检测，可配置检测频率和超时时间
- 📊 **数据统计**：提供详细的检测统计和数据分析
- 🔄 **定时任务**：支持创建定时检测任务，自动检测链接有效性
- 💾 **数据持久化**：使用 MySQL 存储检测记录，Redis 缓存失效链接
- 🎨 **现代化界面**：基于 React + TypeScript 的现代化管理后台
- 🐳 **容器化部署**：提供 Docker Compose 一键部署方案

## 📦 支持的网盘平台

- 夸克网盘 
- UC网盘 
- 百度网盘 
- 天翼云盘 
- 123网盘 
- 115网盘 
- 阿里云盘 
- 迅雷云盘 
- 中国移动云盘 

## 🚀 快速开始

### 环境要求

- Docker 和 Docker Compose
- 或 Go 1.23+ 和 Node.js 18+（本地开发）

### 使用 Docker Hub 部署

docker-compose.yml

```bash
services:
  pancheck:
    image: lampon/pancheck:latest
    container_name: pancheck
    ports:
      - "6080:6080"
    environment:
      - SERVER_PORT=6080 # 服务端口
      - SERVER_MODE=release # 服务模式
      - SERVER_CORS_ORIGINS=* # 跨域请求允许的源
      - DATABASE_TYPE=mysql # 数据库类型
      - DATABASE_HOST=db # 数据库地址
      - DATABASE_PORT=3306 # 数据库端口
      - DATABASE_USER=root # 数据库用户名
      - DATABASE_PASSWORD=your_password # 数据库密码
      - DATABASE_DATABASE=pancheck # 数据库名称
      - DATABASE_CHARSET=utf8mb4 # 数据库字符集
      - CHECKER_DEFAULT_CONCURRENCY=5 # 默认并发数
      - CHECKER_TIMEOUT=30 # 超时时间（秒）
      - REDIS_ENABLED=true # 是否启用Redis
      - REDIS_HOST=redis # Redis地址
      - REDIS_PORT=6379 # Redis端口
      - REDIS_USERNAME= # Redis用户名
      - REDIS_PASSWORD= # Redis密码
      - REDIS_INVALID_TTL=168   # 失效链接缓存时间（小时）
      - ADMIN_PASSWORD=admin123 # 后台管理密码
    volumes:
      - ./data:/app/data
    restart: unless-stopped
    depends_on:
      - db
      - redis
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:6080/api/v1/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
  db:
    image: mysql:8.0
    container_name: pancheck-db
    environment:
      - MYSQL_ROOT_PASSWORD=your_password
      - MYSQL_DATABASE=pancheck
      - MYSQL_CHARACTER_SET_SERVER=utf8mb4
      - MYSQL_COLLATION_SERVER=utf8mb4_unicode_ci
    volumes:
      - mysql_data:/var/lib/mysql
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost", "-u", "root", "-p$$MYSQL_ROOT_PASSWORD"]
      interval: 10s
      timeout: 5s
      retries: 5
  redis:
    image: redis:latest
    container_name: pancheck-redis
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    restart: unless-stopped
    command: redis-server --appendonly yes
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 3
volumes:
  redis-data:
  mysql_data:
    driver: local
```



主要配置项说明：

```env
# 服务器配置
SERVER_PORT=6080                    # 服务端口
SERVER_MODE=release                 # 服务模式（release/debug）
SERVER_CORS_ORIGINS=*               # 跨域请求允许的源

# 数据库配置
DATABASE_TYPE=mysql                 # 数据库类型
DATABASE_HOST=db                    # 数据库地址（Docker 中使用服务名）
DATABASE_PORT=3306                  # 数据库端口
DATABASE_USER=root                  # 数据库用户名
DATABASE_PASSWORD=your_password     # 数据库密码（请修改）
DATABASE_DATABASE=pancheck          # 数据库名称
DATABASE_CHARSET=utf8mb4            # 数据库字符集

# 检测器配置
CHECKER_DEFAULT_CONCURRENCY=5      # 默认并发数
CHECKER_TIMEOUT=30                  # 超时时间（秒）

# Redis 配置（可选）
REDIS_ENABLED=true                  # 是否启用 Redis
REDIS_HOST=redis                    # Redis 地址
REDIS_PORT=6379                     # Redis 端口
REDIS_PASSWORD=                     # Redis 密码（可选）
REDIS_INVALID_TTL=168               # 失效链接缓存时间（小时）

# 管理员密码配置
ADMIN_PASSWORD=admin123             # 后台管理密码（请修改）
```


服务启动后，访问 `http://localhost:6080` 即可使用。

### 本地开发部署

#### 后端

1. **安装依赖**

```bash
go mod download
```

2. **配置数据库**

创建 MySQL 数据库，并配置 `.env` 文件中的数据库连接信息。

3. **运行服务**

```bash
go run cmd/api/main.go
```

#### 前端

1. **安装依赖**

```bash
cd frontend
pnpm install
```

2. **开发模式运行**

```bash
pnpm run dev
```

3. **构建生产版本**

```bash
pnpm run build
```

构建后的文件会输出到 `frontend/dist` 目录。

## 📡 API 接口使用

### 检测网盘链接

**接口地址：** `POST /api/v1/links/check`

**请求头：**

```
Content-Type: application/json
```

**请求体：**

```json
{
  "links": [
    "https://pan.baidu.com/s/1example",
    "https://www.aliyundrive.com/s/2example"
  ],
  "selectedPlatforms": ["baidu", "aliyun"]  // 可选，指定要检测的平台
}
```

**参数说明：**

- `links` (必需): 要检测的链接数组，每行一个链接
- `selectedPlatforms` (可选): 指定要检测的平台数组，如果为空或未提供则检测所有平台

**`selectedPlatforms` 可选值：**

| 参数值 | 对应网盘平台 |
|--------|------------|
| `quark` | 夸克网盘 |
| `uc` | UC网盘 |
| `baidu` | 百度网盘 |
| `tianyi` | 天翼云盘 |
| `pan123` | 123网盘 |
| `pan115` | 115网盘 |
| `aliyun` | 阿里云盘 |
| `xunlei` | 迅雷云盘 |
| `cmcc` | 中国移动云盘 |

**示例：**

```json
{
  "links": [
    "https://pan.baidu.com/s/1example",
    "https://www.aliyundrive.com/s/2example",
    "https://pan.quark.cn/s/3example"
  ],
  "selectedPlatforms": ["baidu", "aliyun", "quark"]  // 只检测这三个平台
}
```

**响应示例：**

```json
{
  "submission_id": 1,
  "valid_links": [
    "https://pan.baidu.com/s/1example"
  ],
  "invalid_links": [
    "https://www.aliyundrive.com/s/2example"
  ],
  "pending_links": [],
  "total_duration": 2.5
}
```

**响应字段说明：**

- `submission_id`: 提交记录 ID
- `valid_links`: 有效链接列表
- `invalid_links`: 失效链接列表
- `pending_links`: 待检测链接列表（可能因频率限制等原因延迟检测）
- `total_duration`: 检测总耗时（秒）

**使用示例（cURL）：**

```bash
curl -X POST http://localhost:6080/api/v1/links/check \
  -H "Content-Type: application/json" \
  -d '{
    "links": [
      "https://pan.baidu.com/s/1example",
      "https://www.aliyundrive.com/s/2example"
    ]
  }'
```

**使用示例（JavaScript）：**

```javascript
const response = await fetch('http://localhost:6080/api/v1/links/check', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    links: [
      'https://pan.baidu.com/s/1example',
      'https://www.aliyundrive.com/s/2example'
    ],
    selectedPlatforms: ['baidu', 'aliyun'] // 可选
  })
});

const result = await response.json();
console.log('有效链接:', result.valid_links);
console.log('失效链接:', result.invalid_links);
```



## 🔐 管理后台

### 访问管理后台

1. 在浏览器中访问：`http://localhost:6080/admin/login`

2. 输入管理员密码（默认密码在 `.env` 文件中的 `ADMIN_PASSWORD` 配置，默认值为 `admin`）

3. 登录成功后，可以访问以下功能：

   - **仪表盘** (`/admin/dashboard`): 查看检测统计、数据概览
   - **设置** (`/admin/settings`): 配置检测频率、Redis 缓存等
   - **定时任务** (`/admin/scheduled-tasks`): 创建和管理定时检测任务

### 修改管理员密码

修改 `.env` 文件中的 `ADMIN_PASSWORD` 配置项，然后重启服务：

```bash
docker-compose restart pancheck
```

或在 `docker-compose.yml` 中直接修改 `ADMIN_PASSWORD` 环境变量。



## ⚙️ 配置说明

### 检测器配置

可以在管理后台的设置页面配置各平台的检测参数：

- **并发数** (Concurrency): 同时检测的链接数量
- **请求延迟** (Request Delay): 每次请求之间的延迟（毫秒）
- **每秒最大请求数** (Max Requests Per Second): 限制请求频率
- **缓存 TTL** (Cache TTL): 失效链接缓存时间（小时）

### Redis 配置

启用 Redis 可以缓存失效链接，避免重复检测，提高性能。在 `.env` 文件中配置：

```env
REDIS_ENABLED=true
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password
REDIS_INVALID_TTL=168  # 失效链接缓存 168 小时（7天）
```

## 🔧 故障排查


### 检测结果不准确

1. 检查网络连接是否正常
2. 某些平台可能有频率限制，适当调整检测频率配置
3. 查看后台日志了解详细错误信息

### 管理后台无法访问

1. 确认服务已正常启动
2. 检查防火墙设置
3. 确认访问地址正确：`http://localhost:6080/admin/login`




