-- 排期字段：早期初始化脚本创建的项目表已包含这两列，
-- 用 information_schema 守卫保证本迁移可重复执行（MySQL 不支持 ADD COLUMN IF NOT EXISTS）。
SET @has_target_date := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'projects' AND COLUMN_NAME = 'target_date'
);
SET @sql := IF(@has_target_date = 0,
  'ALTER TABLE projects ADD COLUMN target_date DATE NULL AFTER completed_at',
  'DO 0');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_weekly_focus := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'projects' AND COLUMN_NAME = 'weekly_focus_minutes'
);
SET @sql := IF(@has_weekly_focus = 0,
  'ALTER TABLE projects ADD COLUMN weekly_focus_minutes INT UNSIGNED NOT NULL DEFAULT 0 AFTER target_date',
  'DO 0');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
