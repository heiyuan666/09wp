## Meilisearch Docker 安装与运行（本项目用）

参考文档：[使用 Docker 运行 Meilisearch](https://meilisearch.com.cn/docs/guides/docker)

### 1) 前置条件

- 已安装 Docker / Docker Desktop
- 本项目后端配置里需要填写 Meili 地址与 Key（如有）

### 2) 拉取镜像

官方不建议长期使用 `latest`（不同机器可能拉到不同版本）。建议固定一个版本号，例如 `v1.16`。

```bash
docker pull getmeili/meilisearch:v1.16
```

### 3) 启动（开发环境，数据持久化到本机目录）

在你希望保存 Meili 数据的目录执行（会生成 `meili_data/`）：

```bash
docker run -it --rm ^
  -p 7700:7700 ^
  -v %cd%/meili_data:/meili_data ^
  getmeili/meilisearch:v1.16
```

说明：

- **端口**：容器 `7700` 映射到宿主机 `7700`，访问 `http://127.0.0.1:7700`
- **数据文件**：容器内 `/meili_data/data.ms`，通过 `-v` 映射到宿主机目录实现持久化

### 4) 启动（带主密钥 / API Key）

如果你要启用鉴权（推荐），用环境变量设置 master key：

```bash
docker run -it --rm ^
  -p 7700:7700 ^
  -e MEILI_MASTER_KEY="MASTER_KEY" ^
  -v %cd%/meili_data:/meili_data ^
  getmeili/meilisearch:v1.16
```

### 5) 本项目后端配置示例

在后台「系统配置」里开启 Meilisearch，填写：

- **Meili URL**：`http://127.0.0.1:7700`
- **Meili API Key**：填你在 `MEILI_MASTER_KEY` 里设置的 key（或留空表示无鉴权）
- **Meili Index**：默认 `resources`

保存后：

1. 点击「测试连接」
2. 首次使用点击「重建索引」（把 MySQL 里的 `resources` 全量导入 Meili）

### 6) 常用维护：导出/导入 dump（可选）

导入 dump 需要用 CLI 参数启动（注意 dump 路径要在挂载卷中）：

```bash
docker run -it --rm ^
  -p 7700:7700 ^
  -v %cd%/meili_data:/meili_data ^
  getmeili/meilisearch:v1.16 ^
  meilisearch --import-dump /meili_data/dumps/20200813-042312213.dump
```

> 如果你使用了持久化目录（`-v`），在导入 dump 前需要先删除卷内的旧库文件 `/meili_data/data.ms`（按官方文档要求）。

### 7) 常用维护：快照（可选）

创建快照：

```bash
docker run -it --rm ^
  -p 7700:7700 ^
  -v %cd%/meili_data:/meili_data ^
  getmeili/meilisearch:v1.16 ^
  meilisearch --schedule-snapshot --snapshot-dir /meili_data/snapshots
```

导入快照：

```bash
docker run -it --rm ^
  -p 7700:7700 ^
  -v %cd%/meili_data:/meili_data ^
  getmeili/meilisearch:v1.16 ^
  meilisearch --import-snapshot /meili_data/snapshots/data.ms.snapshot
```

