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

INSERT INTO patterns (title, category, difficulty, yarn, hook_size, duration_minutes) VALUES
  ('初夏小花束', '居家装饰', 'beginner', '棉线 2 股', '3.0 mm', 90),
  ('奶油吐司收纳篮', '居家装饰', 'intermediate', '再生棉绳', '5.0 mm', 180),
  ('海风渔夫帽', '穿搭配饰', 'intermediate', '亚麻棉混纺', '4.0 mm', 240)
ON DUPLICATE KEY UPDATE title = VALUES(title);
