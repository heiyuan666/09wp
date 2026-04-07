### 通用说明

所有接口统一前缀为 **`/api/v1`**，统一返回格式为：

```json
{
  "code": 0,
  "message": "ok",
  "data": ...
}
```

- `code = 0` 表示成功，其它为错误码（如 400 参数错误、401 未登录、403 无权限、500 服务器错误）。
- 需要登录的接口通过 `Authorization: Bearer <token>` 传递 JWT。

---

## 1. 鉴权相关

### 1.1 管理员登录

- **URL**：`POST /admin/login`
- **说明**：后台管理员登录，获取管理员 token。
- **请求体**：

```json
{
  "username": "admin",
  "password": "123456"
}
```

- **成功响应**：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "token": "xxxx",
    "admin": {
      "id": 1,
      "username": "admin",
      "email": "admin@example.com"
    }
  }
}
```

### 1.2 用户注册

- **URL**：`POST /auth/register`
- **说明**：前台普通用户注册。
- **请求体**：

```json
{
  "username": "test",
  "email": "test@example.com",
  "password": "123456"
}
```

- **成功响应**：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "username": "test",
    "email": "test@example.com"
  }
}
```

### 1.3 用户登录

- **URL**：`POST /auth/login`
- **说明**：前台普通用户登录，用户名或邮箱均可。
- **请求体**：

```json
{
  "username": "test",
  "password": "123456"
}
```

- **成功响应**：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "token": "xxxx",
    "user": {
      "id": 1,
      "username": "test",
      "email": "test@example.com"
    }
  }
}
```

---

## 2. 前台接口

### 2.1 首页数据

- **URL**：`GET /home`
- **说明**：返回最新资源、热门资源、分类导航。
- **成功响应**：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "latest": [],
    "hot": [],
    "categories": []
  }
}
```

### 2.2 资源列表

- **URL**：`GET /resources`
- **说明**：按分类查看资源列表，支持分页、排序。
- **查询参数**：
  - `page`（可选，默认 `1`）
  - `page_size`（可选，默认 `20`，最大 `100`）
  - `category_id`（可选）
  - `sort`（可选，`latest` 或 `hot`，默认 `latest`）

- **成功响应**：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "list": [],
    "total": 0
  }
}
```

### 2.3 资源详情

- **URL**：`GET /resources/:id`
- **说明**：资源详情，自动点击量 +1。

### 2.4 搜索资源

- **URL**：`GET /search`
- **说明**：按标题关键词模糊搜索。
- **查询参数**：
  - `q`（必填，关键词）
  - `page`（可选）
  - `page_size`（可选）

- **成功响应**（分页结构同上）。

### 2.5 用户个人中心

#### 2.5.1 获取个人信息（需登录）

- **URL**：`GET /user/profile`

#### 2.5.2 修改密码（需登录）

- **URL**：`PUT /user/password`
- **请求体**：

```json
{
  "old_password": "123456",
  "new_password": "654321"
}
```

#### 2.5.3 查看收藏资源（需登录）

- **URL**：`GET /user/favorites`

#### 2.5.4 收藏资源（需登录）

- **URL**：`POST /user/favorites/:resource_id`

#### 2.5.5 取消收藏（需登录）

- **URL**：`DELETE /user/favorites/:resource_id`

---

## 3. 后台管理接口（需管理员登录）

> 所有后台接口需携带管理员 JWT，`Authorization: Bearer <admin_token>`。

### 3.1 用户管理

#### 3.1.1 用户列表

- **URL**：`GET /admin/users`
- **查询参数**：
  - `page`
  - `page_size`
  - `keyword`（用户名/邮箱模糊搜索）
  - `status`

#### 3.1.2 修改用户状态

- **URL**：`PUT /admin/users/:id/status`
- **请求体**：

```json
{
  "status": 1
}
```

#### 3.1.3 删除用户

- **URL**：`DELETE /admin/users/:id`

---

### 3.2 分类管理

#### 3.2.1 分类列表

- **URL**：`GET /admin/categories`

#### 3.2.2 新增分类

- **URL**：`POST /admin/categories`
- **请求体**：

```json
{
  "name": "电影",
  "slug": "movies",
  "sort_order": 10,
  "status": 1
}
```

#### 3.2.3 编辑分类

- **URL**：`PUT /admin/categories/:id`

#### 3.2.4 删除分类

- **URL**：`DELETE /admin/categories/:id`

#### 3.2.5 修改分类状态

- **URL**：`PUT /admin/categories/:id/status`

#### 3.2.6 修改分类排序

- **URL**：`PUT /admin/categories/:id/sort`

---

### 3.3 资源管理

#### 3.3.1 资源列表

- **URL**：`GET /admin/resources`
- **查询参数**：
  - `page`
  - `page_size`
  - `title`
  - `category_id`
  - `status`

#### 3.3.2 新增资源

- **URL**：`POST /admin/resources`
- **请求体**：

```json
{
  "title": "资源标题",
  "link": "https://pan.baidu.com/...",
  "category_id": 1,
  "description": "描述",
  "extract_code": "abcd",
  "cover": "https://img...",
  "tags": "标签1,标签2",
  "sort_order": 10,
  "status": 1
}
```

#### 3.3.3 编辑资源

- **URL**：`PUT /admin/resources/:id`

#### 3.3.4 删除资源

- **URL**：`DELETE /admin/resources/:id`

#### 3.3.5 批量删除资源

- **URL**：`POST /admin/resources/batch-delete`
- **请求体**：

```json
{
  "ids": [1, 2, 3]
}
```

#### 3.3.6 批量修改资源状态

- **URL**：`POST /admin/resources/batch-status`
- **请求体**：

```json
{
  "ids": [1, 2, 3],
  "status": 1
}
```

### 3.4 网盘凭证与转存

#### 3.4.1 更新网盘凭证（阿里 / 百度 / 夸克 / 迅雷等）

- **URL**：`PUT /api/v1/system/netdisk-credentials`
- **认证**：管理员 Token
- **说明**：统一维护各网盘的 Cookie / refresh_token / 目标目录等参数，保存后会同步到 `system_configs`。
- **请求体示例（只列部分字段）**：

```json
{
  "quark_cookie": "QUARK_SESS=...",
  "quark_auto_save": true,
  "quark_target_folder_id": "0",
  "baidu_cookie": "BDUSS=...; STOKEN=...; BDCLND=...",
  "baidu_auto_save": true,
  "baidu_target_path": "/DFAN/转存",
  "aliyun_refresh_token": "xxxxx",
  "aliyun_auto_save": true,
  "aliyun_target_parent_file_id": "root",
  "xunlei_cookie": "a1.xxxxx",
  "xunlei_auto_save": true,
  "xunlei_target_folder_id": "0",
  "replace_link_after_transfer": true
}
```

> 建议先用管理后台「网盘凭证」页面填一遍，再通过 `GET /api/v1/system/netdisk-credentials` 查看完整字段结构。

#### 3.4.2 手动重试单个资源转存

- **URL**：`POST /api/v1/admin/resources/:id/retry-transfer`
- **认证**：管理员 Token
- **说明**：对指定资源 ID（主链接 + extra_links）按当前开启的自动转存配置重试一次；常用于转存失败后的人工重试。

#### 3.4.3 通过链接直接发起转存

- **URL**：`POST /api/v1/netdisk/transfer`
- **认证**：管理员 Token
- **说明**：传入单条分享链接和可选提取码，后端自动识别平台并调用对应网盘的转存流程。

请求示例：

```json
{
  "link": "https://pan.baidu.com/s/xxxx?pwd=0202",
  "password": "0202",
  "platform": "auto"
}
```

响应示例：

```json
{
  "code": 0,
  "message": "转存完成",
  "data": {
    "platform": "baidu",
    "old_link": "https://pan.baidu.com/s/xxxx?pwd=0202",
    "new_link": "https://pan.baidu.com/s/本人分享链接",
    "own_share_url": "https://pan.baidu.com/s/本人分享链接"
  }
}
```

#### 3.4.4 批量链接转存

- **URL**：`POST /api/v1/netdisk/transfer/batch`
- **认证**：管理员 Token
- **说明**：一次提交多条链接进行转存，适合表格导入或脚本调用。

```json
{
  "links": [
    "https://pan.quark.cn/s/xxxx",
    "https://www.alipan.com/s/xxxx",
    "https://pan.xunlei.com/s/xxxx"
  ]
}
```


## Open Netdisk API

These endpoints are public and return JSON for third-party callers.

### GET /api/v1/open/netdisk/resources

Query params:
- `page`: default `1`
- `page_size`: default `20`, max `100`
- `q`: keyword, matches title/description/tags
- `category_id`: category id
- `platform`: `baidu` / `aliyun` / `quark` / `xunlei` / `uc` / `tianyi` / `yidong` / `115` / `123pan` / `other`
- `link_valid`: `true` / `false` / `1` / `0`
- `sort`: `latest` / `hot`

Example:
```bash
curl "http://localhost:8080/api/v1/open/netdisk/resources?page=1&page_size=10&platform=quark"
```

### GET /api/v1/open/netdisk/resources/:id

Returns one public resource record with the latest transfer summary when available.

Example:
```bash
curl "http://localhost:8080/api/v1/open/netdisk/resources/12"
```

> 提示：后端还集成了 Swagger 文档，开发调试时可通过 `/swagger/index.html` 以可视化方式浏览、试调所有 API。
