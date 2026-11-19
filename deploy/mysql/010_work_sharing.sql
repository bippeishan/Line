ALTER TABLE projects
  ADD COLUMN work_tags JSON NULL AFTER palette,
  ADD COLUMN share_token CHAR(32) NULL AFTER actual_minutes,
  ADD COLUMN share_show_materials BOOLEAN NOT NULL DEFAULT TRUE AFTER share_token,
  ADD COLUMN share_show_note BOOLEAN NOT NULL DEFAULT TRUE AFTER share_show_materials,
  ADD UNIQUE KEY uk_projects_share_token (share_token);

UPDATE projects SET work_tags = JSON_ARRAY() WHERE work_tags IS NULL;

ALTER TABLE projects
  MODIFY COLUMN work_tags JSON NOT NULL;
