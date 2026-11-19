#!/usr/bin/env bash
# 在干净的 MySQL 上按编号顺序执行 deploy/mysql 的全部迁移，并校验最终结构。
# 脚本会新建一个临时库（默认 line_verify）跑完整流程，不影响已有数据。
#
# 用法：
#   ./scripts/verify-migrations.sh
#   MYSQL_HOST=127.0.0.1 MYSQL_PORT=3306 ./scripts/verify-migrations.sh
#
# 环境变量：MYSQL_HOST / MYSQL_PORT / MYSQL_USER / MYSQL_PASSWORD / MYSQL_VERIFY_DATABASE / MYSQL_WAIT_SECONDS
#          MYSQL_CLIENT=auto|cli|docker  客户端选择，默认 auto（优先本机 mysql，回退 mysql:8.4 容器）
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MIGRATION_DIR="$ROOT_DIR/deploy/mysql"
MYSQL_HOST="${MYSQL_HOST:-127.0.0.1}"
MYSQL_PORT="${MYSQL_PORT:-3306}"
MYSQL_USER="${MYSQL_USER:-root}"
MYSQL_PASSWORD="${MYSQL_PASSWORD:-root_secret}"
MYSQL_VERIFY_DATABASE="${MYSQL_VERIFY_DATABASE:-line_verify}"
MYSQL_WAIT_SECONDS="${MYSQL_WAIT_SECONDS:-60}"

MYSQL_CLIENT_MODE="${MYSQL_CLIENT:-auto}"
mysql_client=()
if [ "$MYSQL_CLIENT_MODE" != "docker" ] && command -v mysql >/dev/null 2>&1; then
  mysql_client=(mysql --protocol=TCP -h"$MYSQL_HOST" -P"$MYSQL_PORT" -u"$MYSQL_USER")
elif command -v docker >/dev/null 2>&1; then
  # 容器里的 127.0.0.1 指向容器自身，换成 host-gateway 才能连到宿主上映射出来的 MySQL。
  case "$MYSQL_HOST" in
    127.0.0.1|localhost|::1) docker_host="host.docker.internal" ;;
    *) docker_host="$MYSQL_HOST" ;;
  esac
  mysql_client=(docker run --rm -i -e MYSQL_PWD --add-host=host.docker.internal:host-gateway mysql:8.4 \
    mysql --protocol=TCP -h"$docker_host" -P"$MYSQL_PORT" -u"$MYSQL_USER")
else
  echo "需要 mysql 客户端或 docker 才能验证迁移" >&2
  exit 1
fi

export MYSQL_PWD="$MYSQL_PASSWORD"
run_server_sql() { "${mysql_client[@]}" -N -B "$@"; }
run_sql() { "${mysql_client[@]}" -N -B -D "$MYSQL_VERIFY_DATABASE" "$@"; }

echo "等待 MySQL ${MYSQL_HOST}:${MYSQL_PORT} 就绪…"
deadline=$((SECONDS + MYSQL_WAIT_SECONDS))
until run_server_sql -e 'SELECT 1' >/dev/null 2>&1; do
  if [ "$SECONDS" -ge "$deadline" ]; then
    echo "MySQL 在 ${MYSQL_WAIT_SECONDS}s 内未就绪" >&2
    exit 1
  fi
  sleep 2
done

echo "重建验证库 ${MYSQL_VERIFY_DATABASE}："
run_server_sql -e "DROP DATABASE IF EXISTS \`$MYSQL_VERIFY_DATABASE\`"
run_server_sql -e "CREATE DATABASE \`$MYSQL_VERIFY_DATABASE\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"
echo "  ✓ 空库已就绪"

echo "按顺序执行迁移："
for file in $(ls "$MIGRATION_DIR"/*.sql | sort); do
  name="$(basename "$file")"
  if ! run_sql < "$file" >/dev/null; then
    echo "迁移失败: $name" >&2
    exit 1
  fi
  echo "  ✓ $name"
done

# 迁移脚本必须可重复执行：守卫写法（information_schema / IF NOT EXISTS）覆盖的编号再跑一次。
echo "重跑带守卫的迁移（011–013、019、020）："
for name in 011_shopping_list.sql 012_project_yarn_plans.sql 013_project_schedule.sql 019_reminder_delivery.sql 020_project_defaults.sql; do
  if ! run_sql < "$MIGRATION_DIR/$name" >/dev/null; then
    echo "迁移不可重复执行: $name" >&2
    exit 1
  fi
  echo "  ✓ $name"
done

echo "校验最终结构："
failures=0
expect() {
  local actual
  actual="$(run_sql -e "$1")"
  if [ "$actual" != "$2" ]; then
    echo "  ✗ $3（期望 $2，实际 ${actual:-空}）" >&2
    failures=$((failures + 1))
  else
    echo "  ✓ $3"
  fi
}
expect "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='$MYSQL_VERIFY_DATABASE'" 25 "25 张表"
for column in pattern_files.object_key pattern_files.mime_type pattern_files.size_bytes pattern_annotations.row_number \
  projects.pattern_file_id projects.completion_photo_data projects.palette projects.work_tags projects.share_token \
  projects.target_date projects.weekly_focus_minutes projects.version users.email_verified_at \
  yarns.color_hex yarns.remaining_grams project_reminders.remind_at project_reminders.channel \
  project_reminders.email_sent_at project_reminders.attempts project_reminders.last_error; do
  table="${column%%.*}"
  column="${column##*.}"
  expect "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema='$MYSQL_VERIFY_DATABASE' AND table_name='$table' AND column_name='$column'" 1 "$table.$column"
done
for table in media_assets audit_events user_permissions stitch_terms user_stitch_favorites calculation_records \
  project_timers project_reminders project_yarn_plans project_accessories shopping_items pattern_annotations; do
  expect "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='$MYSQL_VERIFY_DATABASE' AND table_name='$table'" 1 "表 $table"
done
expect "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema='$MYSQL_VERIFY_DATABASE' AND table_name='pattern_files' AND column_name='file_data' AND is_nullable='YES'" 1 "pattern_files.file_data 允许为空"
expect "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema='$MYSQL_VERIFY_DATABASE' AND table_name='projects' AND column_name='completion_note' AND column_default IS NOT NULL" 1 "projects.completion_note 有默认值"

# 冒烟：创建项目不能依赖应用层补字段（历史上 1364 就出在这里）。
run_sql -e "INSERT INTO users (name,email,role) VALUES ('迁移校验','verify@line.local','user') ON DUPLICATE KEY UPDATE name=VALUES(name)" >/dev/null
run_sql -e "INSERT INTO projects (user_id,name) SELECT id,'迁移冒烟' FROM users WHERE email='verify@line.local'" >/dev/null
expect "SELECT COUNT(*) FROM projects WHERE name='迁移冒烟'" 1 "项目插入冒烟"

if [ "$failures" -ne 0 ]; then
  echo "结构校验失败 $failures 项" >&2
  exit 1
fi
echo "迁移验证通过：001–020 可全量执行且可重复。"
