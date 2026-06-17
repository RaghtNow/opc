UPDATE classroom_teachers
SET account_id = CONCAT('teacher-account-', id)
WHERE account_status = 'bound'
  AND account_id <> ''
  AND account_id <> CONCAT('teacher-account-', id);
