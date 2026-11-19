-- 配色模板库：官方内置方案 + 用户自建收藏。
-- owner_key 由 user_id 派生：内置模板落 0，自建落具体用户，
-- 唯一键立在 (owner_key, name) 上，于是内置名唯一、用户各自主建名唯一。
-- owner_key 必须是 VIRTUAL：STORED 生成列与其基列的 ON DELETE CASCADE 外键互斥（MySQL 1215）。
-- INSERT ... ON DUPLICATE KEY 让种子可重复执行。

CREATE TABLE IF NOT EXISTS palette_templates (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NULL,
  owner_key BIGINT UNSIGNED AS (COALESCE(user_id, 0)) VIRTUAL,
  name VARCHAR(60) NOT NULL,
  style VARCHAR(40) NOT NULL DEFAULT '',
  occasion VARCHAR(40) NOT NULL DEFAULT '',
  description VARCHAR(200) NOT NULL DEFAULT '',
  colors JSON NOT NULL,
  is_builtin BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_palette_templates_owner_name (owner_key, name),
  INDEX idx_palette_templates_style (is_builtin, style),
  CONSTRAINT fk_palette_templates_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO palette_templates (user_id, name, style, occasion, description, colors, is_builtin) VALUES
  (NULL, '奶油草莓熊', '甜系', '玩偶', '浅奶油打底，草莓粉做鼻子与耳朵，深莓色收边。', '["#f6e7dd", "#e8a0a8", "#a8576b"]', TRUE),
  (NULL, '春日小花束', '清新', '花束', '嫩芽绿配鹅黄与奶白，适合花束、花环与挂饰。', '["#d9e6cf", "#f4e3a1", "#fbf7ef"]', TRUE),
  (NULL, '秋日栗子', '大地色', '玩偶', '焦糖与栗棕，勾兔、熊一类最稳的一组配色。', '["#c98a5e", "#8c5a3c", "#f0ddc7"]', TRUE),
  (NULL, '莫兰迪灰蓝', '低饱和', '服饰', '灰蓝加奶灰与雾白，弱对比、耐看，适合围巾与帽饰。', '["#9aa9b3", "#d5dbdd", "#eceae4"]', TRUE),
  (NULL, '森林苔藓', '自然', '居家', '苔藓绿、松针绿与燕麦色，适合收纳篮与地垫。', '["#6f7f5c", "#41503c", "#e3dcc7"]', TRUE),
  (NULL, '夏日海盐', '清爽', '穿搭', '海盐蓝与沙白、珊瑚橙点缀，适合渔夫帽、夏凉包。', '["#7fa8c9", "#eef2f4", "#e8967a"]', TRUE),
  (NULL, '莓果拿铁', '暖调', '玩偶', '拿铁棕打底，莓红点睛，秋冬玩偶的常用组合。', '["#c8ab8d", "#9a3b52", "#f2e8dc"]', TRUE),
  (NULL, '马卡龙三色', '糖果', '挂饰', '薄荷、樱粉与柠檬黄，小件挂饰与钥匙扣很讨喜。', '["#a8dfd0", "#f7b7c4", "#f8e08e"]', TRUE),
  (NULL, '水墨灰白', '极简', '家居', '冷灰与米白加一点墨色，适合简洁线条的杯垫与餐垫。', '["#b8bcbd", "#f0efeb", "#3f4650"]', TRUE),
  (NULL, '蜜桃橘光', '明亮', '玩偶', '蜜桃橘搭奶白与暖棕，成品气色特别好。', '["#f2a377", "#f7e6d5", "#b3765a"]', TRUE)
ON DUPLICATE KEY UPDATE
  style = VALUES(style),
  occasion = VALUES(occasion),
  description = VALUES(description),
  colors = VALUES(colors);

-- 校验：内置模板数量与总行数。
SELECT SUM(is_builtin) AS builtin_templates, COUNT(*) AS total_templates FROM palette_templates;
