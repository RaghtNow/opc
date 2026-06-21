ALTER TABLE classroom_teachers
  ADD COLUMN mobile VARCHAR(32) NOT NULL DEFAULT '' AFTER teacher,
  ADD COLUMN account_id VARCHAR(64) NOT NULL DEFAULT '' AFTER account_status,
  ADD COLUMN account_bound_at VARCHAR(32) NOT NULL DEFAULT '' AFTER account_id,
  ADD COLUMN permission_synced_at VARCHAR(32) NOT NULL DEFAULT '' AFTER permission_status;
