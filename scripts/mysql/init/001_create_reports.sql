CREATE DATABASE IF NOT EXISTS infohub
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE infohub;

CREATE TABLE IF NOT EXISTS reports (
  id BIGINT NOT NULL AUTO_INCREMENT,
  generated_at DATETIME(6) NOT NULL,
  markdown LONGTEXT NOT NULL,
  items_json LONGTEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_generated_at (generated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS user_preferences (
  user_id VARCHAR(191) NOT NULL,
  tags_json LONGTEXT NOT NULL,
  sources_json LONGTEXT NOT NULL,
  keywords_json LONGTEXT NOT NULL,
  tag_weight DOUBLE NOT NULL,
  source_weight DOUBLE NOT NULL,
  keyword_weight DOUBLE NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (user_id),
  KEY idx_updated_at (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
