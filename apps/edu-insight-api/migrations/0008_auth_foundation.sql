CREATE TABLE IF NOT EXISTS auth_users (
  id VARCHAR(64) PRIMARY KEY,
  name VARCHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS auth_user_identities (
  id VARCHAR(64) PRIMARY KEY,
  user_id VARCHAR(64) NOT NULL,
  identity_type VARCHAR(32) NOT NULL,
  identifier VARCHAR(128) NOT NULL,
  verified TINYINT NOT NULL DEFAULT 0,
  last_verified_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_auth_identity (identity_type, identifier),
  KEY idx_auth_identity_user (user_id),
  CONSTRAINT fk_auth_identity_user FOREIGN KEY (user_id) REFERENCES auth_users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS auth_sms_codes (
  id VARCHAR(64) PRIMARY KEY,
  mobile VARCHAR(32) NOT NULL,
  scene VARCHAR(32) NOT NULL,
  code_hash VARCHAR(128) NOT NULL,
  expires_at DATETIME NOT NULL,
  used_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_auth_sms_lookup (mobile, scene, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS auth_sessions (
  id VARCHAR(64) PRIMARY KEY,
  user_id VARCHAR(64) NOT NULL,
  token_hash VARCHAR(128) NOT NULL,
  expires_at DATETIME NOT NULL,
  revoked_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_auth_session_token (token_hash),
  KEY idx_auth_session_user (user_id),
  CONSTRAINT fk_auth_session_user FOREIGN KEY (user_id) REFERENCES auth_users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS auth_work_identities (
  id VARCHAR(64) PRIMARY KEY,
  user_id VARCHAR(64) NOT NULL,
  role_type VARCHAR(64) NOT NULL,
  role_label VARCHAR(64) NOT NULL,
  primary_label VARCHAR(128) NOT NULL,
  secondary_label VARCHAR(255) NOT NULL,
  scope_type VARCHAR(64) NOT NULL,
  scope_id VARCHAR(64) NOT NULL,
  subject VARCHAR(64) NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  sort_order INT NOT NULL DEFAULT 100,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_auth_work_identity_user (user_id, status),
  CONSTRAINT fk_auth_work_identity_user FOREIGN KEY (user_id) REFERENCES auth_users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO auth_users (id, name, status)
VALUES
  ('user-teacher-li', '李老师', 'active'),
  ('user-teacher-wang', '王老师', 'active'),
  ('user-teacher-zhang', '张老师', 'active')
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  status = VALUES(status);

INSERT INTO auth_user_identities (id, user_id, identity_type, identifier, verified, last_verified_at)
VALUES
  ('identity-mobile-13800001003', 'user-teacher-li', 'mobile', '13800001003', 1, NOW()),
  ('identity-mobile-13800001002', 'user-teacher-wang', 'mobile', '13800001002', 1, NOW()),
  ('identity-mobile-13800001001', 'user-teacher-zhang', 'mobile', '13800001001', 1, NOW())
ON DUPLICATE KEY UPDATE
  user_id = VALUES(user_id),
  verified = VALUES(verified),
  last_verified_at = VALUES(last_verified_at);

INSERT INTO auth_work_identities (
  id, user_id, role_type, role_label, primary_label, secondary_label, scope_type, scope_id, subject, status, sort_order
) VALUES
  ('homeroom', 'user-teacher-li', 'homeroom_teacher', '班主任', '班级管理', '高二（3）班 / 高一（8）班', 'admin_class_group', 'homeroom-li', '', 'active', 10),
  ('subject-english', 'user-teacher-li', 'subject_teacher', '任课老师', '英语任课分析', '高二（3）班 / 高一（8）班', 'teaching_scope', 'teacher-english-li', '英语', 'active', 20),
  ('subject-math', 'user-teacher-wang', 'subject_teacher', '任课老师', '数学任课分析', '高二（3）班 / 高一（8）班', 'teaching_scope', 'teacher-math-wang', '数学', 'active', 10),
  ('subject-chinese', 'user-teacher-zhang', 'subject_teacher', '任课老师', '语文任课分析', '高二（3）班', 'teaching_scope', 'teacher-chinese-zhang', '语文', 'active', 10)
ON DUPLICATE KEY UPDATE
  role_type = VALUES(role_type),
  role_label = VALUES(role_label),
  primary_label = VALUES(primary_label),
  secondary_label = VALUES(secondary_label),
  scope_type = VALUES(scope_type),
  scope_id = VALUES(scope_id),
  subject = VALUES(subject),
  status = VALUES(status),
  sort_order = VALUES(sort_order);
