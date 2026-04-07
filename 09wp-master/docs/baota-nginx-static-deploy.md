# 宝塔：静态前端 + 后端接口直通（Nginx）

适用于 **Vite 构建产物**（`dist/`）与 **Go 后端**（本机如 `127.0.0.1:8080`，路由前缀 `/api/v1`）同域部署。

---

## 一、部署前准备

| 步骤 | 说明 |
|------|------|
| 前端构建 | 项目根目录执行 `pnpm run build` 或 `npm run build`，得到 `dist/` |
| 后端二进制 | `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o server ./cmd/server`，上传服务器 |
| 配置文件 | 将 `backend/config.json` 放到与 `server` 同级或按你启动方式指定路径 |
| 运行后端 | 监听 `8080`（或与下文 `proxy_pass` 端口一致） |

---

## 二、宝塔站点设置

1. **网站** → **添加站点** → 域名填写你的域名。  
2. **根目录** 指向前端静态目录，例如：`/www/wwwroot/你的域名/dist`（把 `dist` 内文件放根目录也可，则根目录为 `.../dist` 且内部为 `index.html`、`assets/`）。  
3. **PHP 版本** 选「纯静态」或关闭 PHP（按宝塔版本界面为准）。  
4. **SSL**：按需申请证书并开启强制 HTTPS。

---

## 三、Nginx 配置（整段可放入「网站 → 设置 → 配置文件」的 `server { }` 内）

按宝塔默认结构，在 `server { ... }` 里保留 `listen`、`server_name`、`ssl` 等，**用下面逻辑替换或合并 `location` 段**（注意与现有 `location ~ \.php` 等冲突时以静态+反代为主）。

```nginx
# 站点根目录：指向前端 dist（按实际路径修改）
# root /www/wwwroot/example.com/dist;
# index index.html;

# 静态资源后缀：直出文件，不走 SPA 回退
location ~* \.(js|css|png|jpg|jpeg|gif|svg|ico|webp|json|txt|woff|woff2|ttf|eot)$ {
    try_files $uri =404;
    expires 30d;
    access_log off;
}

# API：整段前缀交给后端（Go 一般为 /api/v1/...）
location ^~ /api/ {
    try_files $uri @backend;
}

# Sitemap / RSS：由后端提供
location = /sitemap.xml {
    try_files $uri @backend;
}

location = /rss.xml {
    try_files $uri @backend;
}

# 后端托管的公开静态目录（可选：若不放 nginx 直出，可反代给 Go）
# location ^~ /public/ {
#     try_files $uri @backend;
# }

# 单页应用：其余路径回退 index.html
location / {
    try_files $uri $uri/ /index.html;
}

# 反代到本机 Go 服务（端口与 config.json / 启动参数一致）
location @backend {
    proxy_pass http://127.0.0.1:8080;
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_connect_timeout 60s;
    proxy_send_timeout 60s;
    proxy_read_timeout 60s;
}
```

**说明：**

- `location ^~ /api/` 会匹配所有以 `/api/` 开头的请求（含 `/api/v1/...`），由 `@backend` 转发到 Go。  
- 前端 `VITE_API_BASE_URL` 一般为同源 `/api/v1` 或完整域名 `https://你的域名/api/v1`。  
- `try_files $uri @backend` 在静态目录没有同名文件时进入反代；若你希望 **所有** `/api/` 不查磁盘，可改为：

```nginx
location ^~ /api/ {
    proxy_pass http://127.0.0.1:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

（与 `@backend` 二选一即可，避免重复。）

---

## 四、宝塔里「改静态」的常见操作

| 项目 | 操作 |
|------|------|
| 站点类型 | 不使用 PHP 运行时，根目录只放 `dist` 构建结果 |
| 伪静态 | 选择「不配置」或删除会拦截 `.html` 的规则；SPA 依赖上面 `location /` 的 `try_files` |
| 防跨域 / 缓存 | 静态 `expires` 可按需调整 |
| 保存重载 | 修改配置后 **保存** 并在宝塔中 **重载 Nginx** |

---

## 五、自检清单

- [ ] 浏览器打开 `https://你的域名/` 能出首页  
- [ ] `https://你的域名/api/v1/health` 或实际健康检查接口返回正常  
- [ ] `https://你的域名/sitemap.xml`、`/rss.xml` 能访问（若后端已启用）  
- [ ] 强制刷新（Ctrl+F5）后 JS/CSS 仍为 200，且缓存头符合预期  

---

## 六、导出 / 备份建议

- 宝塔 **网站 → 设置 → 配置文件** 可复制整段 `server` 块到文本备份。  
- 或使用宝塔 **计划任务 / 备份** 备份网站目录与 Nginx 配置。  

本文路径（仓库内）：`docs/baota-nginx-static-deploy.md`。
