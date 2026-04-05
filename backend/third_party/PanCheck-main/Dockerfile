# 多阶段构建 Dockerfile

# ====================================================================
# 第一阶段：构建 React 前端
# ====================================================================
FROM node:18-alpine AS frontend-builder

# 安装 pnpm
RUN npm install -g pnpm

# 设置工作目录
WORKDIR /app/frontend

# 复制前端依赖文件
COPY frontend/package.json frontend/pnpm-lock.yaml ./

# 安装前端依赖
RUN pnpm install --frozen-lockfile

# 复制前端源代码
COPY frontend/ ./

# 构建前端应用
RUN pnpm run build

# ====================================================================
# 第二阶段：构建 Go 后端
# ====================================================================
FROM golang:1.24-alpine AS backend-builder

# 安装构建工具
RUN apk add --no-cache gcc g++ musl-dev

# 设置工作目录
WORKDIR /app/backend

# 复制 Go 依赖文件
COPY go.mod go.sum ./

# 下载 Go 依赖
RUN go mod download

# 复制后端源代码
COPY . .

# 从前端构建阶段复制编译后的文件到后端 static 目录
COPY --from=frontend-builder /app/frontend/dist ./static

# 构建 Go 应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# ====================================================================
# 第三阶段：最终运行镜像
# ====================================================================
FROM alpine:latest

# 安装必要的包
RUN apk --no-cache add ca-certificates tzdata libstdc++ libgcc

# 设置时区
ENV TZ=Asia/Shanghai

# 设置工作目录
WORKDIR /app

# 从后端构建阶段复制二进制文件
COPY --from=backend-builder /app/backend/main .

# 从后端构建阶段复制静态文件
COPY --from=backend-builder /app/backend/static ./static

# 创建数据目录
RUN mkdir -p /app/data

# 暴露端口
EXPOSE 6080

# 启动应用
CMD ["./main"]

