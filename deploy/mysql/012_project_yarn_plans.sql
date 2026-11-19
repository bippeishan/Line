CREATE TABLE IF NOT EXISTS project_yarn_plans (
  project_id BIGINT UNSIGNED NOT NULL,
  yarn_id BIGINT UNSIGNED NOT NULL,
  planned_grams DECIMAL(10,2) NOT NULL,
  PRIMARY KEY (project_id, yarn_id),
  CONSTRAINT fk_project_yarn_plans_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
  CONSTRAINT fk_project_yarn_plans_yarn FOREIGN KEY (yarn_id) REFERENCES yarns(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
