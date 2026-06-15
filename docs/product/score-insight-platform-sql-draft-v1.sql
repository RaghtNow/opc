-- 家校协同成绩洞察平台 SQL 初稿 V1
-- 更新时间：2026-06-15
-- 说明：面向 MySQL 8+，用于后端骨架启动，不代表最终定版

CREATE TABLE schools (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(128) NOT NULL,
  code VARCHAR(64) NOT NULL UNIQUE,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE grades (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  school_id BIGINT NOT NULL,
  name VARCHAR(64) NOT NULL,
  academic_year VARCHAR(32) NOT NULL,
  stage VARCHAR(32) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_school_grade (school_id, name)
);

CREATE TABLE users (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  mobile VARCHAR(32) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NULL,
  wechat_openid VARCHAR(128) NULL UNIQUE,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE teachers (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  school_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  name VARCHAR(64) NOT NULL,
  employee_no VARCHAR(64) NULL,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_school_teacher (school_id, name)
);

CREATE TABLE admin_classes (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  school_id BIGINT NOT NULL,
  grade_id BIGINT NOT NULL,
  name VARCHAR(64) NOT NULL,
  teacher_id BIGINT NULL,
  selection_stage VARCHAR(32) NOT NULL DEFAULT 'pre-selection',
  source_type VARCHAR(32) NOT NULL DEFAULT 'manual',
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_school_grade_class (school_id, grade_id, name)
);

CREATE TABLE subjects (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  school_id BIGINT NOT NULL,
  code VARCHAR(32) NOT NULL,
  name VARCHAR(64) NOT NULL,
  is_core_subject TINYINT NOT NULL DEFAULT 0,
  has_converted_score TINYINT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_school_subject_code (school_id, code)
);

CREATE TABLE subject_combinations (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  school_id BIGINT NOT NULL,
  grade_id BIGINT NOT NULL,
  name VARCHAR(64) NOT NULL,
  subject_codes VARCHAR(255) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE students (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  school_id BIGINT NOT NULL,
  grade_id BIGINT NOT NULL,
  admin_class_id BIGINT NOT NULL,
  user_id BIGINT NULL,
  student_no VARCHAR(64) NOT NULL,
  name VARCHAR(64) NOT NULL,
  gender VARCHAR(16) NOT NULL,
  mobile VARCHAR(32) NULL,
  subject_combination_id BIGINT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_school_student_no (school_id, student_no),
  KEY idx_class_student (admin_class_id, name)
);

CREATE TABLE parent_contacts (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  school_id BIGINT NOT NULL,
  user_id BIGINT NULL,
  name VARCHAR(64) NOT NULL,
  mobile VARCHAR(32) NOT NULL,
  relation_type VARCHAR(32) NOT NULL,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_school_parent_mobile (school_id, mobile)
);

CREATE TABLE student_parent_relations (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  school_id BIGINT NOT NULL,
  student_id BIGINT NOT NULL,
  parent_contact_id BIGINT NOT NULL,
  is_primary TINYINT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_student_parent (student_id, parent_contact_id)
);

CREATE TABLE student_subject_selections (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  school_id BIGINT NOT NULL,
  student_id BIGINT NOT NULL,
  subject_id BIGINT NOT NULL,
  selection_type VARCHAR(32) NOT NULL DEFAULT 'selected',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_student_subject (student_id, subject_id)
);

CREATE TABLE teaching_classes (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  school_id BIGINT NOT NULL,
  grade_id BIGINT NOT NULL,
  subject_id BIGINT NOT NULL,
  name VARCHAR(64) NOT NULL,
  teacher_id BIGINT NULL,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE teaching_class_students (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  school_id BIGINT NOT NULL,
  teaching_class_id BIGINT NOT NULL,
  student_id BIGINT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_teaching_class_student (teaching_class_id, student_id)
);

CREATE TABLE teacher_roles (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  school_id BIGINT NOT NULL,
  teacher_id BIGINT NOT NULL,
  role_type VARCHAR(32) NOT NULL,
  scope_type VARCHAR(32) NOT NULL,
  scope_id BIGINT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_teacher_scope (teacher_id, scope_type, scope_id)
);

CREATE TABLE teacher_class_subject_relations (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  school_id BIGINT NOT NULL,
  teacher_id BIGINT NOT NULL,
  subject_id BIGINT NOT NULL,
  admin_class_id BIGINT NULL,
  teaching_class_id BIGINT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE exam_types (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  school_id BIGINT NOT NULL,
  code VARCHAR(32) NOT NULL,
  name VARCHAR(64) NOT NULL,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_school_exam_type (school_id, code)
);

CREATE TABLE exams (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  school_id BIGINT NOT NULL,
  grade_id BIGINT NOT NULL,
  admin_class_id BIGINT NOT NULL,
  exam_type_id BIGINT NOT NULL,
  name VARCHAR(128) NOT NULL,
  exam_date DATE NOT NULL,
  selection_stage VARCHAR(32) NOT NULL,
  created_by BIGINT NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'draft',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_exam_class_date (admin_class_id, exam_date)
);

CREATE TABLE exam_subject_configs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  school_id BIGINT NOT NULL,
  exam_id BIGINT NOT NULL,
  subject_id BIGINT NOT NULL,
  score_mode VARCHAR(32) NOT NULL DEFAULT 'raw_only',
  include_in_total TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_exam_subject (exam_id, subject_id)
);

CREATE TABLE score_import_batches (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  school_id BIGINT NOT NULL,
  exam_id BIGINT NOT NULL,
  uploaded_by BIGINT NOT NULL,
  file_name VARCHAR(255) NOT NULL,
  file_type VARCHAR(16) NOT NULL,
  total_count INT NOT NULL DEFAULT 0,
  success_count INT NOT NULL DEFAULT 0,
  failed_count INT NOT NULL DEFAULT 0,
  status VARCHAR(32) NOT NULL DEFAULT 'processing',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE score_import_batch_items (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  school_id BIGINT NOT NULL,
  batch_id BIGINT NOT NULL,
  row_no INT NOT NULL,
  student_no VARCHAR(64) NULL,
  student_name VARCHAR(64) NULL,
  raw_payload JSON NOT NULL,
  parse_status VARCHAR(32) NOT NULL DEFAULT 'pending',
  error_message VARCHAR(255) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_batch_row (batch_id, row_no)
);

CREATE TABLE score_records (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  school_id BIGINT NOT NULL,
  exam_id BIGINT NOT NULL,
  batch_id BIGINT NULL,
  student_id BIGINT NOT NULL,
  subject_id BIGINT NOT NULL,
  raw_score DECIMAL(8,2) NULL,
  converted_score DECIMAL(8,2) NULL,
  absent_flag TINYINT NOT NULL DEFAULT 0,
  remark VARCHAR(255) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_exam_student_subject (exam_id, student_id, subject_id)
);

CREATE TABLE score_aggregations (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  school_id BIGINT NOT NULL,
  exam_id BIGINT NOT NULL,
  student_id BIGINT NOT NULL,
  total_raw_score DECIMAL(10,2) NULL,
  total_converted_score DECIMAL(10,2) NULL,
  final_total_score DECIMAL(10,2) NULL,
  rank_in_class INT NULL,
  rank_in_grade INT NULL,
  score_band VARCHAR(64) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_exam_student_agg (exam_id, student_id)
);

CREATE TABLE analysis_snapshots (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  school_id BIGINT NOT NULL,
  exam_id BIGINT NOT NULL,
  scope_type VARCHAR(32) NOT NULL,
  scope_id BIGINT NOT NULL,
  snapshot_type VARCHAR(32) NOT NULL,
  snapshot_data JSON NOT NULL,
  generated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_exam_scope_snapshot (exam_id, scope_type, scope_id, snapshot_type)
);

CREATE TABLE alert_rule_templates (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  school_id BIGINT NOT NULL,
  name VARCHAR(128) NOT NULL,
  rule_type VARCHAR(32) NOT NULL,
  rule_config JSON NOT NULL,
  enabled TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE teacher_alert_rules (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  school_id BIGINT NOT NULL,
  teacher_id BIGINT NOT NULL,
  admin_class_id BIGINT NOT NULL,
  name VARCHAR(128) NOT NULL,
  rule_type VARCHAR(32) NOT NULL,
  rule_config JSON NOT NULL,
  enabled TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE student_alerts (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  school_id BIGINT NOT NULL,
  exam_id BIGINT NOT NULL,
  student_id BIGINT NOT NULL,
  rule_source VARCHAR(32) NOT NULL,
  rule_id BIGINT NOT NULL,
  alert_type VARCHAR(32) NOT NULL,
  alert_level VARCHAR(32) NOT NULL,
  alert_message VARCHAR(255) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_exam_student_alert (exam_id, student_id, status)
);

CREATE TABLE display_policies (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  school_id BIGINT NOT NULL,
  scope_type VARCHAR(32) NOT NULL,
  scope_id BIGINT NOT NULL,
  show_class_rank TINYINT NOT NULL DEFAULT 0,
  show_grade_rank TINYINT NOT NULL DEFAULT 0,
  show_score_band TINYINT NOT NULL DEFAULT 0,
  show_class_avg_compare TINYINT NOT NULL DEFAULT 0,
  created_by BIGINT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE notification_tasks (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  school_id BIGINT NOT NULL,
  exam_id BIGINT NULL,
  sender_teacher_id BIGINT NOT NULL,
  channel VARCHAR(32) NOT NULL,
  receiver_type VARCHAR(32) NOT NULL,
  content_type VARCHAR(32) NOT NULL,
  payload JSON NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE notification_receipts (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  school_id BIGINT NOT NULL,
  task_id BIGINT NOT NULL,
  receiver_user_id BIGINT NULL,
  receiver_mobile VARCHAR(32) NULL,
  receiver_type VARCHAR(32) NOT NULL,
  send_status VARCHAR(32) NOT NULL,
  read_status VARCHAR(32) NOT NULL DEFAULT 'unread',
  fail_reason VARCHAR(255) NULL,
  sent_at DATETIME NULL,
  read_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_task_receiver (task_id, receiver_type, send_status)
);
