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

## 开发环境登录

首次启动 API 会为初始化管理员设置密码。默认账号为 `admin@line.local`，密码读取 `BOOTSTRAP_ADMIN_PASSWORD`（示例值为 `line_admin_2026`）。生产环境请通过安全的环境变量管理方式覆盖该值。

## 数据库迁移

新数据库使用 `deploy/mysql/init.sql` 初始化。已有数据库请先备份，再按编号手动执行尚未应用的迁移脚本；当前最新为 `013_project_schedule.sql`。这些脚本均为一次性迁移，必须按编号顺序执行。

在 MySQL 容器中可按如下方式执行单个迁移：

```bash
docker compose exec -T mysql mysql -u root -p"$MYSQL_ROOT_PASSWORD" line < deploy/mysql/013_project_schedule.sql
```

当前最新迁移为 `017_fine_grained_permissions.sql`，它增加邮箱验证、密码重置令牌、对象图片元数据、项目版本控制、审计记录和细粒度权限覆盖。执行后建议确认表和字段已就绪：

```sql
SHOW TABLES LIKE 'shopping_items';
SHOW TABLES LIKE 'project_yarn_plans';
SHOW COLUMNS FROM projects LIKE 'target_date';
SHOW COLUMNS FROM projects LIKE 'weekly_focus_minutes';
SHOW COLUMNS FROM users LIKE 'email_verified_at';
SHOW TABLES LIKE '%verification_tokens';
SHOW TABLES LIKE '%reset_tokens';
SHOW TABLES LIKE 'media_assets';
SHOW COLUMNS FROM projects LIKE 'version';
SHOW TABLES LIKE 'audit_events';
SHOW TABLES LIKE 'user_permissions';
```

## 对象存储

默认使用本地目录 `MEDIA_ROOT`。生产环境可设置 `MEDIA_DRIVER=s3` 并配置 `S3_BUCKET`、`S3_REGION`、`S3_ACCESS_KEY`、`S3_SECRET_KEY`；对于 MinIO 或 S3 兼容 OSS 网关，再设置 `S3_ENDPOINT` 与 `S3_FORCE_PATH_STYLE=true`。媒体仍通过应用的鉴权接口读取，不直接暴露存储凭据或私有对象 URL。

## 常用命令

```bash
docker compose up -d              # 启动本地 MySQL 与 API
cd backend && go test ./...       # 后端测试
cd frontend && npm run build      # 前端类型检查和构建
```
