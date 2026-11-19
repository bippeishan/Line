ALTER TABLE pattern_annotations
  ADD COLUMN row_number INT UNSIGNED NOT NULL DEFAULT 0 AFTER page_number;
