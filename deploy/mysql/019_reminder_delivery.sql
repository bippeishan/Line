-- 提醒触达：站内提醒之外增加邮件通道与投递状态。
-- 使用 information_schema 守卫，保证在已手工加过列的环境里可重复执行。
SET @has_remind_at := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'project_reminders' AND COLUMN_NAME = 'remind_at'
);
SET @sql := IF(@has_remind_at = 0,
  'ALTER TABLE project_reminders ADD COLUMN remind_at DATETIME NULL AFTER is_done',
  'DO 0');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_channel := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'project_reminders' AND COLUMN_NAME = 'channel'
);
SET @sql := IF(@has_channel = 0,
  "ALTER TABLE project_reminders ADD COLUMN channel ENUM('in_app','email') NOT NULL DEFAULT 'in_app' AFTER remind_at",
  'DO 0');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_sent_at := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'project_reminders' AND COLUMN_NAME = 'email_sent_at'
);
SET @sql := IF(@has_sent_at = 0,
  'ALTER TABLE project_reminders ADD COLUMN email_sent_at TIMESTAMP NULL AFTER channel',
  'DO 0');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_attempts := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'project_reminders' AND COLUMN_NAME = 'attempts'
);
SET @sql := IF(@has_attempts = 0,
  'ALTER TABLE project_reminders ADD COLUMN attempts INT UNSIGNED NOT NULL DEFAULT 0 AFTER email_sent_at',
  'DO 0');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_last_error := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'project_reminders' AND COLUMN_NAME = 'last_error'
);
SET @sql := IF(@has_last_error = 0,
  'ALTER TABLE project_reminders ADD COLUMN last_error VARCHAR(255) NULL AFTER attempts',
  'DO 0');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_due_index := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'project_reminders' AND INDEX_NAME = 'idx_project_reminders_due'
);
SET @sql := IF(@has_due_index = 0,
  'CREATE INDEX idx_project_reminders_due ON project_reminders (channel, email_sent_at, remind_at)',
  'DO 0');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 校验：待发送的邮件提醒数量。
SELECT COUNT(*) AS pending_email_reminders
FROM project_reminders
WHERE channel = 'email' AND email_sent_at IS NULL AND is_done = FALSE;
