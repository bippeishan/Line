CREATE TABLE stitch_terms (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  code VARCHAR(32) NOT NULL,
  name_zh VARCHAR(80) NOT NULL,
  name_en VARCHAR(120) NOT NULL DEFAULT '',
  category VARCHAR(40) NOT NULL,
  description TEXT NOT NULL,
  instruction TEXT NOT NULL,
  caution TEXT NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_stitch_terms_code (code),
  INDEX idx_stitch_terms_category (category)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE user_stitch_favorites (
  user_id BIGINT UNSIGNED NOT NULL,
  stitch_term_id BIGINT UNSIGNED NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (user_id, stitch_term_id),
  CONSTRAINT fk_stitch_favorite_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_stitch_favorite_term FOREIGN KEY (stitch_term_id) REFERENCES stitch_terms(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO stitch_terms (code,name_zh,name_en,category,description,instruction,caution) VALUES
  ('ch','锁针','chain stitch','基础','用于起针、转行或连接。','线绕钩针后从钩上的线圈中一次拉出。','保持松紧一致，起针过紧会影响后续宽度。'),
  ('sl st','引拔针','slip stitch','基础','用于连接、收尾或移动到指定位置。','将钩针插入针目，挂线后同时穿过针目和钩上原有线圈。','针目很低，连续使用会使织片收紧。'),
  ('sc','短针','single crochet','基础','紧密结实的基础针法，玩偶常用。','插入针目，挂线拉出一圈，再挂线一次拉过两圈。','注意每行起始针和结束针，避免漏针。'),
  ('inc','加针','increase','增减针','在同一个针目中完成两针，用于扩展圆形或增加宽度。','连续完成两次指定针法，并确认都落在同一针目。','每圈加针位置要均匀，才能保持圆形平整。'),
  ('dec','减针','decrease','增减针','将两针合并为一针，用于收口或塑形。','分别穿入相邻两针并保留线圈，最后一次拉过钩上全部线圈。','玩偶建议使用隐形减针，外观更平整。'),
  ('hdc','中长针','half double crochet','长针','高度介于短针和长针之间。','先绕线，插入针目拉出线圈，再一次拉过钩上三圈。','转行高度通常需要两条锁针。'),
  ('dc','长针','double crochet','长针','形成较高且有垂坠感的针目。','先绕线，插入针目拉出线圈，分两次各拉过两圈。','注意起立锁针高度和边缘针目位置。'),
  ('BLO','只钩后半针','back loop only','进阶','只挑取针目远离自己的半针，形成横向纹理。','确认 V 形针目的后侧半针后再完成指定针法。','若图解未说明，默认应挑两股针目。'),
  ('FLO','只钩前半针','front loop only','进阶','只挑取针目靠近自己的半针，常用于留下装饰边。','确认 V 形针目的前侧半针后再完成指定针法。','留下的另一半针可用于后续钩边。')
ON DUPLICATE KEY UPDATE name_zh=VALUES(name_zh),name_en=VALUES(name_en),category=VALUES(category),description=VALUES(description),instruction=VALUES(instruction),caution=VALUES(caution);
