-- row_number 是 MySQL 8 的保留函数名，必须使用反引号。
ALTER TABLE pattern_annotations
  ADD COLUMN `row_number` INT UNSIGNED NOT NULL DEFAULT 0 AFTER page_number;
