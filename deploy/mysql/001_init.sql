CREATE TABLE IF NOT EXISTS patterns (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  title VARCHAR(120) NOT NULL,
  category VARCHAR(60) NOT NULL,
  difficulty ENUM('beginner', 'intermediate', 'advanced') NOT NULL DEFAULT 'beginner',
  yarn VARCHAR(120) NOT NULL,
  hook_size VARCHAR(30) NOT NULL,
  duration_minutes INT UNSIGNED NOT NULL DEFAULT 0,
  cover_url VARCHAR(500) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  INDEX idx_patterns_category (category)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(80) NOT NULL,
  email VARCHAR(255) NOT NULL,
  password_hash VARCHAR(255) NULL,
  role ENUM('admin', 'user') NOT NULL DEFAULT 'user',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_users_email (email),
  INDEX idx_users_role (role)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS user_sessions (
  token CHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (token),
  INDEX idx_user_sessions_user_id (user_id),
  INDEX idx_user_sessions_expires_at (expires_at),
  CONSTRAINT fk_user_sessions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS yarns (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  photo_data MEDIUMTEXT NULL,
  name VARCHAR(120) NOT NULL,
  brand VARCHAR(120) NOT NULL DEFAULT '',
  color VARCHAR(80) NOT NULL DEFAULT '',
  material VARCHAR(120) NOT NULL DEFAULT '',
  thickness VARCHAR(80) NOT NULL DEFAULT '',
  plies TINYINT UNSIGNED NOT NULL DEFAULT 1,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  INDEX idx_yarns_user_id (user_id),
  CONSTRAINT fk_yarns_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS tools (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  photo_data MEDIUMTEXT NULL,
  name VARCHAR(120) NOT NULL,
  type VARCHAR(80) NOT NULL DEFAULT '',
  thickness VARCHAR(80) NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  INDEX idx_tools_user_id (user_id),
  CONSTRAINT fk_tools_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS projects (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  cover_data MEDIUMTEXT NULL,
  name VARCHAR(120) NOT NULL,
  pattern_name VARCHAR(160) NOT NULL DEFAULT '',
  target_size VARCHAR(80) NOT NULL DEFAULT '',
  hook_size VARCHAR(80) NOT NULL DEFAULT '',
  yarn_names JSON NOT NULL,
  tool_names JSON NOT NULL,
  status ENUM('planning', 'in_progress', 'completed') NOT NULL DEFAULT 'planning',
  current_row INT UNSIGNED NOT NULL DEFAULT 0,
  total_rows INT UNSIGNED NOT NULL DEFAULT 0,
  breakpoint_note TEXT NOT NULL,
  completed_at TIMESTAMP NULL,
  target_date DATE NULL,
  weekly_focus_minutes INT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  INDEX idx_projects_user_status (user_id, status),
  CONSTRAINT fk_projects_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS shopping_items (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  yarn_id BIGINT UNSIGNED NULL,
  yarn_name VARCHAR(120) NOT NULL,
  target_grams DECIMAL(10,2) NOT NULL,
  estimated_price DECIMAL(10,2) NOT NULL DEFAULT 0,
  priority ENUM('low','normal','high') NOT NULL DEFAULT 'normal',
  status ENUM('pending','purchased') NOT NULL DEFAULT 'pending',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  INDEX idx_shopping_items_user_status (user_id, status),
  CONSTRAINT fk_shopping_items_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_shopping_items_yarn FOREIGN KEY (yarn_id) REFERENCES yarns(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS project_yarn_plans (
  project_id BIGINT UNSIGNED NOT NULL,
  yarn_id BIGINT UNSIGNED NOT NULL,
  planned_grams DECIMAL(10,2) NOT NULL,
  PRIMARY KEY (project_id, yarn_id),
  CONSTRAINT fk_project_yarn_plans_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
  CONSTRAINT fk_project_yarn_plans_yarn FOREIGN KEY (yarn_id) REFERENCES yarns(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO patterns (title, category, difficulty, yarn, hook_size, duration_minutes) VALUES
  ('初夏小花束', '居家装饰', 'beginner', '棉线 2 股', '3.0 mm', 90),
  ('奶油吐司收纳篮', '居家装饰', 'intermediate', '再生棉绳', '5.0 mm', 180),
  ('海风渔夫帽', '穿搭配饰', 'intermediate', '亚麻棉混纺', '4.0 mm', 240)
ON DUPLICATE KEY UPDATE title = VALUES(title);

INSERT INTO users (name, email, role) VALUES
  ('线禾管理员', 'admin@line.local', 'admin')
ON DUPLICATE KEY UPDATE name = VALUES(name), role = VALUES(role);
