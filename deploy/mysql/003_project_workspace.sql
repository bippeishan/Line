ALTER TABLE projects
  MODIFY COLUMN status ENUM('planning', 'in_progress', 'completed', 'archived') NOT NULL DEFAULT 'planning',
  ADD COLUMN pattern_file_id BIGINT UNSIGNED NULL AFTER pattern_name,
  ADD CONSTRAINT fk_projects_pattern_file FOREIGN KEY (pattern_file_id) REFERENCES pattern_files(id) ON DELETE SET NULL;

CREATE TABLE project_activities (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  project_id BIGINT UNSIGNED NOT NULL,
  kind VARCHAR(40) NOT NULL,
  detail VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  INDEX idx_project_activities_project (project_id, created_at),
  CONSTRAINT fk_project_activities_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
