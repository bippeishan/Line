# 线禾 Line

一个面向钩织爱好者的作品与图解管理应用。仓库包含 Go API、React + TypeScript Web 应用和 MySQL 开发环境。

## 目录

```text
backend/        Go HTTP API
frontend/       React + TypeScript + Vite 应用
deploy/         MySQL 初始化迁移
docker-compose.yml
```

## 本地运行

1. 启动数据库：`docker compose up -d mysql`
2. 复制环境变量：`cp backend/.env.example backend/.env`
3. 启动 API：`cd backend && go run ./cmd/api`
4. 新开终端启动前端：`cd frontend && npm install && npm run dev`

打开 `http://localhost:5173`。前端会将 `/api` 请求代理至 `http://localhost:8080`。

## 常用命令

```bash
docker compose up -d              # 启动本地 MySQL 与 API
cd backend && go test ./...       # 后端测试
cd frontend && npm run build      # 前端类型检查和构建
```
