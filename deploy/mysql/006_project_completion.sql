ALTER TABLE projects
  ADD COLUMN completion_photo_data MEDIUMTEXT NULL AFTER cover_data,
  ADD COLUMN completion_note TEXT NOT NULL AFTER breakpoint_note,
  ADD COLUMN actual_minutes INT UNSIGNED NOT NULL DEFAULT 0 AFTER completed_at;
