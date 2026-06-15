# 家校协同成绩洞察平台 数据库表设计初稿 V1.0

更新日期：2026-06-14  
用途：开发建模、接口设计、后续 ORM 设计输入

## 1. 设计原则

- 多租户，核心业务表必须带 `school_id`
- 优先支持高中选科走班
- 支持行政班、教学班、选科组合并存
- 支持基础建档与成绩导入分离
- 支持排名自动计算
- 支持规则配置化与快照缓存

## 2. 表清单总览

### 2.1 组织与身份
- `schools`
- `grades`
- `admin_classes`
- `teaching_classes`
- `teachers`
- `teacher_roles`
- `users`
- `students`
- `parent_contacts`
- `student_parent_relations`

### 2.2 学科与选科
- `subjects`
- `subject_combinations`
- `student_subject_selections`
- `teaching_class_students`
- `teacher_class_subject_relations`

### 2.3 考试与成绩
- `exam_types`
- `exams`
- `exam_subject_configs`
- `score_import_batches`
- `score_import_batch_items`
- `score_records`
- `score_aggregations`

### 2.4 分析与预警
- `analysis_snapshots`
- `alert_rule_templates`
- `teacher_alert_rules`
- `student_alerts`
- `student_tags`

### 2.5 展示策略与通知
- `display_policies`
- `notification_tasks`
- `notification_receipts`

## 3. 核心表设计

## 3.1 schools

用途：租户学校表

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| name | varchar(128) | 学校名称 |
| code | varchar(64) | 学校编码 |
| status | tinyint | 状态 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

## 3.2 grades

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| name | varchar(64) | 年级名称 |
| academic_year | varchar(32) | 学年 |
| stage | varchar(32) | 学段，如 high_school / junior_high |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

## 3.3 admin_classes

用途：行政班

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| grade_id | bigint | 年级 ID |
| name | varchar(64) | 班级名称 |
| teacher_id | bigint | 班主任 ID |
| source_type | varchar(32) | 创建来源，manual / school_import |
| status | tinyint | 状态 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

## 3.4 teaching_classes

用途：教学班

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| grade_id | bigint | 年级 ID |
| subject_id | bigint | 学科 ID |
| name | varchar(64) | 教学班名称 |
| teacher_id | bigint | 授课老师 ID |
| status | tinyint | 状态 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

## 3.5 users

用途：统一账号表

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| mobile | varchar(32) | 手机号 |
| password_hash | varchar(255) | 密码摘要 |
| wechat_openid | varchar(128) | 小程序 openid |
| status | tinyint | 状态 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

## 3.6 teachers

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| user_id | bigint | 账号 ID |
| name | varchar(64) | 姓名 |
| employee_no | varchar(64) | 工号 |
| status | tinyint | 状态 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

## 3.7 teacher_roles

用途：老师角色表

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| teacher_id | bigint | 老师 ID |
| role_type | varchar(32) | homeroom_teacher / subject_teacher / grade_leader / school_admin |
| scope_type | varchar(32) | school / grade / admin_class / teaching_class |
| scope_id | bigint | 作用范围 ID |
| created_at | datetime | 创建时间 |

## 3.8 students

用途：学生基础档案

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| grade_id | bigint | 年级 ID |
| admin_class_id | bigint | 行政班 ID |
| user_id | bigint null | 账号 ID，可为空 |
| student_no | varchar(64) | 学号 |
| name | varchar(64) | 姓名 |
| gender | varchar(16) | 性别 |
| mobile | varchar(32) | 手机号 |
| subject_combination_id | bigint null | 选科组合 ID |
| status | tinyint | 状态 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

## 3.9 parent_contacts

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| user_id | bigint null | 账号 ID |
| name | varchar(64) | 姓名 |
| mobile | varchar(32) | 手机号 |
| relation_type | varchar(32) | father / mother / guardian |
| status | tinyint | 状态 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

## 3.10 student_parent_relations

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| student_id | bigint | 学生 ID |
| parent_contact_id | bigint | 家长 ID |
| is_primary | tinyint | 是否主联系人 |
| created_at | datetime | 创建时间 |

## 3.11 subjects

用途：学科配置

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| code | varchar(32) | 学科编码 |
| name | varchar(64) | 学科名称 |
| is_core_subject | tinyint | 是否主科 |
| has_converted_score | tinyint | 是否支持赋分 |
| status | tinyint | 状态 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

## 3.12 subject_combinations

用途：选科组合

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| grade_id | bigint | 年级 ID |
| name | varchar(64) | 组合名称，如 物化生 |
| subject_codes | varchar(255) | 组合学科编码集合 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

## 3.13 student_subject_selections

用途：学生选科关系

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| student_id | bigint | 学生 ID |
| subject_id | bigint | 学科 ID |
| selection_type | varchar(32) | selected / required |
| created_at | datetime | 创建时间 |

## 3.14 teaching_class_students

用途：学生与教学班关系

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| teaching_class_id | bigint | 教学班 ID |
| student_id | bigint | 学生 ID |
| created_at | datetime | 创建时间 |

## 3.15 teacher_class_subject_relations

用途：老师与班级学科关系

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| teacher_id | bigint | 老师 ID |
| subject_id | bigint | 学科 ID |
| admin_class_id | bigint null | 行政班 ID |
| teaching_class_id | bigint null | 教学班 ID |
| created_at | datetime | 创建时间 |

## 3.16 exam_types

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| code | varchar(32) | weekly / monthly / midterm / final / mock / joint / custom |
| name | varchar(64) | 类型名称 |
| status | tinyint | 状态 |
| created_at | datetime | 创建时间 |

## 3.17 exams

用途：考试主表

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| grade_id | bigint | 年级 ID |
| admin_class_id | bigint | 行政班 ID |
| exam_type_id | bigint | 考试类型 ID |
| name | varchar(128) | 考试名称 |
| exam_date | date | 考试日期 |
| created_by | bigint | 创建人 teacher_id |
| status | tinyint | 状态 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

## 3.18 exam_subject_configs

用途：考试涉及学科配置

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| exam_id | bigint | 考试 ID |
| subject_id | bigint | 学科 ID |
| score_mode | varchar(32) | raw_only / raw_and_converted |
| include_in_total | tinyint | 是否计入总分 |
| created_at | datetime | 创建时间 |

## 3.19 score_import_batches

用途：成绩导入批次

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| exam_id | bigint | 考试 ID |
| uploaded_by | bigint | 上传人 teacher_id |
| file_name | varchar(255) | 文件名 |
| total_count | int | 总条数 |
| success_count | int | 成功条数 |
| failed_count | int | 失败条数 |
| status | varchar(32) | processing / success / partial_failed / failed |
| created_at | datetime | 创建时间 |

## 3.20 score_import_batch_items

用途：导入批次明细

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| batch_id | bigint | 批次 ID |
| row_no | int | 行号 |
| student_no | varchar(64) | 学号 |
| student_name | varchar(64) | 姓名 |
| subject_name | varchar(64) | 科目名称 |
| raw_payload | text | 原始内容 |
| parse_status | varchar(32) | success / failed |
| error_message | varchar(255) | 错误信息 |
| created_at | datetime | 创建时间 |

## 3.21 score_records

用途：单科成绩明细

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| exam_id | bigint | 考试 ID |
| batch_id | bigint | 导入批次 ID |
| student_id | bigint | 学生 ID |
| subject_id | bigint | 学科 ID |
| raw_score | decimal(8,2) null | 实际分 |
| converted_score | decimal(8,2) null | 赋分 |
| absent_flag | tinyint | 是否缺考 |
| remark | varchar(255) | 备注 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

## 3.22 score_aggregations

用途：考试聚合结果

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| exam_id | bigint | 考试 ID |
| student_id | bigint | 学生 ID |
| total_raw_score | decimal(10,2) null | 总实际分 |
| total_converted_score | decimal(10,2) null | 总赋分 |
| final_total_score | decimal(10,2) null | 最终统计总分 |
| rank_in_class | int null | 班级排名 |
| rank_in_grade | int null | 年级排名 |
| score_band | varchar(64) null | 分数段 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

## 3.23 analysis_snapshots

用途：分析结果快照

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| exam_id | bigint | 考试 ID |
| scope_type | varchar(32) | school / grade / admin_class / teaching_class / student |
| scope_id | bigint | 范围 ID |
| snapshot_type | varchar(32) | overview / trend / layering / warning |
| snapshot_data | json | 快照数据 |
| generated_at | datetime | 生成时间 |

## 3.24 alert_rule_templates

用途：学校级预警模板

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| name | varchar(128) | 规则名称 |
| rule_type | varchar(32) | decline / fluctuation / weak_subject / absent |
| rule_config | json | 规则配置 |
| enabled | tinyint | 是否启用 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

## 3.25 teacher_alert_rules

用途：班主任个性规则

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| teacher_id | bigint | 班主任 ID |
| admin_class_id | bigint | 行政班 ID |
| name | varchar(128) | 规则名称 |
| rule_type | varchar(32) | 规则类型 |
| rule_config | json | 规则配置 |
| enabled | tinyint | 是否启用 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

## 3.26 student_alerts

用途：学生预警结果

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| exam_id | bigint | 考试 ID |
| student_id | bigint | 学生 ID |
| rule_source | varchar(32) | school_template / teacher_rule |
| rule_id | bigint | 规则 ID |
| alert_type | varchar(32) | 预警类型 |
| alert_level | varchar(32) | low / medium / high |
| alert_message | varchar(255) | 提示语 |
| status | varchar(32) | active / resolved |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

## 3.27 student_tags

用途：学生标签

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| student_id | bigint | 学生 ID |
| tag_type | varchar(32) | top_student / borderline / unstable / support_needed |
| source | varchar(32) | system / teacher |
| created_at | datetime | 创建时间 |

## 3.28 display_policies

用途：展示策略

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| scope_type | varchar(32) | school / admin_class |
| scope_id | bigint | 范围 ID |
| show_class_rank | tinyint | 是否展示班级排名 |
| show_grade_rank | tinyint | 是否展示年级排名 |
| show_score_band | tinyint | 是否展示分数段 |
| show_class_avg_compare | tinyint | 是否展示均分对比 |
| created_by | bigint | 配置人 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

## 3.29 notification_tasks

用途：通知任务主表

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| exam_id | bigint null | 考试 ID |
| sender_teacher_id | bigint | 发起人 |
| channel | varchar(32) | mini_program / sms |
| receiver_type | varchar(32) | parent / student / teacher |
| content_type | varchar(32) | score_sync / alert_sync / reminder |
| payload | json | 发送内容 |
| status | varchar(32) | pending / success / partial_failed / failed |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

## 3.30 notification_receipts

用途：通知接收结果

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint pk | 主键 |
| school_id | bigint | 学校 ID |
| task_id | bigint | 通知任务 ID |
| receiver_user_id | bigint null | 接收账号 ID |
| receiver_mobile | varchar(32) null | 接收手机号 |
| receiver_type | varchar(32) | parent / student / teacher |
| send_status | varchar(32) | success / failed |
| read_status | varchar(32) | unread / read |
| fail_reason | varchar(255) null | 失败原因 |
| sent_at | datetime null | 发送时间 |
| read_at | datetime null | 阅读时间 |
| created_at | datetime | 创建时间 |

## 4. 索引建议

建议重点索引：
- `students(school_id, grade_id, admin_class_id, student_no)`
- `score_records(school_id, exam_id, student_id, subject_id)`
- `score_aggregations(school_id, exam_id, student_id)`
- `student_alerts(school_id, exam_id, student_id, status)`
- `notification_receipts(task_id, receiver_user_id, send_status)`

## 5. 首期可延后表

若要压缩首期范围，可先简化或后置：
- `teaching_classes`
- `teaching_class_students`
- `teacher_class_subject_relations`
- `analysis_snapshots` 中的复杂快照类型
- `student_tags` 的人工标签体系

## 6. 首期实现建议

首期建议优先实现这些表：
- schools
- grades
- admin_classes
- users
- teachers
- students
- parent_contacts
- student_parent_relations
- subjects
- subject_combinations
- student_subject_selections
- exam_types
- exams
- score_import_batches
- score_import_batch_items
- score_records
- score_aggregations
- alert_rule_templates
- teacher_alert_rules
- student_alerts
- display_policies
- notification_tasks
- notification_receipts
