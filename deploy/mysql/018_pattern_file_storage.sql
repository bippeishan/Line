-- 018_pattern_file_storage.sql
-- 图解文件不再以 base64 写入 pattern_files.file_data，改为对象存储 + 元数据。
-- 存量行保留 file_data，由 API 在首次读取时惰性迁移（迁移后可手工把该列置空）。
-- 执行前请备份数据库。

ALTER TABLE pattern_files
  MODIFY COLUMN file_data MEDIUMTEXT NULL,
  ADD COLUMN object_key VARCHAR(255) NULL AFTER file_type,
  ADD COLUMN mime_type VARCHAR(80) NOT NULL DEFAULT 'application/octet-stream' AFTER object_key,
  ADD COLUMN size_bytes BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER mime_type;

CREATE INDEX idx_pattern_files_user_created ON pattern_files (user_id, created_at);

-- 校验：object_key 已填写即表示文件已出库；file_data 为空且 object_key 为空的记录需要重新上传。
SELECT
  SUM(object_key IS NOT NULL) AS migrated_rows,
  SUM(object_key IS NULL AND file_data IS NOT NULL) AS pending_rows,
  SUM(object_key IS NULL AND file_data IS NULL) AS broken_rows
FROM pattern_files;
