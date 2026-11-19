ALTER TABLE projects
  ADD COLUMN target_date DATE NULL AFTER completed_at,
  ADD COLUMN weekly_focus_minutes INT UNSIGNED NOT NULL DEFAULT 0 AFTER target_date;
