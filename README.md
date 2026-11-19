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

新数据库使用 `deploy/mysql/001_init.sql` 作为基线，随后按编号顺序执行 `002`–`021` 迁移。`docker-compose.yml` 挂载了整个目录，MySQL 首次启动会按文件名顺序自动执行，因此基线必须保持 `001_` 前缀。已有数据库请先备份，再按编号手动执行尚未应用的迁移脚本。

在 MySQL 容器中可按如下方式执行单个迁移：

```bash
docker compose exec -T mysql mysql -u root -p"$MYSQL_ROOT_PASSWORD" line < deploy/mysql/021_palette_templates.sql
```

当前最新迁移为 `021_palette_templates.sql`，它建立配色模板库并写入 10 组内置方案（`owner_key` 用 VIRTUAL 生成列保证内置名唯一、用户自建名各自唯一）。`020_project_defaults.sql` 给 `projects` 上 NOT NULL 且无默认值的文本/JSON 列补默认值（此前干净库上创建项目会报 1364）。`019_reminder_delivery.sql` 增加提醒的发送时间、渠道与投递状态；`018_pattern_file_storage.sql` 把图解文件移出数据库，`pattern_files` 只保留 `object_key`、`mime_type`、`size_bytes` 等元数据。执行后建议确认表和字段已就绪：

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
SHOW COLUMNS FROM pattern_files LIKE 'object_key';
SHOW COLUMNS FROM project_reminders LIKE 'channel';
SHOW COLUMNS FROM projects LIKE 'completion_note';
SHOW TABLES LIKE 'palette_templates';
```

## 对象存储

默认使用本地目录 `MEDIA_ROOT`。生产环境可设置 `MEDIA_DRIVER=s3` 并配置 `S3_BUCKET`、`S3_REGION`、`S3_ACCESS_KEY`、`S3_SECRET_KEY`；对于 MinIO 或 S3 兼容 OSS 网关，再设置 `S3_ENDPOINT` 与 `S3_FORCE_PATH_STYLE=true`。媒体仍通过应用的鉴权接口读取，不直接暴露存储凭据或私有对象 URL。

项目封面与完成照走 `POST /api/media/images`（8 MB 上限、统一压缩为 JPEG、每账号 100 MB 配额）。图解文件走 `POST /api/pattern-files`（PDF/JPEG/PNG/GIF/WebP，12 MB 上传、20 MB 解码上限、每账号 200 MB 配额），写入同一对象存储后由 `GET /api/pattern-files/{id}/file` 鉴权回传，数据库只保存元数据。两类文件在存储中的键分别是 `{id}.jpg` 与 `pattern-{id}.{ext}`。

## 配色模板

`GET /api/palette-templates` 返回官方内置模板（`is_builtin=1`，`user_id` 为空）与当前用户自建模板；`POST` 收藏一组配色，`DELETE /api/palette-templates/{id}` 只能删自建模板。颜色统一为 `#RRGGBB` 大写，单组 2–6 个且不得重复；同名（同一用户或同名内置）返回 409。

前端 `配色推荐` 页顶部是模板库，点卡片直接套用模板颜色；改基准色或切换配色规则会自动切回算法生成。内置 10 组覆盖玩偶、花束、服饰、居家与挂饰。

## 邮件与提醒

提醒支持两种渠道：`in_app`（仅站内列表）与 `email`（到点发邮件）。配置 `SMTP_HOST` 后，后台调度器每分钟扫描到期提醒并投递：

- 只发给已验证邮箱（`users.email_verified_at` 非空）且项目仍在进行中的提醒；
- 每次最多 50 封，单条最多重试 5 次，失败原因写入 `last_error` 并在项目活动流与审计里留痕；
- 未配置 `SMTP_HOST` 时调度器不启动，避免把提醒误标为已发送；
- `POST /api/projects/{id}/reminders/{reminderID}/send` 可立即发送单条提醒，用于确认邮件通道。

`SMTP_FROM` 可以写 `Line <no-reply@example.com>` 形式：信封发件人会取尖括号内的裸地址，`From` 头保留显示名。本地想直接看邮件，可运行 `docker run --rm -p 1025:1025 -p 8025:8025 axllent/mailpit:v1.21`，再设 `SMTP_HOST=127.0.0.1 SMTP_PORT=1025`，收件箱在 `http://localhost:8025`。

## 常用命令

```bash
./scripts/verify-migrations.sh    # 在临时库 line_verify 上全量跑 001–021 并校验结构
                                  # 无 mysql 客户端时自动用 mysql:8.4 容器（MYSQL_CLIENT=cli|docker 可强制）
docker compose up -d              # 启动本地 MySQL 与 API
cd backend && go test -race ./... # 后端测试
cd frontend && npm run lint       # 前端 ESLint
cd frontend && npm run build      # 前端类型检查和构建
cd frontend && npm run test:e2e   # Playwright 端到端（需 API 与 MySQL 已启动；设 MAILPIT_URL 可验证邮件投递）
```

## 持续集成

`.github/workflows/ci.yml` 在推送到 `main`、PR 和手动触发时运行四个任务：

| 任务 | 内容 |
| --- | --- |
| 后端 | `gofmt` 校验、`go vet`、`go build`、`go test -race` |
| 前端 | `npm ci`、`npm run lint`、`tsc -b`、`vite build` |
| 迁移 | 临时 MySQL 8.4 容器上执行 `001`–`021`，重跑带守卫的迁移，断言 26 张表、关键字段、10 组内置配色与项目插入冒烟 |
| 端到端 | 准备数据库、启动 API（含 Mailpit SMTP）与前端，运行 Playwright（登录、图解上传与鉴权回传、提醒邮件投递、配色模板套用与收藏） |

在合入前本地复现同一套检查：

```bash
./scripts/verify-migrations.sh
(cd backend && gofmt -l . && go vet ./... && go test -race ./...)
(cd frontend && npm ci && npm run lint && npm run build)
```

迁移脚本必须可重复执行：新增列用 `information_schema` 守卫，新增表用 `CREATE TABLE IF NOT EXISTS`，否则迁移任务会直接失败。
