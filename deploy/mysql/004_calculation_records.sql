CREATE TABLE calculation_records (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  project_id BIGINT UNSIGNED NULL,
  mode ENUM('circle', 'size', 'yarn') NOT NULL,
  input_data JSON NOT NULL,
  result_data JSON NOT NULL,
  formula VARCHAR(500) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  INDEX idx_calculation_records_user_created (user_id, created_at),
  INDEX idx_calculation_records_project_created (project_id, created_at),
  CONSTRAINT fk_calculation_records_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_calculation_records_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
