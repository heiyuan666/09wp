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
