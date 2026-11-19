ALTER TABLE projects
  ADD COLUMN palette JSON NULL AFTER tool_names;

UPDATE projects SET palette = JSON_ARRAY() WHERE palette IS NULL;

ALTER TABLE projects
  MODIFY COLUMN palette JSON NOT NULL;
