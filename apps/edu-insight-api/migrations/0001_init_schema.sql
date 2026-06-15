-- 初始迁移占位
-- 后续按 docs/product/score-insight-platform-sql-draft-v1.sql 拆分迁移

CREATE TABLE schema_migrations_placeholder (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(128) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
