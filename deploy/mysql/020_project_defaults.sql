-- 006 把 completion_note 声明为 TEXT NOT NULL 但没有默认值，
-- 导致 createProjectHandler 的 INSERT 在干净库上直接失败（1364）。
-- 这里给 projects 上所有 NOT NULL 且无默认值的文本/JSON 列补默认值，
-- 让任何遗漏字段的 INSERT 都不会在运行时炸掉。

SET @has_default := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'projects' AND COLUMN_NAME = 'completion_note'
    AND COLUMN_DEFAULT IS NOT NULL
);
SET @sql := IF(@has_default = 0,
  "ALTER TABLE projects MODIFY completion_note TEXT NOT NULL DEFAULT ('')",
  'DO 0');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_default := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'projects' AND COLUMN_NAME = 'breakpoint_note'
    AND COLUMN_DEFAULT IS NOT NULL
);
SET @sql := IF(@has_default = 0,
  "ALTER TABLE projects MODIFY breakpoint_note TEXT NOT NULL DEFAULT ('')",
  'DO 0');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_default := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'projects' AND COLUMN_NAME = 'yarn_names'
    AND COLUMN_DEFAULT IS NOT NULL
);
SET @sql := IF(@has_default = 0,
  "ALTER TABLE projects MODIFY yarn_names JSON NOT NULL DEFAULT ('[]')",
  'DO 0');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_default := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'projects' AND COLUMN_NAME = 'tool_names'
    AND COLUMN_DEFAULT IS NOT NULL
);
SET @sql := IF(@has_default = 0,
  "ALTER TABLE projects MODIFY tool_names JSON NOT NULL DEFAULT ('[]')",
  'DO 0');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_default := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'projects' AND COLUMN_NAME = 'palette'
    AND COLUMN_DEFAULT IS NOT NULL
);
SET @sql := IF(@has_default = 0,
  "ALTER TABLE projects MODIFY palette JSON NOT NULL DEFAULT ('[]')",
  'DO 0');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_default := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'projects' AND COLUMN_NAME = 'work_tags'
    AND COLUMN_DEFAULT IS NOT NULL
);
SET @sql := IF(@has_default = 0,
  "ALTER TABLE projects MODIFY work_tags JSON NOT NULL DEFAULT ('[]')",
  'DO 0');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 校验：projects 上不应再有 NOT NULL 且无默认值的文本/JSON 列。
SELECT column_name
FROM information_schema.COLUMNS
WHERE table_schema = DATABASE() AND table_name = 'projects'
  AND is_nullable = 'NO' AND column_default IS NULL
  AND data_type IN ('text', 'mediumtext', 'longtext', 'json', 'varchar', 'char');
