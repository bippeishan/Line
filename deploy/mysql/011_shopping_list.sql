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
