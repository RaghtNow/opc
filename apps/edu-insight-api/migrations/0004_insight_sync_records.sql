CREATE TABLE IF NOT EXISTS insight_sync_records (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  class_id VARCHAR(64) NOT NULL,
  target VARCHAR(128) NOT NULL,
  channel VARCHAR(64) NOT NULL,
  status VARCHAR(64) NOT NULL,
  published_at VARCHAR(32) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_insight_sync_records_class_created (class_id, created_at),
  CONSTRAINT fk_insight_sync_records_class FOREIGN KEY (class_id) REFERENCES classroom_classes(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
